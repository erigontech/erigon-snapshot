package snapshothashes

import (
	_ "embed"
)

//go:embed mainnet.toml
var Mainnet []byte

//go:embed sepolia.toml
var Sepolia []byte

//go:embed gnosis.toml
var Gnosis []byte

//go:embed chiado.toml
var Chiado []byte

//go:embed hoodi.toml
var Hoodi []byte

//go:embed arb-sepolia.toml
var ArbSepolia []byte

//go:embed bloatnet.toml
var Bloatnet []byte

// TODO: this type is used by erigon 3.4, DO NOT REMOVE IT YET; we should internalize it in erigon/main and remove it from here on erigon-snapshot/main
type SnapshotSource int

// TODO: these constants are used by erigon 3.4, DO NOT REMOVE IT YET; we should internalize them in erigon/main and remove them from here on erigon-snapshot/main
const (
	Github SnapshotSource = 0
	R2     SnapshotSource = 1
)
