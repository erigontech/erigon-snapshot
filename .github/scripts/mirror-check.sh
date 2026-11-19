#!/usr/bin/env bash
# Checks that the CDN bucket holds the same *.toml files as the head of each published branch of this
# repository, and copies the ones that differ. The Mirror workflow only runs when a PR is merged and can
# fail or be cancelled; this check catches what it missed.
#
# Branches checked (minus EXCLUDE_PATTERNS, and only ONLY_PATTERNS when set):
#   - every branch that already has *.toml files in the bucket (published by the Mirror workflow), and
#   - every branch matching BRANCH_PATTERNS whose last commit is at most NEW_BRANCH_MAX_AGE_DAYS old.
# Bucket folders that belong to no branch are only reported.
#
# Environment:
#   REMOTE                   bucket root, e.g. erigon-snapshots:public-erigon-snapshots (rclone uses RCLONE_CONFIG)
#   REPO_URL                 git URL of this repository
#   ALERT                    file that receives one line per finding that needs attention
#   BRANCH_PATTERNS          space-separated globs for new branches          (default: "main release/*")
#   EXCLUDE_PATTERNS         space-separated globs never checked             (default: none)
#   ONLY_PATTERNS            space-separated globs, check only these         (default: none = all)
#   NEW_BRANCH_MAX_AGE_DAYS                                                  (default: 90)
#   REPAIR                   true: copy differing files; false: report only  (default: true)
#   REMOTE_TYPE              expected rclone backend type of the remote      (default: s3)
#   REPORT                   markdown report                                 (default: /dev/stdout)
# Exit status: 1 if the bucket could not be checked or a branch could not be repaired, else 0.

set -euo pipefail

: "${REMOTE:?}" "${REPO_URL:?}" "${ALERT:?}"
BRANCH_PATTERNS=${BRANCH_PATTERNS:-"main release/*"}
EXCLUDE_PATTERNS=${EXCLUDE_PATTERNS:-}
ONLY_PATTERNS=${ONLY_PATTERNS:-}
NEW_BRANCH_MAX_AGE_DAYS=${NEW_BRANCH_MAX_AGE_DAYS:-90}
REPAIR=${REPAIR:-true}
REMOTE_TYPE=${REMOTE_TYPE:-s3}
REPORT=${REPORT:-/dev/stdout}

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

tomls() { (cd "$1" && find . -type f -name '*.toml' | sed 's|^\./||' | sort); } # *.toml paths in DIR
toml_count() { tomls "$1" | wc -l | tr -d ' '; }

# count_extra SRC BRANCH: sets n_extra, the number of *.toml files in the bucket folder of BRANCH (as listed
# at the start) that SRC does not have
count_extra() {
	n_extra=$(comm -13 <(tomls "$1") <(awk -v p="$2/" 'index($0, p) == 1 { print substr($0, length(p) + 1) }' "$WORK/bucket.txt" | sort) |
		wc -l | tr -d ' ')
}

file_list() { # first 5 lines of FILE, comma separated
	local n list
	n=$(wc -l <"$1" | tr -d ' ')
	list=$(head -n 5 "$1" | paste -sd, - | sed 's/,/, /g')
	((n > 5)) && list+=" (+$((n - 5)) more)"
	echo "$list"
}

# compare SRC BRANCH: byte-compares the *.toml files of SRC with the bucket folder of BRANCH and sets n_equal,
# n_missing (not in the bucket), n_differ and n_extra (see count_extra); $WORK/diff.txt lists the files that
# are missing or differ in the bucket. Fails when rclone could not compare every file of SRC.
compare() {
	local combined=$WORK/combined rc=0
	: >"$combined"
	# --one-way: files only in the bucket are not differences, so a non-zero exit status without a missing or
	# different file means that the comparison itself failed (e.g. a listing error).
	# rclone also logs every difference as an error; keep its log only when the comparison failed.
	rclone check "$1" "$REMOTE/$2" --one-way --include '*.toml' --download --combined "$combined" \
		--log-level ERROR 2>"$WORK/check.log" || rc=$?
	# combined report: "=" identical, "+" only in SRC, "*" different, "!" error
	n_equal=$(grep -c '^= ' "$combined" || true)
	n_missing=$(grep -c '^+ ' "$combined" || true)
	n_differ=$(grep -c '^\* ' "$combined" || true)
	{ grep -E '^[+*] ' "$combined" || true; } | cut -c3- | sort >"$WORK/diff.txt"
	count_extra "$1" "$2"
	if ((n_equal + n_missing + n_differ != $(toml_count "$1"))) || grep -q '^! ' "$combined" ||
		((rc != 0 && n_missing + n_differ == 0)); then
		cat "$WORK/check.log" >>"$WORK/rclone.log"
		return 1
	fi
}

