package cmd

import (
	"fmt"
	"os"

	"github.com/bradford-hamilton/apple-1/internal/vm"
	"github.com/spf13/cobra"
)

// runCmd runs the appleone virtual machine and waits for a shutdown signal to exit
var runCmd = &cobra.Command{
	Use:   "run `path/to/program`",
	Short: "run the Apple 1 emulator",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("The run command takes one argument: a `path/to/program`")
			os.Exit(1)
		}
		pathToProgram := os.Args[2]
		fmt.Println("path to program:", pathToProgram)

		vm := vm.New()
		go vm.Run()
		<-vm.ShutdownC
	},
}
