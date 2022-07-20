package snapshothashes

import (
	_ "embed"
)

//go:embed mainnet.toml
var Mainnet []byte

//go:embed goerli.toml
var Goerli []byte

//go:embed bsc.toml
var Bsc []byte

//go:embed ropsten.toml
var Ropsten []byte

//go:embed mumbai.toml
var Mumbai []byte

//go:embed bor-mainnet.toml
var BorMainnet []byte
