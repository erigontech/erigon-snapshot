package snapshothashes

import (
	"context"
	"testing"
)

func TestFetchSnapshotHashes(t *testing.T) {
	dat, err := fetchSnapshotHashes(context.Background(), Github, "https://raw.githubusercontent.com/erigontech/erigon-snapshot/main/mainnet.toml")
	if err != nil {
		t.Errorf("fetchSnapshotHashes() failed: %v", err)
	}
	if len(dat) == 0 {
		t.Errorf("fetchSnapshotHashes() failed: empty data")
	}
}

func TestFetchSnapshotHashesAll(t *testing.T) {
	err := LoadSnapshots(context.Background(), Github, "main")
	if err != nil {
		t.Errorf("LoadSnapshots() failed %s", err)
	}
	ok := err == nil
	if !ok {
		t.Errorf("LoadSnapshots() failed")
	}
	if len(Mainnet) == 0 {
		t.Errorf("Mainnet is empty")
	}
	if len(Sepolia) == 0 {
		t.Errorf("Sepolia is empty")
	}
	if len(Amoy) == 0 {
		t.Errorf("Amoy is empty")
	}
	if len(BorMainnet) == 0 {
		t.Errorf("BorMainnet is empty")
	}
	if len(Gnosis) == 0 {
		t.Errorf("Gnosis is empty")
	}
	if len(Chiado) == 0 {
		t.Errorf("Chiado is empty")
	}
	if len(Holesky) == 0 {
		t.Errorf("Holesky is empty")
	}
	if len(Hoodi) == 0 {
		t.Errorf("Hoodi is empty")
	}
}
