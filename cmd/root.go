package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// currentReleaseVersion is used to print the version the user currently has downloaded
const currentReleaseVersion = "v0.1.0"

// rootCmd is the base for all commands.
var rootCmd = &cobra.Command{
	Use:   "appleone [command]",
	Short: "appleone is an Apple 1 emulator",
	Long:  "appleone is an Apple 1 emulator",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Requires at least 1 argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Unknown command. Try `appleone help` for more information")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs appleone according to the user's command/subcommand(s)/flag(s)
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
