// cmd/pst/require.go

package pst

import (
	"fmt"
	"github.com/forsvunnet/project-sync-tool/internal/collections"
	"github.com/spf13/cobra"
)


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

        // Step 1: Check for changes to ensure local files aren't newer
        changeStatus, err := collections.CheckForChanges(collectionName, targetPath)
        if err != nil {
            return fmt.Errorf("failed to check changes for collection %s: %w", collectionName, err)
        }

        // If there are local files newer than the central ones, fail the command
		if len(changeStatus.LocalNewer) > 0 && !force {
            return fmt.Errorf("require aborted: local files are newer than central files for collection %s", collectionName)
        }

        // Proceed with requiring the collection if all checks pass
        if err := collections.RequireCollection(collectionName, targetPath); err != nil {
            return fmt.Errorf("failed to require collection: %w", err)
        }

        fmt.Printf("Collection %s required successfully to %s.\n", collectionName, targetPath)
        return nil
    },
}



