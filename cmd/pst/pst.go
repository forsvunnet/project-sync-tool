// cmd/pst/load.go

package pst

import (
    "github.com/spf13/cobra"
)

var targetDir string
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
    requireCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Specify a target directory to load the collection into")
    requireCmd.Flags().BoolVarP(&force, "force", "f", false, "Forcefully overwrite local files even if they are newer")
    pushCmd.Flags().BoolVarP(&force, "force", "f", false, "Forcefully overwrite central files even if they are newer")
}
