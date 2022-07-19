package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.PersistentFlags().Bool("test", false, "use the testnet (if available)")
	rootCommand.PersistentFlags().Bool("debug", false, "log debug info to the console")
	rootCommand.CompletionOptions.HiddenDefaultCmd = true
}

var rootCommand = &cobra.Command{
	Use:   "ladder",
	Short: "incremental buying or selling of any crypto asset",
}

// Returns true if you are running a development build (not a release build), otherwise false.
func development() bool {
	return rootCommand.Version == "99.99.999"
}

func Execute(version string) error {
	rootCommand.Version = version
	return rootCommand.Execute()
}
