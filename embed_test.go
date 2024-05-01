package snapshothashes

import "testing"

func TestFetchSnapshotHashes(t *testing.T) {
	dat, err := fetchSnapshotHashes("https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/mainnet.toml")
	if err != nil {
		t.Errorf("fetchSnapshotHashes() failed: %v", err)
	}
	if len(dat) == 0 {
		t.Errorf("fetchSnapshotHashes() failed: empty data")
	}
}
