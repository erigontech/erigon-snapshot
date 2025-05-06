package snapshothashes

import (
	"iter"
	"testing"
)

func TestFetchMainnetMainSnapshotHashes(t *testing.T) {
	hashes, err := FetchSnapshot(t.Context(), "main", "mainnet")
	if err != nil {
		t.Fatalf("error fetching snapshot hashes: %v", err)
	}
	if len(hashes) == 0 {
		t.Fatal("snapshot hashes are empty")
	}
}

// Can't pull this from erigon-lib which seems to be where the canonical list is. Also can't find a
// strongly-typed chain enum.
func allChains() iter.Seq[string] {
	return func(yield func(string) bool) {
		entries, err := Tomls.ReadDir(".")
		if err != nil {
			panic(err)
		}
		for _, e := range entries {
			if !yield(e.Name()) {
				return
			}
		}
	}
}

func TestFetchSnapshotHashesAll(t *testing.T) {
	for chain := range allChains() {
		// Well technically this branch name isn't going to always be correct.
		hashes, err := FetchSnapshot(t.Context(), "main", chain)
		if err != nil {
			t.Errorf("failed to fetch snapshot hashes for %v: %v", chain, err)
			continue
		}
		if len(hashes) == 0 {
			t.Errorf("snapshot hashes for %v are empty", chain)
			continue
		}
	}
}
