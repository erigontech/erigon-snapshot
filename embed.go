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
	return "main"
}

//go:embed mainnet.toml
var Mainnet []byte

//go:embed sepolia.toml
var Sepolia []byte

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

func LoadSnapshots() (fetched bool, err error) {
	var (
		mainnetUrl    = getURLByChain("mainnet")
		sepoliaUrl    = getURLByChain("sepolia")
		amoyUrl       = getURLByChain("amoy")
		borMainnetUrl = getURLByChain("bor-mainnet")
		gnosisUrl     = getURLByChain("gnosis")
		chiadoUrl     = getURLByChain("chiado")
		holeskyUrl    = getURLByChain("holesky")
	)
	var hashes []byte
	// Try to fetch the latest snapshot hashes from the web
	if hashes, err = fetchSnapshotHashes(mainnetUrl); err != nil {
		fetched = false
		return
	}
	Mainnet = hashes

	if hashes, err = fetchSnapshotHashes(sepoliaUrl); err != nil {
		fetched = false
		return
	}
	Sepolia = hashes

	if hashes, err = fetchSnapshotHashes(amoyUrl); err != nil {
		fetched = false
		return
	}
	Amoy = hashes

	if hashes, err = fetchSnapshotHashes(borMainnetUrl); err != nil {
		fetched = false
		return
	}
	BorMainnet = hashes

	if hashes, err = fetchSnapshotHashes(gnosisUrl); err != nil {
		fetched = false
		return
	}
	Gnosis = hashes

	if hashes, err = fetchSnapshotHashes(chiadoUrl); err != nil {
		fetched = false
		return
	}
	Chiado = hashes

	if hashes, err = fetchSnapshotHashes(holeskyUrl); err != nil {
		fetched = false
		return
	}
	Holesky = hashes

	fetched = true
	return fetched, nil
}

func fetchSnapshotHashes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch snapshot hashes by %q: status code %d %s", url, resp.StatusCode, resp.Status)
	}
	res, err := io.ReadAll(resp.Body)
	if len(res) == 0 {
		return nil, fmt.Errorf("empty response from %s", url)
	}
	return res, err
}
