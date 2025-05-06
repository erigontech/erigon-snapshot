package webseed

import (
	"embed"
)

//go:embed *.toml
var Tomls embed.FS

func ForChain(chain string) []byte {
	data, err := Tomls.ReadFile(chain)
	if err != nil {
		panic(err)
	}
	return data
}
