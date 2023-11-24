package flag

import (
	"fmt"
	"github.com/spf13/cobra"
	consts "github.com/svanas/ladder/constants"
)

func getBool(name string) bool {
	if exists(name) {
		value := get(name)
		return value == "" || value[0] == 'T' || value[0] == 't' || value[0] == 'Y' || value[0] == 'y'
	}
	return false
}

func getString(name string) (string, error) {
	if exists(name) {
		value := get(name)
		if value == "" {
			return "", fmt.Errorf("--%s is empty", name)
		}
		return value, nil
	}
	return "", fmt.Errorf("--%s does not exist", name)
}

// --mult=[1..2]
func Mult(cmd *cobra.Command) (float64, error) {
	value, err := GetFloat64(cmd, consts.FLAG_MULT)
	if err == nil {
		if value < 1 || value >= 2 {
			err = fmt.Errorf("--%s is invalid. valid values are between 1 and 2", consts.FLAG_MULT)
		}
	}
	return value, err
}

// --api-key=XXX
func ApiKey() (string, error) {
	return getString(consts.FLAG_API_KEY)
}

// --api-secret=YYY
func ApiSecret() (string, error) {
	return getString(consts.FLAG_API_SECRET)
}
