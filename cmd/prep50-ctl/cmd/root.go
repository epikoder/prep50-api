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
	initializeCmd := &cobra.Command{
		Use:   "init",
		Run:   initialize,
		Short: "Initialize the application on first Run",
	}
	initializeCmd.Flags().BoolP("auto", "y", false, "Setup app automatically")
	initializeCmd.Flags().BoolP("exams", "e", true, "Initialize exams table")
	initializeCmd.Flags().BoolP("providers", "p", true, "Initialize auth-providers table")
	initializeCmd.Flags().BoolP("admin", "a", true, "Setup admin")
	initializeCmd.Flags().BoolP("jwt", "j", true, "Setup JWT token")
	initializeCmd.Flags().BoolP("jwtf", "f", false, "Force create JWT token")

	rootCmd.AddCommand(initializeCmd)
}
