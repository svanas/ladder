package flag

import (
	"fmt"
	"github.com/spf13/cobra"
)

func GetFloat64(cmd *cobra.Command, name string) (float64, error) {
	out, err := cmd.Flags().GetFloat64(name)
	if out == 0 && err == nil {
		err = fmt.Errorf("--%s cannot be zero", name)
	}
	return out, err
}
