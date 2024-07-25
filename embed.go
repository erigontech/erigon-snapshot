package snapshothashes

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"

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

func LoadSnapshots() (couldFetch bool) {
	var (
		mainnetUrl    = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/mainnet.toml"
		sepoliaUrl    = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/sepolia.toml"
		mumbaiUrl     = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/mumbai.toml"
		amoyUrl       = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/amoy.toml"
		borMainnetUrl = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/bor-mainnet.toml"
		gnosisUrl     = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/gnosis.toml"
		chiadoUrl     = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/chiado.toml"
		holeskyUrl    = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/holesky.toml"
	)
	var hashes []byte
	var err error
	// Try to fetch the latest snapshot hashes from the web
	if hashes, err = fetchSnapshotHashes(mainnetUrl); err != nil {
		couldFetch = false
		return
	}
	Mainnet = hashes

	if hashes, err = fetchSnapshotHashes(sepoliaUrl); err != nil {
		couldFetch = false
		return
	}
	Sepolia = hashes

	if hashes, err = fetchSnapshotHashes(mumbaiUrl); err != nil {
		couldFetch = false
		return
	}
	Mumbai = hashes

	if hashes, err = fetchSnapshotHashes(amoyUrl); err != nil {
		couldFetch = false
		return
	}
	Amoy = hashes

	if hashes, err = fetchSnapshotHashes(borMainnetUrl); err != nil {
		couldFetch = false
		return
	}
	BorMainnet = hashes

	if hashes, err = fetchSnapshotHashes(gnosisUrl); err != nil {
		couldFetch = false
		return
	}
	Gnosis = hashes

	if hashes, err = fetchSnapshotHashes(chiadoUrl); err != nil {
		couldFetch = false
		return
	}
	Chiado = hashes

	if hashes, err = fetchSnapshotHashes(holeskyUrl); err != nil {
		couldFetch = false
		return
	}
	Holesky = hashes

	couldFetch = true
	return
}

func fetchSnapshotHashes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if len(res) == 0 {
		return nil, fmt.Errorf("empty response from %s", url)
	}
	return res, err
}
