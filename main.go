package main

import (
	"os"

	"github.com/svanas/ladder/command"
)

var (
	APP_VERSION = "99.99.999"
)

func main() {
	if err := command.Execute(APP_VERSION); err != nil {
		os.Exit(1)
	}
}
