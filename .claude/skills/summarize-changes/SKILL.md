---
name: summarize-changes
description: Analyze snapshot automation PR changes and produce a structured review report. Use when asked to summarize or review a snapshot PR.
argument-hint: "<PR number or URL>"
---

Analyze the changes in a snapshot automation PR and produce a structured review report.

The PR to analyze: $PR_URL_OR_NUMBER

If $PR_URL_OR_NUMBER is a full URL, extract the PR number from it. If it is just a number, use it directly. The repo is erigontech/erigon-snapshot.

## Step 1: Validate PR Origin

Before doing any analysis, verify this PR was created by the snapshot release automation.

Run:
```
gh pr view <number> --repo erigontech/erigon-snapshot --json title,labels --jq '{title: .title, labels: [.labels[].name]}'
```

The PR **must** satisfy **both** conditions:
1. Title starts with `[automation]`
2. Has the `automation` label

If either condition is not met, **stop immediately** and respond:

> This PR was not created by the snapshot release automation (expected `[automation]` title prefix and `automation` label). This skill only analyzes automated snapshot PRs.

Do NOT proceed with any further steps.

## Step 2: Fetch PR Data and Run Analysis

Run these commands using the Bash tool. **IMPORTANT:** Use the PR number in temp file paths to avoid collisions when multiple analyses run in parallel.

1. `gh pr view <number> --repo erigontech/erigon-snapshot` to get PR title, description, and metadata
2. Save the diff to a temp file:
   ```
   gh pr diff <number> --repo erigontech/erigon-snapshot > /tmp/pr_<number>_diff.txt
   ```
3. Fetch the full toml file from the PR's head branch (with fallback for merged PRs where the branch may be deleted):
   ```
   HEAD_REF=$(gh pr view <number> --repo erigontech/erigon-snapshot --json headRefName -q .headRefName)
   MERGE_COMMIT=$(gh pr view <number> --repo erigontech/erigon-snapshot --json mergeCommit -q .mergeCommit.oid)
   TOML_FILE=$(gh pr diff <number> --repo erigontech/erigon-snapshot --name-only | head -1)
   # Try head branch first; if deleted (merged PR), fall back to merge commit, then main
   gh api "repos/erigontech/erigon-snapshot/contents/$TOML_FILE" \
     -H "Accept: application/vnd.github.raw" -F ref="$HEAD_REF" > /tmp/pr_<number>_toml.txt 2>/dev/null \
   || gh api "repos/erigontech/erigon-snapshot/contents/$TOML_FILE" \
     -H "Accept: application/vnd.github.raw" -F ref="$MERGE_COMMIT" > /tmp/pr_<number>_toml.txt 2>/dev/null \
   || gh api "repos/erigontech/erigon-snapshot/contents/$TOML_FILE" \
     -H "Accept: application/vnd.github.raw" -F ref="main" > /tmp/pr_<number>_toml.txt
   ```
4. Run the analysis script with both files:
   ```
   python3 "$(git rev-parse --show-toplevel)/.claude/skills/summarize-changes/analyze_diff.py" /tmp/pr_<number>_diff.txt /tmp/pr_<number>_toml.txt
   ```

The script (`analyze_diff.py` in the skill directory) parses the diff, classifies all changes, and outputs structured sections. Use its output to build the final report.

**CRITICAL: You MUST faithfully report the script's output.** Do NOT fabricate, guess, or paraphrase results. For every critical section (hash changes, unexpected deletions, version conflicts, version downgrades), copy the exact counts, file types, and ranges from the script output. If the script says `count=0` for version conflicts, report zero — do not infer conflicts from other observations.

### What the script detects

- **Hash Changes**: same filename in both removed and added sets with different hash (CRITICAL)
- **Range Merges**: multiple smaller removed ranges replaced by a single larger added range
- **Version Upgrades**: removed entries at one version replaced by added entries at a newer version
- **Version Downgrades**: removed entries at one version replaced by added entries at a LOWER version (CRITICAL — indicates regression)
- **New Data Pruned from MDBX**: added entries beyond the previously highest block number
- **Unexpected Deletions**: removed entries not covered by any of the above (CRITICAL)
- **Version Conflicts**: multiple versions of the same file type and range coexisting in the final toml (CRITICAL)

