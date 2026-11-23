#!/usr/bin/env bash
# Checks that the CDN bucket holds the same *.toml files as the head of each published branch of this
# repository, and copies the ones that differ. The Mirror workflow only runs when a PR is merged and can
# fail or be cancelled; this check catches what it missed.
#
# Branches checked (only ONLY_PATTERNS when set):
#   - every branch that already has *.toml files in the bucket (published by the Mirror workflow), and
#   - every branch matching BRANCH_PATTERNS whose last commit is at most NEW_BRANCH_MAX_AGE_DAYS old.
# Bucket folders that belong to no branch are only reported.
#
# Environment:
#   REMOTE          bucket root, e.g. erigon-snapshots:public-erigon-snapshots (rclone uses RCLONE_CONFIG)
#   REPO_URL        git URL of this repository
#   ALERT           file that receives one line per finding that needs attention
#   ONLY_PATTERNS   space-separated globs, check only these         (default: none = all)
#   REPAIR          true: copy differing files; false: report only  (default: true)
#   REPORT          markdown report                                 (default: /dev/stdout)
# Exit status: 1 if the bucket could not be checked or a branch could not be repaired, else 0.

set -euo pipefail

: "${REMOTE:?}" "${REPO_URL:?}" "${ALERT:?}"
ONLY_PATTERNS=${ONLY_PATTERNS:-}
REPAIR=${REPAIR:-true}
REPORT=${REPORT:-/dev/stdout}
BRANCH_PATTERNS="main release/*" # new branches to publish
NEW_BRANCH_MAX_AGE_DAYS=90

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
: >"$ALERT"
failed=0

report() { echo "$*" >>"$REPORT"; }
alert() { echo "$*" >>"$ALERT"; }

# rclone errors go to the public job log: hide the account id in R2 endpoints and any other 32-hex id
show_rclone_errors() {
	[[ -s $WORK/rclone.log ]] || return 0
	echo "rclone errors:"
	sed -E 's/[0-9a-fA-F]{32}/<redacted>/g' "$WORK/rclone.log" | tail -n 20
	: >"$WORK/rclone.log"
}

fail() { # fail BRANCH MESSAGE
	failed=1
	report "| \`$1\` | **FAILED**: $2 | |"
	alert "$1: FAILED: $2"
	echo "::error::$1: $2"
	show_rclone_errors
}

# matches NAME PATTERNS: NAME matches one of the space-separated globs in PATTERNS
matches() {
	local p pats
	read -r -a pats <<<"$2"
	for p in "${pats[@]}"; do
		# shellcheck disable=SC2053 # $p is a glob
		[[ $1 == $p ]] && return 0
	done
	return 1
}

file_list() { # first 5 lines of FILE, comma separated
	local n list
	n=$(wc -l <"$1" | tr -d ' ')
	list=$(head -n 5 "$1" | paste -sd, - | sed 's/,/, /g')
	((n > 5)) && list+=" (+$((n - 5)) more)"
	echo "$list"
}

# process BRANCH PUBLISHED: copies the *.toml files of the branch head that are missing or differ in the bucket
# folder of BRANCH (with REPAIR other than true, only lists them).
# --checksum compares size and MD5, which R2 keeps as the ETag of these single-part files, so only the files that
# differ are sent, and rclone verifies each upload. A merge during the run gets its own Mirror run.
process() {
	local name=$1 published=$2 src=$WORK/src combined=$WORK/combined age n_bad n_equal n_extra note dry=()
	if ! git -C "$WORK/git" fetch -q --depth 1 --no-tags "$REPO_URL" "refs/heads/$name"; then
		fail "$name" "cannot fetch the branch"
		return 0
	fi
	if [[ $published == false ]]; then
		age=$((($(date +%s) - $(git -C "$WORK/git" log -1 --format=%ct FETCH_HEAD)) / 86400))
		if ((age > NEW_BRANCH_MAX_AGE_DAYS)); then
			report "| \`$name\` | skipped: not published, last commit ${age}d ago | |"
			return 0
		fi
	fi
	rm -rf "$src" && mkdir -p "$src"
	if ! git -C "$WORK/git" archive FETCH_HEAD | tar -x -C "$src"; then
		fail "$name" "cannot extract the branch"
		return 0
	fi
	if [[ $published == false && -z $(find "$src" -type f -name '*.toml' -print -quit) ]]; then
		report "| \`$name\` | skipped: no toml files | |"
		return 0
	fi

	[[ $REPAIR == true ]] || dry=(--dry-run)
	: >"$combined"
	if ! rclone copy "$src" "$REMOTE/$name" --checksum --include '*.toml' --combined "$combined" "${dry[@]}" \
		--log-level ERROR 2>>"$WORK/rclone.log"; then
		fail "$name" "copy to the bucket failed"
		return 0
	fi
	# combined report: "=" identical, "+" missing in the bucket, "*" different, "-" only in the bucket
	{ grep -E '^[+*] ' "$combined" || true; } | cut -c3- | sort >"$WORK/diff.txt"
	n_bad=$(wc -l <"$WORK/diff.txt" | tr -d ' ')
	n_equal=$(grep -c '^= ' "$combined" || true)
	n_extra=$(grep -c '^- ' "$combined" || true)
	note=""
	((n_extra > 0)) && note="$n_extra toml files only in the bucket"

	if ((n_bad == 0)); then
		report "| \`$name\` | in sync ($n_equal files) | $note |"
	elif [[ $REPAIR != true ]]; then # report only: the summary shows it, no notification
		report "| \`$name\` | **OUT OF SYNC**, not repaired: $n_bad files ($(file_list "$WORK/diff.txt")) | $note |"
	elif [[ $published == false ]]; then
		report "| \`$name\` | **PUBLISHED** new branch: $n_bad files | $note |"
		alert "$name: new branch published to the CDN ($n_bad toml files)"
	else
		report "| \`$name\` | **REPAIRED**: copied $n_bad files ($(file_list "$WORK/diff.txt")) | $note |"
		alert "$name: $n_bad toml files were missing or outdated on the CDN and have been copied ($(file_list "$WORK/diff.txt"))"
	fi
}