# process BRANCH PUBLISHED: brings the bucket folder of BRANCH in line with the branch head.
# Each round fetches the current head and compares it with the bucket; when it differs, the differing files
# are copied and the next round verifies the copy (and picks up a head that moved meanwhile). The last round
# only verifies.
process() {
	local name=$1 published=$2 src=$WORK/src round age n_bad=0 copied note
	: >"$WORK/copied.txt"
	for round in 1 2 3; do
		if ! git -C "$WORK/git" fetch -q --depth 1 --no-tags "$REPO_URL" "refs/heads/$name"; then
			fail "$name" "cannot fetch the branch"
			return 0
		fi
		if [[ $published == false && $round == 1 ]]; then
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
		if [[ $(toml_count "$src") == 0 ]]; then
			if [[ $published == true ]]; then
				count_extra "$src" "$name"
				report "| \`$name\` | no toml files in the branch | $n_extra toml files only in the bucket |"
			else
				report "| \`$name\` | skipped: no toml files | |"
			fi
			return 0
		fi

		if [[ $published == true || $round -gt 1 ]]; then
			if ! compare "$src" "$name"; then
				fail "$name" "could not compare with the bucket"
				return 0
			fi
		else # nothing in the bucket yet
			n_equal=0 n_differ=0 n_extra=0 n_missing=$(toml_count "$src")
			tomls "$src" >"$WORK/diff.txt"
		fi
		n_bad=$((n_missing + n_differ))
		((n_bad == 0)) && break

		if [[ $REPAIR != true ]]; then
			report "| \`$name\` | **OUT OF SYNC**, not repaired: $n_bad files ($(file_list "$WORK/diff.txt")) | |"
			alert "$name: $n_bad toml files on the CDN differ from the branch head, not repaired: $(file_list "$WORK/diff.txt")"
			return 0
		fi
		((round == 3)) && break # a copy now would stay unverified
		# only the files that differ; --ignore-times: upload them even if size and time happen to match
		if ! rclone copy "$src" "$REMOTE/$name" --files-from "$WORK/diff.txt" --ignore-times --log-level ERROR 2>>"$WORK/rclone.log"; then
			fail "$name" "copy to the bucket failed"
			return 0
		fi
		sort -u -o "$WORK/copied.txt" "$WORK/copied.txt" "$WORK/diff.txt"
	done
	if ((n_bad > 0)); then
		fail "$name" "$n_bad toml files still differ after copying"
		return 0
	fi

	note=""
	((n_extra > 0)) && note="$n_extra toml files only in the bucket"
	copied=$(wc -l <"$WORK/copied.txt" | tr -d ' ')
	if ((copied == 0)); then
		report "| \`$name\` | in sync ($n_equal files) | $note |"
	elif [[ $published == false ]]; then
		report "| \`$name\` | **PUBLISHED** new branch: $copied files | $note |"
		alert "$name: new branch published to the CDN ($copied toml files)"
	else
		report "| \`$name\` | **REPAIRED**: copied $copied files ($(file_list "$WORK/copied.txt")) | $note |"
		alert "$name: $copied toml files were missing or outdated on the CDN and have been copied ($(file_list "$WORK/copied.txt"))"
	fi
}

# 1. the rclone remote must exist and have a type (a config without "type" broke every Mirror run in 2026-11)
remote_name=${REMOTE%%:*}
if ! rclone listremotes --long 2>>"$WORK/rclone.log" |
	awk -v r="$remote_name:" -v t="$REMOTE_TYPE" '$1 == r && $2 == t { ok = 1 } END { exit !ok }'; then
	echo "::error::rclone config has no remote '$remote_name' of type '$REMOTE_TYPE'"
	alert "rclone config has no remote '$remote_name' of type '$REMOTE_TYPE' (check the MIRROR_CONF secret)"
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
	matches "$name" "$EXCLUDE_PATTERNS" && continue
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
