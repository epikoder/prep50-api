/*
Copyright © 2022 Efedua Believe efedua.bell@gmail.com

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prep50_ctl",
	Short: "Prep50 command line tool",
	Long:  `Manage prep50 system from the command line.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	migrateCmd := &cobra.Command{
		Use: "migrate",
		Run: func(cmd *cobra.Command, args []string) {
			migrate(cmd, args)
		},
		Short: "Manage database migration",
	}
	migrateCmd.Flags().BoolP("rollback", "r", false, "Rollback migrations")
	migrateCmd.Flags().BoolP("reset", "f", false, "Reset migrations")
	migrateCmd.Flags().StringP("table", "t", "", "Run migration on a specific table")
	rootCmd.AddCommand(migrateCmd)
}
