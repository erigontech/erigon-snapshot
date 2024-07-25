#!/usr/bin/env bash

# Default to mainnet
NETWORK="mainnet"

# Get script arguments and set variables
while [ "$1" != "" ]; do
  case $1 in
  # Set Network
  -n | --network)
    shift
    NETWORK=$1
    ;;
  esac
  shift
done

# Convert network to lowercase
NETWORK=$(echo "$NETWORK" | tr '[:upper:]' '[:lower:]')

# Try to use wget, if not available try to use curl
if ! command -v wget &>/dev/null; then
  if ! command -v curl &>/dev/null; then
    echo "wget or curl is required to download files"
    exit 1
  else
    DOWNLOADER="curl -fso"
  fi
else
  DOWNLOADER="wget -qO"
fi

# Default case for Linux sed, just use "-i"
sedi=(-i)
case "$(uname)" in
  # For macOS, use two parameters
  Darwin*) sedi=(-i "")
esac

# Function for URL encoding the trackers
rawurlencode() {
  local string="${1}"
  local strlen=${#string}
  local encoded=""
  local pos c o

  for ((pos = 0; pos < strlen; pos++)); do
    c=${string:$pos:1}
    case "$c" in
    [-_.~a-zA-Z0-9]) o="${c}" ;;
    *) printf -v o '%%%02x' "'$c" ;;
    esac
    encoded+="${o}"
  done
  echo "${encoded}"
}

# Name of snapshot file
SNAPSHOTFILE="erigon_snapshots_${NETWORK}.toml"

# Download TOML file
URL="https://raw.githubusercontent.com/erigontech/erigon-snapshot/main/${NETWORK}.toml"
${DOWNLOADER} ${SNAPSHOTFILE} ${URL}
if [ $? -ne 0 ]; then
  echo "Failed to download ${URL}"
  exit 1
fi

# Remove quotes from TOML file
sed "${sedi[@]}" 's|["'\'']||g' ${SNAPSHOTFILE}

# Download Trackers
URL="https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best.txt"
${DOWNLOADER} trackers_best.txt ${URL}
if [ $? -ne 0 ]; then
  echo "Failed to download ${URL}"
  exit 1
fi

# Remove empty lines from trackers file
sed "${sedi[@]}" '/^[[:space:]]*$/d' trackers_best.txt

# Generate trackers string
TRACKERS=""
for LINE in $(cat trackers_best.txt); do
  TRACKERS="${TRACKERS}&tr=$(rawurlencode $LINE)"
done

# Echo out all the magnet links with trackers
cat ${SNAPSHOTFILE} | while read LINE; do
  NAME=$(echo $LINE | awk '{print $1}')
  HASH=$(echo $LINE | awk '{print $3}')
  echo "magnet:?xt=urn:btih:${HASH}&dn=${NAME}${TRACKERS}"
done
