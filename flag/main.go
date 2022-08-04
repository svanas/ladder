package flag

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Get() finds a named flag in the args list and returns its value
func get(name string) string {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-"+name) || strings.HasPrefix(arg, "--"+name) {
			i := strings.Index(arg, "=")
			if i > -1 {
				if strings.Contains(arg, "-"+name+"=") {
					return arg[i+1:]
				}
			}
		}
	}
	return ""
}

// Exists() determines if a flag exists, even if it doesn't have a value
func exists(name string) bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-"+name) || strings.HasPrefix(arg, "--"+name) {
			return true
		}
	}
	return false
}

func GetFloat64(cmd *cobra.Command, name string) (float64, error) {
	out, err := cmd.Flags().GetFloat64(name)
	if out == 0 && err == nil {
		err = fmt.Errorf("--%s cannot be zero", name)
	}
	return out, err
}

func GetString(cmd *cobra.Command, name string) (string, error) {
	out, err := cmd.Flags().GetString(name)
	if out == "" && err == nil {
		err = fmt.Errorf("--%s cannot be empty", name)
	}
	return out, err
}
