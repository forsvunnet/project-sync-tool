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

