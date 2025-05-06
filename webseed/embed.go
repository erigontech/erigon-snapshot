package webseed

import (
	"embed"
)

//go:embed *.toml
var tomls embed.FS

func ForChain(chain string) []byte {
	data, err := tomls.ReadFile(chain + ".toml")
	if err != nil {
		panic(err)
	}
	return data
}
