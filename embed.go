package snapshothashes

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"io"
	"net/http"

	_ "github.com/erigontech/erigon-snapshot/webseed"
)

//go:embed *.toml
var Tomls embed.FS

// Get embedded bytes for a chain that must exist. TODO: Reference through a helper that does the
// path manipulation.
func MustGetBytes(chain string) []byte {
	data, err := Tomls.ReadFile(chain)
	if err != nil {
		panic(err)
	}
	return data
}

func getURLByChain(branch, chain string) string {
	return fmt.Sprintf("https://erigon-snapshots.erigon.network/%s/%s.toml", branch, chain)
}

func FetchSnapshot(ctx context.Context, branch string, chain string) (hashes []byte, err error) {
	u := getURLByChain(chain, branch)
	// Try to fetch the latest snapshot hashes from the web
	return fetchSnapshotHashes(ctx, u)
}

func fetchSnapshotHashes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
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