# 1. the rclone remote must exist and have type s3 (a config without "type" broke every Mirror run in 2026-11)
remote_name=${REMOTE%%:*}
if ! rclone listremotes --long 2>>"$WORK/rclone.log" |
	awk -v r="$remote_name:" '$1 == r && $2 == "s3" { ok = 1 } END { exit !ok }'; then
	echo "::error::rclone config has no remote '$remote_name' of type 's3'"
	alert "rclone config has no remote '$remote_name' of type 's3' (check the MIRROR_CONF secret)"
	show_rclone_errors
	exit 1
fi

# 2. what is in the bucket and which branches exist
if ! rclone lsf -R --files-only --include '*.toml' "$REMOTE" 2>>"$WORK/rclone.log" | sort >"$WORK/bucket.txt"; then
	echo "::error::cannot list $REMOTE"
	alert "cannot list the CDN bucket"
	show_rclone_errors
	exit 1
fi
if ! git ls-remote --heads "$REPO_URL" | sed 's|\trefs/heads/|\t|' | sort -t $'\t' -k 2 >"$WORK/heads.txt"; then
	echo "::error::cannot list the branches of $REPO_URL"
	alert "cannot list the branches of $REPO_URL"
	exit 1
fi
git init -q "$WORK/git"

# a bucket path belongs to the branch whose name is a leading part of it (git forbids both "a" and "a/b")
: >"$WORK/published.txt"
awk -F '\t' -v out="$WORK/published.txt" '
	NR == FNR { branch[$2] = 1; next }
	{
		n = split($0, p, "/"); pre = ""; hit = ""
		for (i = 1; i < n; i++) { pre = (i == 1 ? p[1] : pre "/" p[i]); if (pre in branch) { hit = pre; break } }
		if (hit != "") { print hit > out; next }
		d = $0; sub(/\/?[^\/]*$/, "", d); print (d == "" ? "(bucket root)" : d)
	}' "$WORK/heads.txt" "$WORK/bucket.txt" | sort -u >"$WORK/orphans.txt"
sort -u -o "$WORK/published.txt" "$WORK/published.txt"

report "Bucket: \`$REMOTE\` · repair: $REPAIR · new branches: \`$BRANCH_PATTERNS\` (last commit ≤ ${NEW_BRANCH_MAX_AGE_DAYS}d)"
report ""
report "| Branch | Status | Note |"
report "|---|---|---|"

# 3. check each branch
while IFS=$'\t' read -r -u 3 _ name; do
	[[ -z $ONLY_PATTERNS ]] || matches "$name" "$ONLY_PATTERNS" || continue
	if grep -qxF "$name" "$WORK/published.txt"; then
		process "$name" true
	elif matches "$name" "$BRANCH_PATTERNS"; then
		process "$name" false
	fi
done 3<"$WORK/heads.txt"

if [[ -s $WORK/orphans.txt ]]; then
	report ""
	report "Bucket folders without a branch (not changed): $(paste -sd, "$WORK/orphans.txt" | sed 's/,/, /g')"
fi
show_rclone_errors
exit "$failed"
