package flag

import (
	"fmt"

	consts "github.com/svanas/ladder/constants"
)

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

// --test=[true|false]
func Test() bool {
	if exists(consts.FLAG_TEST) {
		arg := get(consts.FLAG_TEST)
		return arg == "" || arg[0] == 'T' || arg[0] == 't'
	}
	return false
}

// --api-key=XXX
func ApiKey() (string, error) {
	return getString(consts.FLAG_API_KEY)
}

// --api-secret=YYY
func ApiSecret() (string, error) {
	return getString(consts.FLAG_API_SECRET)
}

// --api-passphrase=ZZZ
func ApiPassphrase() (string, error) {
	return getString(consts.FLAG_API_PASSPHRASE)
}
