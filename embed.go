package snapshothashes

import (
	_ "embed"
)

//go:embed erigon-snapshots/mainnet.toml
var Mainnet []byte

//go:embed erigon-snapshots/goerli.toml
var Goerli []byte

//go:embed erigon-snapshots/bsc.toml
var Bsc []byte

//go:embed erigon-snapshots/ropsten.toml
var Ropsten []byte

//go:embed erigon-snapshots/mumbai.toml
var Mumbai []byte

//go:embed erigon-snapshots/bor-mainnet.toml
var BorMainnet []byte