## Step 3: Generate Report

### File grouping

Throughout the ENTIRE report, group files into these high-level sections:

- **State Snapshots**: files under `accessor/`, `domain/`, `history/`, `idx/`
- **CL Snapshots** (Consensus Layer): files under `caplin/`
- **EL Block Snapshots** (Execution Layer): root-level files with version-range patterns (bodies, headers, transactions, beaconblocks and their indices)
- **Other Files**: root-level files without version-range patterns (e.g., `salt-blocks.txt`, `salt-state.txt`)

Apply this grouping to ALL sections: Hash Changes, Unexpected Deletions, Merged Ranges, New Data Pruned from MDBX, Version Upgrades, and Other Changes.

### Output structure

---

## Snapshot PR Summary: [PR Title]

**PR:** #N | **Chain:** chain | **Top Block:** X

---

### HASH CHANGES

If hash changes exist, use 🚨 emoji and show:

### 🚨🚨🚨 HASH CHANGES — ACTION REQUIRED

> **Changed hashes mean existing snapshot content was regenerated. Nodes that already downloaded the old version will have mismatched data.**

Group changed files by State Snapshots / CL Snapshots / EL Block Snapshots:

| File | Old Hash | New Hash |
|------|----------|----------|
| ... | ... | ... |

If NO hash changes, use ✅ emoji:

### ✅ Hash Changes

**No hash changes detected.** All existing files retain their original hashes.

---

### UNEXPECTED DELETIONS

If unexpected deletions exist, use 🚨 emoji and show:

### 🚨🚨🚨 UNEXPECTED DELETIONS — ACTION REQUIRED

> **Deleted files not accounted for by merges or version upgrades could mean data loss.**

Group by State Snapshots / CL Snapshots / EL Block Snapshots. List each with its range and explain why it's concerning.

If NO unexpected deletions, use ✅ emoji:

### ✅ Unexpected Deletions

**No unexpected deletions detected.** All removed files are accounted for by range merges or version upgrades.

---

### VERSION CONFLICTS

A version conflict means the SAME file type, SAME extension, AND SAME range (identical start-end) has multiple versions in the final toml. For example, both `accessor/v2.0-logaddrs.0-64.efi` and `accessor/v2.1-logaddrs.0-64.efi` existing simultaneously is a conflict. Different versions covering DIFFERENT ranges (e.g., v2.0 for 0-256 and v2.1 for 256-512) is NOT a conflict — that is a normal version transition boundary.

**CRITICAL: Only report conflicts that appear in the script output.** The script already applies the correct definition (same type + same range). Copy the exact types, ranges, and counts from the `=== VERSION CONFLICTS ===` section. Do NOT infer or fabricate conflicts.

If version conflicts exist (script reports count > 0), use 🚨 emoji:

### 🚨🚨🚨 VERSION CONFLICTS — DO NOT MERGE

> **Multiple versions of the same file for the same range must not coexist. This will cause download conflicts for nodes.**

Group by State Snapshots / CL Snapshots / EL Block Snapshots. Show a table with one row per conflict from the script output:

| Subdir | Type | Ext | Range | Versions |
|--------|------|-----|-------|----------|
| accessor | accounts | .vi | 0-32 | v1.1, v1.2 |

When many conflicts share the same pattern (same subdir, type, ext, versions) across consecutive ranges, they can be summarized as a single row with a range span, e.g., "0-64 through 1792-1856 (29 ranges)".

If NO version conflicts (script reports count=0), use ✅ emoji:

### ✅ Version Conflicts

**No version conflicts detected.** Each file type has exactly one version per range.

---

### Merged Ranges

Present merges in a table format, with one table per high-level group (State Snapshots / CL Snapshots / EL Block Snapshots). Sort rows by subdir (accessor, domain, history, idx, caplin, or root) then by snapshot type (datatype).

If a merge also involves a version upgrade, note it in the Notes column.

Table format:

| Subdir | Type | Ext | Old Ranges | New Range | Notes |
|--------|------|-----|------------|-----------|-------|
| accessor | code | .vi | 32-40, 40-44, 44-46 | 32-48 [v1.1] | |
| domain | accounts | .kv | 32-40, 40-44, 44-46 | 32-48 [v2.0] | cross-version: absorbs v1.1 |

