package snapshothashes

import (
	_ "embed"
)

//go:embed erigon-snapshots/mainnet.toml
var mainnet []byte

//go:embed erigon-snapshots/goerli.toml
var goerli []byte

//go:embed erigon-snapshots/bsc.toml
var bsc []byte

//go:embed erigon-snapshots/ropsten.toml
var ropsten []byte

//go:embed erigon-snapshots/mumbai.toml
var mumbai []byte

//go:embed erigon-snapshots/bor-mainnet.toml
var borMainnet []byte
