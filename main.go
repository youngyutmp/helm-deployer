package main

//go:generate enumer -type=ResponseStatus -json enums
//go:generate mkdir -p embedded
//go:generate mkdir -p dist
//go:generate esc -o embedded/assets.go -pkg embedded -prefix "dist/" dist

import (
	"fmt"
	"os"

	"github.com/entwico/helm-deployer/cmd"
)

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to run command: %v\n", err)
		os.Exit(1)
	}
}
