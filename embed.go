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
var tomls embed.FS

// Get embedded bytes for a chain that must exist. TODO: Reference through a helper that does the
// path manipulation.
func MustGetBytes(chain string) []byte {
	data, err := tomls.ReadFile(chain + ".toml")
	if err != nil {
		panic(err)
	}
	return data
}

type snapshotSource struct {
	UrlFormat string
	Headers   http.Header
}

var (
	Github = snapshotSource{
		UrlFormat: "https://raw.githubusercontent.com/erigontech/erigon-snapshot/%s/%s.toml",
	}
	R2 = snapshotSource{
		UrlFormat: "https://erigon-snapshots.erigon.network/%s/%s.toml",
		// TODO: this was taken originally from erigon repo, downloader.go; we need to decide on a unique place to store such headers
		Headers: http.Header{
			"lsjdjwcush6jbnjj3jnjscoscisoc5s": []string{"I%OSJDNFKE783DDHHJD873EFSIVNI7384R78SSJBJBCCJBC32JABBJCBJK45"},
		},
	}
)

func FetchSnapshotHashes(ctx context.Context, source snapshotSource, branch, chain string) ([]byte, error) {
	url := fmt.Sprintf(source.UrlFormat, branch, chain)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	for key, values := range source.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
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
