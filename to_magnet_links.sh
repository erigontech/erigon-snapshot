#!/usr/bin/env bash

# Function for URL encoding the trackers
rawurlencode() {
  local string="${1}"
  local strlen=${#string}
  local encoded=""
  local pos c o

  for (( pos=0 ; pos<strlen ; pos++ )); do
     c=${string:$pos:1}
     case "$c" in
        [-_.~a-zA-Z0-9] ) o="${c}" ;;
        * )               printf -v o '%%%02x' "'$c"
     esac
     encoded+="${o}"
  done
  echo "${encoded}"    # You can either set a return variable (FASTER) 
}


# Download TOML file
wget -O erigon_torrents.toml https://raw.githubusercontent.com/ledgerwatch/erigon-snapshot/main/mainnet.toml

# Download Trackers
wget -O trackers_best.txt https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best.txt

# Remove empty lines from trackers file
sed -i '/^[[:space:]]*$/d' trackers_best.txt

# Generate trackers string
TRACKERS=""
for LINE in $(cat trackers_best.txt)
do
    TRACKERS="${TRACKERS}&tr=$(rawurlencode $LINE)"
done

# Echo out all the magnet links with trackers
cat erigon_torrents.toml | while read LINE
do
    NAME=$(echo $LINE | awk '{print $1}' | sed "s/'//g")
    HASH=$(echo $LINE | awk '{print $3}' | sed "s/'//g")
    echo "magnet:?xt=urn:btih:${HASH}&dn=${NAME}${TRACKERS}"
done