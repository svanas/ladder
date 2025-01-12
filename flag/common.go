package flag

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"

	"github.com/spf13/cobra"
	consts "github.com/svanas/ladder/constants"
)

// func getBool(name string) bool {
// 	if exists(name) {
// 		value := get(name)
// 		return value == "" || value[0] == 'T' || value[0] == 't' || value[0] == 'Y' || value[0] == 'y'
// 	}
// 	return false
// }

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

func getInt(name string) (int64, error) {
	value, err := getString(name)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, 0)
}

// --mult=[1..2]
func Mult(cmd cobra.Command) (float64, error) {
	value, err := GetFloat64(cmd, consts.FLAG_MULT)
	if err == nil {
		if value < 1 || value > 2 {
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
	if str := get(consts.FLAG_API_SECRET); str != "" {
		return str, nil
	}
	buf, err := prompt("API secret")
	if err != nil {
		return "", err
	}
	if buf == nil {
		return "", fmt.Errorf("--%s is empty", consts.FLAG_API_SECRET)
	}
	return string(buf), nil
}

// --chain-id=[1..2147483647]
func ChainId() (int64, error) {
	value, err := getInt(consts.FLAG_CHAIN_ID)
	if err == nil {
		if value < 1 || value > math.MaxInt32 {
			err = fmt.Errorf("--%s is invalid. valid values are between 1 and %d", consts.FLAG_CHAIN_ID, math.MaxInt32)
		}
	}
	return value, err
}

// --private-key=['0'..'9', 'A'..'F']
func PrivateKey() ([]byte, error) {
	if str := get(consts.FLAG_PRIVATE_KEY); str != "" {
		return hex.DecodeString(str)
	}
	buf, err := prompt("private key")
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, fmt.Errorf("--%s is empty", consts.FLAG_PRIVATE_KEY)
	}
	return hex.DecodeString(string(buf))
}
