package command

import (
	"github.com/spf13/cobra"
	consts "github.com/svanas/ladder/constants"
)

func init() {
	rootCommand.PersistentFlags().String(consts.FLAG_API_KEY, "", "your API key (optional, CEX-only)")
	rootCommand.PersistentFlags().String(consts.FLAG_API_SECRET, "", "your API secret (optional, CEX-only)")
	rootCommand.PersistentFlags().Int(consts.FLAG_CHAIN_ID, 0, "chain ID (optional, please see https://chainlist.org)")
	rootCommand.PersistentFlags().String(consts.FLAG_PRIVATE_KEY, "", "your private key (optional, DEX-only)")
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
