package command

import (
	"github.com/spf13/cobra"
	consts "github.com/svanas/ladder/constants"
)

func init() {
	rootCommand.PersistentFlags().String(consts.FLAG_API_KEY, "", "your API key")
	rootCommand.PersistentFlags().String(consts.FLAG_API_SECRET, "", "your API secret")
	rootCommand.PersistentFlags().String(consts.FLAG_API_PASSPHRASE, "", "your API passphrase (optional)")
	rootCommand.PersistentFlags().Bool(consts.FLAG_TEST, false, "use the testnet (if available)")
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
