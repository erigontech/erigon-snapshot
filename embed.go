package snapshothashes

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/erigontech/erigon-snapshot/webseed"
)

var branchReference = getBranchReference()

func getBranchReference() string {
	v, _ := os.LookupEnv("SNAPS_GIT_BRANCH")
	if v != "" {
		return v
	}
	return "5be1393c1c2eb4969b66567bdcfe2d82c53e5c42"
}

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

func getURLByChain(chain string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/erigontech/erigon-snapshot/%s/%s.toml", branchReference, chain)
}

func LoadSnapshots() (couldFetch bool) {
	var (
		mainnetUrl    = getURLByChain("mainnet")
		sepoliaUrl    = getURLByChain("sepolia")
		mumbaiUrl     = getURLByChain("mumbai")
		amoyUrl       = getURLByChain("amoy")
		borMainnetUrl = getURLByChain("bor-mainnet")
		gnosisUrl     = getURLByChain("gnosis")
		chiadoUrl     = getURLByChain("chiado")
		holeskyUrl    = getURLByChain("holesky")
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
