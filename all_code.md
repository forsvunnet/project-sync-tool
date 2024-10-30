## init.go
```go
// cmd/pst/init.go
package pst

import (
	"fmt"
	"github.com/forsvunnet/project-sync-tool/internal/collections"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
    Use:   "init <collection-name> [path/to/file/or/folder...]",
    Short: "Add files or folders to a named collection",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        collectionName := args[0]
		// Validate collection name for alphanumeric characters only
        if err := collections.IsValidCollectionName(collectionName); err != nil {
            return err
        }
        paths := args[1:]

        // If no paths are specified, default to the current directory
        if len(paths) == 0 {
            paths = append(paths, ".")
        }

        // Call the internal collections package to handle sharing
        err := collections.AddToCollection(collectionName, paths, force)
        if err != nil {
            return fmt.Errorf("failed to add files to collection: %w", err)
        }

        fmt.Printf("Files added to collection %s successfully.\n", collectionName)
        return nil
    },
}


```

## pst.go
```go
package pst

import (
    "github.com/spf13/cobra"
)

var force bool

// Execute initializes the root command and adds subcommands
func Execute() error {
    rootCmd := &cobra.Command{Use: "pst"}
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(requireCmd)
    rootCmd.AddCommand(pushCmd)
    return rootCmd.Execute()
}

func init() {
    initCmd.Flags().BoolVarP(&force, "force", "f", false, "Forcefully replace existing files in the collection")
    pushCmd.Flags().BoolVarP(&force, "force", "f", false, "Forcefully overwrite central files even if they are newer")
}
```

## push.go
```go
// cmd/pst/push.go

package pst

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/forsvunnet/project-sync-tool/internal/collections"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
    Use:   "push <collection-name...>",
    Short: "Push changes from the current project to the central collection",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get the current working directory
        cwd, err := os.Getwd()
        if err != nil {
            return fmt.Errorf("failed to get current working directory: %w", err)
        }

        // Determine collections to process
        var collectionsToPush []string
        if len(args) > 0 {
            collectionsToPush = args
        } else {
            collectionsToPush, err = collections.ScanForCollections(cwd)
            if err != nil {
                return fmt.Errorf("failed to scan for collections: %w", err)
            }
        }

        for _, collectionName := range collectionsToPush {
            // Step 1: Check for changes
            changeStatus, err := collections.CheckForChanges(collectionName, cwd)
            if err != nil {
                return fmt.Errorf("failed to check changes for collection %s: %w", collectionName, err)
            }

            // Step 2: Handle conflicts
            if len(changeStatus.CentralNewer) > 0 && !force {
                return fmt.Errorf("conflict detected: central files are newer than local files in collection %s. Use --force to overwrite", collectionName)
            }

            // Step 3: Push updates for files that are newer in the local project
            for _, file := range changeStatus.LocalNewer {
                relPath, _ := filepath.Rel(cwd, file)
                centralPath := filepath.Join(collections.GetCollectionPath(collectionName), relPath)
                if err := collections.CopyFile(file, centralPath); err != nil {
                    return fmt.Errorf("failed to copy %s to central collection: %w", file, err)
                }
                fmt.Printf("Updated %s in central collection for %s\n", relPath, collectionName)
            }
        }

        fmt.Println("Push operation completed successfully.")
        return nil
    },
}

```

## require.go
```go
// cmd/pst/require.go

package pst

import (
	"fmt"
	"github.com/forsvunnet/project-sync-tool/internal/collections"
	"github.com/spf13/cobra"
)

var targetDir string

var requireCmd = &cobra.Command{
    Use:   "require <collection-name> [target-path]",
    Short: "Require files from a named collection into the target project",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        collectionName := args[0]

        // Validate the collection name
        if err := collections.IsValidCollectionName(collectionName); err != nil {
            return err
        }

        // Determine the target directory
        targetPath := "."
        if len(args) > 1 {
            targetPath = args[1]
        }

        if err := collections.RequireCollection(collectionName, targetPath); err != nil {
            return fmt.Errorf("failed to require collection: %w", err)
        }

        fmt.Printf("Collection %s requireed successfully to %s.\n", collectionName, targetPath)
        return nil
    },
}

func init() {
    requireCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Specify a target directory to load the collection into")
}


```

