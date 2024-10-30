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

func AddToCollection(collectionName string, paths []string, force bool) error {
    collectionPath := GetCollectionPath(collectionName)

    // Clear and recreate collection directory if force is specified
    if _, err := os.Stat(collectionPath); err == nil && force {
        if err := os.RemoveAll(collectionPath); err != nil {
            return fmt.Errorf("failed to clear collection directory: %w", err)
        }
    }
    if err := os.MkdirAll(collectionPath, os.ModePerm); err != nil {
        return fmt.Errorf("failed to create collection directory: %w", err)
    }

    // Copy each specified path into the collection, preserving relative directory structure
    for _, path := range paths {
        relPath, err := filepath.Rel(".", path) // relative to the current directory
        if err != nil {
            return fmt.Errorf("failed to calculate relative path for %s: %w", path, err)
        }

        destPath := filepath.Join(collectionPath, relPath)
        if err := copyToCollection(path, destPath); err != nil {
            return fmt.Errorf("failed to add %s to collection: %w", path, err)
        }
    }

    // Save collection metadata
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


// copyToCollection copies files or directories to the collection path, creating directories as needed.
func copyToCollection(srcPath, destPath string) error {
    // Ensure the destination directory exists
    destDir := filepath.Dir(destPath)
    if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
        return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
    }

    srcInfo, err := os.Stat(srcPath)
    if err != nil {
        return fmt.Errorf("could not access path %s: %w", srcPath, err)
    }

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

func copyDirectory(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Calculate the relative path and create target path
        relPath, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }
        targetPath := filepath.Join(dst, relPath)

        if info.IsDir() {
            // Create the target directory if it's a directory
            return os.MkdirAll(targetPath, info.Mode())
        }

        // Copy file if it's a regular file
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

func copyCollectionFilesToTarget(collectionPath, targetPath string) error {
    return filepath.Walk(collectionPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Calculate the relative path from the collection base to the current file
        relPath, err := filepath.Rel(collectionPath, path)
        if err != nil {
            return fmt.Errorf("failed to calculate relative path: %w", err)
        }
        destPath := filepath.Join(targetPath, relPath)

        if info.IsDir() {
            // Create the directory in the target path
            return os.MkdirAll(destPath, info.Mode())
        }

        // Copy files using CopyFile
        return CopyFile(path, destPath)
    })
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
