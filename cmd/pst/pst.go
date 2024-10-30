// cmd/pst/load.go

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
