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