## changes.go
```go
// internal/collections/changes.go

package collections

import (
    "os"
    "path/filepath"
    "fmt"
)

// ChangeStatus represents the comparison result between local and central files.
type ChangeStatus struct {
    LocalNewer   []string // Files in the project that are newer than those in central
    CentralNewer []string // Files in central that are newer than those in the project
}

// CheckForChanges compares files in the local directory and central collection.
// First checks file checksums; if they match, no further checks are needed.
func CheckForChanges(collectionName, projectPath string) (ChangeStatus, error) {
    collectionFiles, err := GetCollectionFiles(collectionName)
    if err != nil {
        return ChangeStatus{}, err
    }

    status := ChangeStatus{}
    for _, centralFilePath := range collectionFiles {
        relPath, err := filepath.Rel(GetCollectionPath(collectionName), centralFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to calculate relative path: %w", err)
        }

        projectFilePath := filepath.Join(projectPath, relPath)

        // If the project file doesn’t exist, mark it as a new local file
        projectInfo, err := os.Stat(projectFilePath)
        if os.IsNotExist(err) {
            status.LocalNewer = append(status.LocalNewer, projectFilePath)
            continue
        } else if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to stat project file %s: %w", projectFilePath, err)
        }

        // Calculate checksums for both files
        centralChecksum, err := calculateChecksum(centralFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to calculate checksum for central file %s: %w", centralFilePath, err)
        }

        projectChecksum, err := calculateChecksum(projectFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to calculate checksum for project file %s: %w", projectFilePath, err)
        }

        // If checksums match, skip further checks for this file
        if centralChecksum == projectChecksum {
            continue
        }

        // If checksums differ, compare timestamps to determine which is newer
        centralInfo, err := os.Stat(centralFilePath)
        if err != nil {
            return ChangeStatus{}, fmt.Errorf("failed to stat central file %s: %w", centralFilePath, err)
        }

        if projectInfo.ModTime().After(centralInfo.ModTime()) {
            status.LocalNewer = append(status.LocalNewer, projectFilePath)
        } else if centralInfo.ModTime().After(projectInfo.ModTime()) {
            status.CentralNewer = append(status.CentralNewer, projectFilePath)
        }
    }

    return status, nil
}

```

## checksum.go
```go
// internal/collections/checksum.go

package collections

import (
    "crypto/sha256"
    "fmt"
    "io"
    "os"
)

// calculateChecksum calculates the SHA-256 checksum of a file at the given path.
func calculateChecksum(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to open file for checksum calculation: %w", err)
    }
    defer file.Close()

    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return "", fmt.Errorf("failed to calculate checksum for file: %w", err)
    }

    return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

```

