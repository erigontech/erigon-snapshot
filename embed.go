package snapshothashes

import (
	"context"
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

//go:embed hoodi.toml
var Hoodi []byte

//go:embed arb-sepolia.toml
var ArbSepolia []byte

//go:embed bloatnet.toml
var Bloatnet []byte

type SnapshotSource int

const (
	Github SnapshotSource = 0
	R2     SnapshotSource = 1
)

func getURLByChain(source SnapshotSource, chain, branch string) string {
	if source == Github {
		return fmt.Sprintf("https://raw.githubusercontent.com/erigontech/erigon-snapshot/%s/%s.toml", branch, chain)
	} else if source == R2 {
		return fmt.Sprintf("https://erigon-snapshots.erigon.network/%s/%s.toml", branch, chain)
	}

	panic(fmt.Sprintf("unknown snapshot source: %d", source))
}

// Loads snapshots for all chains from the specified source and branch.
func LoadSnapshots(ctx context.Context, source SnapshotSource, branch string) (err error) {
	// Not going to call out that we're loading *all* chains but there's a fix coming for that.

	// Can't currently log as erigon-lib module does not exist at its canonical import location. See
	// https://github.com/erigontech/erigon/pull/16111.
	//log.Info("Loading remote snapshot hashes")
	var (
		mainnetUrl    = getURLByChain(source, "mainnet", branch)
		sepoliaUrl    = getURLByChain(source, "sepolia", branch)
		amoyUrl       = getURLByChain(source, "amoy", branch)
		borMainnetUrl = getURLByChain(source, "bor-mainnet", branch)
		gnosisUrl     = getURLByChain(source, "gnosis", branch)
		chiadoUrl     = getURLByChain(source, "chiado", branch)
		holeskyUrl    = getURLByChain(source, "holesky", branch)
		hoodiUrl      = getURLByChain(source, "hoodi", branch)
		arbSepoliaUrl = getURLByChain(source, "arb-sepolia", branch)
		bloatnetUrl   = getURLByChain(source, "bloatnet", branch)
	)
	var hashes []byte
	// Try to fetch the latest snapshot hashes from the web
	if hashes, err = fetchSnapshotHashes(ctx, source, mainnetUrl); err != nil {
		return
	}
	Mainnet = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, sepoliaUrl); err != nil {
		return
	}
	Sepolia = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, amoyUrl); err != nil {
		return
	}
	Amoy = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, borMainnetUrl); err != nil {
		return
	}
	BorMainnet = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, gnosisUrl); err != nil {
		return
	}
	Gnosis = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, chiadoUrl); err != nil {
		return
	}
	Chiado = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, holeskyUrl); err != nil {
		return
	}
	Holesky = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, hoodiUrl); err != nil {
		return
	}
	Hoodi = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, arbSepoliaUrl); err != nil {
		return
	}
	ArbSepolia = hashes

	if hashes, err = fetchSnapshotHashes(ctx, source, bloatnetUrl); err != nil {
		return
	}
	Bloatnet = hashes

	return nil
}

func fetchSnapshotHashes(ctx context.Context, source SnapshotSource, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if source == R2 {
		insertCloudflareHeaders(req)
	}
	resp, err := http.DefaultClient.Do(req)
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

// TODO: this was taken originally from erigon repo, downloader.go; we need to decide on a unique place to store such headers
var cloudflareHeaders = http.Header{
	"lsjdjwcush6jbnjj3jnjscoscisoc5s": []string{"I%OSJDNFKE783DDHHJD873EFSIVNI7384R78SSJBJBCCJBC32JABBJCBJK45"},
}

func insertCloudflareHeaders(req *http.Request) {
	for key, value := range cloudflareHeaders {
		req.Header[key] = value
	}
}
