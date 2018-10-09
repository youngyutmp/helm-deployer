package main

//go:generate enumer -type=ResponseStatus -trimprefix=Status -json enums
//go:generate mkdir -p embedded
//go:generate mkdir -p dist
//go:generate esc -o embedded/assets.go -pkg embedded -prefix "dist/" dist

import (
	"github.com/entwico/helm-deployer/cmd"
)

func main() {
	cmd.Execute()
}
