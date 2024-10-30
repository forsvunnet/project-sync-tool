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


