---
name: summarize-changes
description: Analyze snapshot automation PR changes and produce a structured review report. Use when asked to summarize or review a snapshot PR.
argument-hint: "<PR number or URL>"
---

Analyze the changes in a snapshot automation PR and produce a structured review report.

The PR to analyze: $PR_URL_OR_NUMBER

If $PR_URL_OR_NUMBER is a full URL, extract the PR number from it. If it is just a number, use it directly. The repo is erigontech/erigon-snapshot.

## Step 1: Fetch PR Data and Run Analysis

Run these commands using the Bash tool:

1. `gh pr view <number> --repo erigontech/erigon-snapshot` to get PR title, description, and metadata
2. Save the diff to a temp file and run the analysis script:
   ```
   gh pr diff <number> --repo erigontech/erigon-snapshot > /tmp/pr_diff.txt && python3 "$(git rev-parse --show-toplevel)/.claude/skills/summarize-changes/analyze_diff.py" /tmp/pr_diff.txt
   ```

The script (`analyze_diff.py` in the skill directory) parses the diff, classifies all changes, and outputs structured sections. Use its output to build the final report.

### What the script detects

- **Hash Changes**: same filename in both removed and added sets with different hash (CRITICAL)
- **Range Merges**: multiple smaller removed ranges replaced by a single larger added range
- **Version Upgrades**: removed entries at one version replaced by added entries at a newer version
- **New Data Pruned from MDBX**: added entries beyond the previously highest block number
- **Unexpected Deletions**: removed entries not covered by any of the above (CRITICAL)

## Step 4: Generate Report

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

### Version Upgrades

Group by State Snapshots / CL Snapshots / EL Block Snapshots. List version transitions (e.g., v1.1 -> v2.0) by category and datatype.

---

### Other Changes

Any changes to Other Files (salt files, etc.) or anything not fitting the above categories. If none, say "No other changes."

## Step 5: Offer to Post as PR Comment

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
