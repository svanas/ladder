package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().Bool("test", false, "use the testnet (if available)")
	rootCmd.PersistentFlags().Bool("debug", false, "log debug info to the console")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

var rootCmd = &cobra.Command{
	Use:   "ladder",
	Short: "incremental buying or selling of any crypto asset",
}

// Returns true if you are running a development build (not a release build), otherwise false.
func development() bool {
	return rootCmd.Version == "99.99.999"
}

func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