## collections.go
```go
// internal/collections/collections.go
package collections

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "io"
    "os"
    "path/filepath"
)

type CollectionMeta struct {
    Paths []string `yaml:"paths"`
}

// GetCollectionPath returns the path for the central storage of collections
func GetCollectionPath(collectionName string) string {
    return filepath.Join(os.Getenv("HOME"), ".config", "project-sync-tool", "collections", collectionName)
}

// AddToCollection adds files or folders to a collection, overwriting existing files if `force` is true.
func AddToCollection(collectionName string, paths []string, force bool) error {
    collectionPath := GetCollectionPath(collectionName)


    // Check if the collection already exists
    if _, err := os.Stat(collectionPath); err == nil && !force {
        return fmt.Errorf("collection %s already exists; use --force to overwrite", collectionName)
    }

	if force {
        if err := os.RemoveAll(collectionPath); err != nil {
            return fmt.Errorf("failed to clear collection directory: %w", err)
        }
        if err := os.MkdirAll(collectionPath, os.ModePerm); err != nil {
            return fmt.Errorf("failed to recreate collection directory: %w", err)
        }
    } else {
        // Create the collection directory if it doesn't exist
        if err := os.MkdirAll(collectionPath, os.ModePerm); err != nil {
            return fmt.Errorf("failed to create collection directory: %w", err)
        }
    }

    // Copy each specified path into the collection
    for _, path := range paths {
        if err := copyToCollection(path, collectionPath); err != nil {
            return fmt.Errorf("failed to add %s to collection: %w", path, err)
        }
    }

    // Save collection metadata to a YAML file
    if err := saveCollectionMeta(collectionName); err != nil {
        return fmt.Errorf("failed to save collection metadata: %w", err)
    }
    return nil
}

// getMetaFilePath returns the path for storing collection metadata in the meta directory.
func getMetaFilePath(collectionName string) string {
    return filepath.Join(os.Getenv("HOME"), ".config", "project-sync-tool", "meta", fmt.Sprintf("%s.yml", collectionName))
}

// requireCollectionMeta loads the metadata file if it exists or initializes a new CollectionMeta.
func requireCollectionMeta(collectionName string) (CollectionMeta, error) {
    metaFile := getMetaFilePath(collectionName)
    meta := CollectionMeta{}

    data, err := os.ReadFile(metaFile)
    if os.IsNotExist(err) {
        // Return an empty CollectionMeta if the file doesn't exist
        return meta, nil
    } else if err != nil {
        return meta, fmt.Errorf("failed to read metadata file: %w", err)
    }

    // Unmarshal YAML data if the file exists
    if err := yaml.Unmarshal(data, &meta); err != nil {
        return meta, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }

    return meta, nil
}

// saveCollectionMeta saves collection paths as relative to the command execution directory, if possible.
func saveCollectionMeta(collectionName string) error {
    metaFile := getMetaFilePath(collectionName)

    // Create the meta directory if it doesn't exist
    metaDir := filepath.Dir(metaFile)
    if err := os.MkdirAll(metaDir, os.ModePerm); err != nil {
        return fmt.Errorf("failed to create meta directory: %w", err)
    }

    // Require existing metadata if available
    meta, err := requireCollectionMeta(collectionName)
    if err != nil {
        return fmt.Errorf("failed to require existing metadata: %w", err)
    }

    // Get the current working directory where the command is executed
    commandDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("could not determine working directory: %w", err)
    }

    // Add commandDir if not already in relativePaths
    alreadyExists := false
    for _, existingPath := range meta.Paths {
        if existingPath == commandDir {
            alreadyExists = true
            break
        }
    }
    if !alreadyExists {
        meta.Paths = append(meta.Paths, commandDir)
    }

    // Store the updated paths in the YAML file
    data, err := yaml.Marshal(&meta)
    if err != nil {
        return fmt.Errorf("failed to marshal collection metadata: %w", err)
    }

    return os.WriteFile(metaFile, data, 0644)
}

// copyToCollection copies files or directories to the collection path
func copyToCollection(srcPath, destDir string) error {
    srcInfo, err := os.Stat(srcPath)
    if err != nil {
        return fmt.Errorf("could not access path %s: %w", srcPath, err)
    }

    destPath := filepath.Join(destDir, filepath.Base(srcPath))

    if srcInfo.IsDir() {
        return copyDirectory(srcPath, destPath)
    }
    return CopyFile(srcPath, destPath)
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()

    if _, err = io.Copy(out, in); err != nil {
        return err
    }
    return out.Close()
}

// copyDirectory recursively copies a directory from src to dst
func copyDirectory(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        relPath, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }
        targetPath := filepath.Join(dst, relPath)

        if info.IsDir() {
            return os.MkdirAll(targetPath, info.Mode())
        }

        return CopyFile(path, targetPath)
    })
}



// RequireCollection requires all files from the collection directory into the current working directory
// and adds the current working directory to the metadata if it's not already there.
func RequireCollection(collectionName string, cwd string) error {
    collectionPath := GetCollectionPath(collectionName)

    // Check if the collection exists
    if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
        return fmt.Errorf("collection %s does not exist", collectionName)
    }

	if cwd == "." {
		// Get the current working directory (cwd)
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine current working directory: %w", err)
		}
	}

    // Copy each file in the collection to the cwd
    if err := copyCollectionFilesToTarget(collectionPath, cwd); err != nil {
        return fmt.Errorf("failed to copy files to current directory: %w", err)
    }

	saveCollectionMeta(collectionName)
    return nil
}

// copyCollectionFilesToTarget copies all files in collectionPath to targetPath (cwd).
func copyCollectionFilesToTarget(collectionPath, targetPath string) error {
    files, err := os.ReadDir(collectionPath)
    if err != nil {
        return fmt.Errorf("failed to read collection directory: %w", err)
    }

    for _, file := range files {
        srcFilePath := filepath.Join(collectionPath, file.Name())
        destFilePath := filepath.Join(targetPath, file.Name())
        if err := copyToTarget(srcFilePath, destFilePath); err != nil {
            return fmt.Errorf("failed to copy %s to %s: %w", srcFilePath, destFilePath, err)
        }
    }
    return nil
}


// copyToTarget copies a single file from srcPath to destPath, creating directories as needed.
func copyToTarget(srcPath, destPath string) error {
    // Create target directory if it doesn't exist
    if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
        return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
    }

    // Open source file
    srcFile, err := os.Open(srcPath)
    if err != nil {
        return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
    }
    defer srcFile.Close()

    // Open destination file
    destFile, err := os.Create(destPath)
    if err != nil {
        return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
    }
    defer destFile.Close()

    // Copy contents from source to destination
    if _, err = io.Copy(destFile, srcFile); err != nil {
        return fmt.Errorf("failed to copy file from %s to %s: %w", srcPath, destPath, err)
    }

    return nil
}
```