When multiple types share the exact same merge pattern (same old ranges, same new range, same version), they can be combined into a single row with types comma-separated.

IMPORTANT: Keep table columns narrow so they render as proper tables in the terminal. When a merge has more than 3 old ranges, split them across multiple continuation rows. Each continuation row has empty cells for all columns except Old Ranges. Example with 8 old ranges:

| Subdir | Type | Ext | Old Ranges | New Range | Notes |
|--------|------|-----|------------|-----------|-------|
| (root) | bodies, headers, txns | .seg | 2100-2110, 2110-2120, 2120-2130 | 2100-2200 [v1.1] | |
| | | | 2130-2140, 2140-2150 | | |
| | | | 2150-2151, 2151-2152, 2152-2153 | | |

This keeps each row under ~40 chars in the Old Ranges column so the terminal renders it as a proper table.

---

### New Data Pruned from MDBX

Present in a table format, with one table per high-level group (State Snapshots / CL Snapshots / EL Block Snapshots). List every individual file, one per row.

Table format:

| Subdir | File |
|--------|------|
| accessor | accessor/v1.1-code.48-50.vi |
| | accessor/v1.1-commitment.48-50.vi |
| | accessor/v1.1-rcache.48-50.vi |
| | accessor/v1.1-storage.48-50.vi |
| domain | domain/v1.1-accounts.48-50.bt |
| | domain/v1.1-accounts.48-50.kvei |

Use continuation rows (empty Subdir cell) for subsequent files in the same subdir. Start a new subdir label when the subdir changes. Files are sorted by: subdir, extension, snapshot type, range, then version.

---

### VERSION DOWNGRADES

If version downgrades exist, use 🚨 emoji:

### 🚨🚨🚨 VERSION DOWNGRADES — DO NOT MERGE

> **Files were replaced with a LOWER version than they had before. This is a regression that needs investigation.**

Group by State Snapshots / CL Snapshots / EL Block Snapshots. Show a table:

| Subdir | Type | Ext | Old Version | New Version | Ranges |
|--------|------|-----|-------------|-------------|--------|
| accessor | accounts | .vi | v1.2 | v1.1 | 0-256, 256-512, 512-768 |

If NO version downgrades, use ✅ emoji:

### ✅ Version Downgrades

**No version downgrades detected.** All version changes go to equal or higher versions.

---

### Version Upgrades

Group by State Snapshots / CL Snapshots / EL Block Snapshots. List version transitions (e.g., v1.1 -> v2.0) by category and datatype.

---

### Other Changes

Any changes to Other Files (salt files, etc.) or anything not fitting the above categories. If none, say "No other changes."

---

### Reviewer Recommendation

Based on the four critical signals (hash changes, unexpected deletions, version conflicts, and version downgrades), add a final recommendation section:

If NO hash changes AND NO unexpected deletions AND NO version conflicts AND NO version downgrades:

### ✅ Recommendation: Safe to Approve

This PR contains only routine changes: range merges, version upgrades, and new data pruned from MDBX. No anomalies detected.

If hash changes OR unexpected deletions OR version conflicts OR version downgrades exist:

### 🚨 Recommendation: Investigation Required

This PR contains changes that need manual review before approval:
- (list each concern: N hash change(s), N unexpected deletion(s), N version conflict(s), N version downgrade(s), with brief context from the sections above)
- Version conflicts and version downgrades specifically mean "do not merge" — they will cause download conflicts or regressions for nodes

## Step 4: Offer to Post as PR Comment

After displaying the full report, ask the user if they want you to post it as a comment on the PR.

If the user confirms, post the report as a GitHub PR comment using:

```
gh pr comment <number> --repo erigontech/erigon-snapshot --body-file /tmp/pr_comment.txt
```

Before posting, write the comment body to `/tmp/pr_comment.txt`. The comment MUST start with the following header before the report content:

```
> 🤖 This report was generated by [Claude Code](https://claude.com/product/claude-code).
```

Then include the full report (everything from `## Snapshot PR Summary` onward).
