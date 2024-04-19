## Erigon Snapshot

This repo contains the snapshot data used when syncing Erigon

## Contribute new files

`./build/bin/downloader torrent_hashes --datadir=<your>`

Add output as a PR to this repository. Update dependency in erigon's go.mod

## Force generate new files

If your node didn't produce files yet, but db has data (node synced):

```
# stop erigon
./build/bin/erigon snapshots retire --datadir=<your>
./build/bin/downloader torrent_create --datadir=<your>
./build/bin/downloader torrent_hashes --datadir=<your>
```

## Generating Magnet Links

This repo contains a shell script that can be used to
generate [Magnet Links](https://en.wikipedia.org/wiki/Magnet_URI_scheme)

By default it will download and generate magnet links for Ethereum Mainnet, you can override this with the `--network`
argument.

Valid networks are what you see in this repo.

Tested with Linux and OSX (wget or curl required)
