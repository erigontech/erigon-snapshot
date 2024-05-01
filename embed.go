package snapshothashes

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"

	_ "github.com/ledgerwatch/erigon-snapshot/webseed"
)

//go:embed mainnet.toml
var Mainnet []byte

//go:embed goerli.toml
var Goerli []byte

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

func init() {
	var (
		mainnetUrl    = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/mainnet.toml"
		goerliUrl     = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/goerli.toml"
		sepoliaUrl    = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/sepolia.toml"
		mumbaiUrl     = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/mumbai.toml"
		amoyUrl       = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/amoy.toml"
		borMainnetUrl = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/bor-mainnet.toml"
		gnosisUrl     = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/gnosis.toml"
		chiadoUrl     = "https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/chiado.toml"
	)

	// Try to fetch the latest snapshot hashes from the web
	if hashes, err := fetchSnapshotHashes(mainnetUrl); err == nil {
		Mainnet = hashes
	}
	if hashes, err := fetchSnapshotHashes(goerliUrl); err == nil {
		Goerli = hashes
	}
	if hashes, err := fetchSnapshotHashes(sepoliaUrl); err == nil {
		Sepolia = hashes
	}
	if hashes, err := fetchSnapshotHashes(mumbaiUrl); err == nil {
		Mumbai = hashes
	}
	if hashes, err := fetchSnapshotHashes(amoyUrl); err == nil {
		Amoy = hashes
	}
	if hashes, err := fetchSnapshotHashes(borMainnetUrl); err == nil {
		BorMainnet = hashes
	}
	if hashes, err := fetchSnapshotHashes(gnosisUrl); err == nil {
		Gnosis = hashes
	}
	if hashes, err := fetchSnapshotHashes(chiadoUrl); err == nil {
		Chiado = hashes
	}
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