## collections_test.go
```go
package collections
```

## files.go
```go
// internal/collections/files.go

package collections

import (
    "os"
    "path/filepath"
    "fmt"
)

// GetCollectionFiles returns a list of file paths in the central collection directory for a collection.
func GetCollectionFiles(collectionName string) ([]string, error) {
    collectionDir := GetCollectionPath(collectionName)
    files := []string{}

    err := filepath.Walk(collectionDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            files = append(files, path)
        }
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("failed to list files in collection %s: %w", collectionName, err)
    }

    return files, nil
}

```

## scan.go
```go
// internal/collections/scan.go

package collections

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
    "fmt"
)

// scanForCollections searches the metadata directory for collections associated with the specified path.
func ScanForCollections(dir string) ([]string, error) {
    metaDir := filepath.Join(os.Getenv("HOME"), ".config", "project-sync-tool", "meta")
    collections := []string{}

    // Read the meta directory for each collection's YAML metadata file
    entries, err := os.ReadDir(metaDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read metadata directory: %w", err)
    }

    // Check each metadata file to see if the directory matches any listed project path
    for _, entry := range entries {
        if filepath.Ext(entry.Name()) != ".yml" {
            continue
        }

        collectionName := entry.Name()[:len(entry.Name())-4] // Remove ".yml" extension
        metaFilePath := filepath.Join(metaDir, entry.Name())

        meta, err := loadCollectionMeta(metaFilePath)
        if err != nil {
            return nil, fmt.Errorf("failed to load metadata for %s: %w", collectionName, err)
        }

        for _, path := range meta.Paths {
            absPath, _ := filepath.Abs(path)
            targetAbsPath, _ := filepath.Abs(dir)
            if absPath == targetAbsPath {
                collections = append(collections, collectionName)
                break
            }
        }
    }

    return collections, nil
}

// loadCollectionMeta loads the metadata from a specified file path
func loadCollectionMeta(metaFilePath string) (CollectionMeta, error) {
    meta := CollectionMeta{}
    data, err := os.ReadFile(metaFilePath)
    if err != nil {
        return meta, fmt.Errorf("failed to read metadata file: %w", err)
    }

    if err := yaml.Unmarshal(data, &meta); err != nil {
        return meta, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }
    return meta, nil
}

```

## utils.go
```go
// internal/collections/utils.go
package collections

import (
    "errors"
    "regexp"
)

// validNameRegex enforces alphanumeric-only characters for collection names
var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// IsValidCollectionName checks if a collection name contains only alphanumeric characters
func IsValidCollectionName(name string) error {
    if !validNameRegex.MatchString(name) {
        return errors.New("collection name must contain only alphanumeric characters")
    }
    return nil
}

```

## config.go
```go
```

## main.go
```go
package main

import (
	"fmt"
	"os"
	"github.com/forsvunnet/project-sync-tool/cmd/pst"
)

func main() {
    if err := pst.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

```

