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

//go:embed gnosis.toml
var Gnosis []byte

//go:embed history/mainnet.toml
var MainnetHistory []byte

//go:embed history/goerli.toml
var GoerliHistory []byte

//go:embed history/bsc.toml
var BscHistory []byte

//go:embed history/ropsten.toml
var RopstenHistory []byte

//go:embed history/mumbai.toml
var MumbaiHistory []byte

//go:embed history/bor-mainnet.toml
var BorMainnetHistory []byte

//go:embed history/gnosis.toml
var GnosisHistory []byte
