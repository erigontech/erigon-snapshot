package snapshothashes

import (
	_ "embed"

	_ "github.com/erigontech/erigon-snapshot/webseed"
)

//go:embed mainnet.toml
var Mainnet []byte

//go:embed sepolia.toml
var Sepolia []byte

//go:embed mumbai.toml
var Mumbai []byte

//go:embed amoy.toml
var Amoy []byte

//go:embed bor-mainnet.toml
var BorMainnet []byte

//go:embed gnosis.toml
var Gnosis []byte

//go:embed chiado.toml
var Chiado []byte

//go:embed holesky.toml
var Holesky []byte
