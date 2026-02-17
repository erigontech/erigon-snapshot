#!/usr/bin/env python3
"""Analyze a snapshot PR diff file and classify all changes.

Usage: python3 analyze_diff.py <diff_file>

The diff file should be the raw output of `gh pr diff`.
Outputs structured text sections for hash changes, merges, version upgrades,
new data pruned from MDBX, and unexpected deletions.
"""

import re
import sys
from collections import defaultdict


def parse_diff(path):
    removed = {}
    added = {}
    with open(path) as f:
        for line in f:
            line = line.rstrip("\n")
            if not line or line[0] not in ("+", "-"):
                continue
            m = re.match(r"^([+-])'([^']+)'\s*=\s*'([a-f0-9]+)'", line)
            if not m:
                continue
            sign, fname, hsh = m.group(1), m.group(2), m.group(3)
            if sign == "-":
                removed[fname] = hsh
            else:
                added[fname] = hsh
    return removed, added


def parse_filename(fname):
    if fname in ("salt-blocks.txt", "salt-state.txt"):
        return {"cat": "other", "fname": fname}
    m = re.match(r"^caplin/(v[\d.]+)-(\d+)-(\d+)-([^.]+)\.(.+)$", fname)
    if m:
        return {"cat": "caplin", "ver": m[1], "s": int(m[2]), "e": int(m[3]), "dt": m[4], "ext": m[5]}
    m = re.match(r"^(accessor|domain|history|idx)/(v[\d.]+)-([^.]+)\.(\d+)-(\d+)\.(.+)$", fname)
    if m:
        return {"cat": m[1], "ver": m[2], "dt": m[3], "s": int(m[4]), "e": int(m[5]), "ext": m[6]}
    m = re.match(r"^(v[\d.]+)-(\d+)-(\d+)-(transactions-to-block|[^.]+)\.(.+)$", fname)
    if m:
        return {"cat": "blocks", "ver": m[1], "dt": m[4], "s": int(m[2]), "e": int(m[3]), "ext": m[5]}
    return {"cat": "unknown", "fname": fname}


def hgroup(cat):
    if cat in ("accessor", "domain", "history", "idx"):
        return "state"
    if cat == "caplin":
        return "cl"
    if cat == "blocks":
        return "el"
    return "other"


def classify(removed, added):
    # 1. Hash changes
    hash_changes = []
    for fname in removed:
        if fname in added and removed[fname] != added[fname]:
            hash_changes.append((fname, removed[fname], added[fname]))
    hash_change_fnames = set(f for f, _, _ in hash_changes)

    # 2. Build groups by (cat, dt, ext)
    groups = defaultdict(lambda: {"rem": [], "add": []})
    for f, h in removed.items():
        if f in hash_change_fnames:
            continue
        info = parse_filename(f)
        if info["cat"] in ("other", "unknown"):
            continue
        groups[(info["cat"], info["dt"], info["ext"])]["rem"].append({**info, "fname": f})
    for f, h in added.items():
        if f in hash_change_fnames:
            continue
        info = parse_filename(f)
        if info["cat"] in ("other", "unknown"):
            continue
        groups[(info["cat"], info["dt"], info["ext"])]["add"].append({**info, "fname": f})

    merges = []
    version_upgrades_list = []
    frontier = []
    unexpected = []
    explained_r = set()
    explained_a = set()

    for key, data in sorted(groups.items()):
        cat, dt, ext = key
        rem = sorted(data["rem"], key=lambda x: (x["s"], x["e"]))
        add = sorted(data["add"], key=lambda x: (x["s"], x["e"]))

        for a in add:
            covered = [r for r in rem if r["s"] >= a["s"] and r["e"] <= a["e"] and r["fname"] not in explained_r]
            if covered:
                old_vers = list(set(r["ver"] for r in covered))
                is_vu = a["ver"] not in set(r["ver"] for r in covered)
                info = {
                    "cat": cat, "dt": dt, "ext": ext,
                    "rem_ranges": [(r["s"], r["e"], r["ver"]) for r in covered],
                    "add_range": (a["s"], a["e"], a["ver"]),
                    "is_vu": is_vu, "old_vers": old_vers, "new_ver": a["ver"],
                }
                if is_vu:
                    version_upgrades_list.append(info)
                if len(covered) >= 2 or (len(covered) == 1 and (covered[0]["s"] != a["s"] or covered[0]["e"] != a["e"])):
                    merges.append(info)
                for r in covered:
                    explained_r.add(r["fname"])
                explained_a.add(a["fname"])

        for a in add:
            if a["fname"] not in explained_a:
                if not any(r["s"] < a["e"] and r["e"] > a["s"] for r in rem):
                    frontier.append({"cat": cat, "dt": dt, "ext": ext, "s": a["s"], "e": a["e"], "ver": a["ver"], "fname": a["fname"]})
                    explained_a.add(a["fname"])

        for r in rem:
            if r["fname"] not in explained_r:
                unexpected.append({"cat": cat, "dt": dt, "ext": ext, "s": r["s"], "e": r["e"], "ver": r["ver"], "fname": r["fname"]})

    return hash_changes, merges, version_upgrades_list, frontier, unexpected


def print_report(removed, added, hash_changes, merges, version_upgrades_list, frontier, unexpected):
    # Hash changes
    print("=== HASH CHANGES ===")
    for f, oh, nh in hash_changes:
        p = parse_filename(f)
        print(f"  [{hgroup(p['cat'])}] {f}  old={oh}  new={nh}")
    print(f"  count={len(hash_changes)}")

    # Unexpected deletions
    print("=== UNEXPECTED DELETIONS ===")
    for u in unexpected:
        print(f"  [{hgroup(u['cat'])}] {u['fname']}")
    print(f"  count={len(unexpected)}")

    # Merges table
    print("=== MERGES TABLE ===")
    mp = defaultdict(list)
    for m in merges:
        rr = tuple((s, e) for s, e, v in m["rem_ranges"])
        ar = (m["add_range"][0], m["add_range"][1])
        ov = tuple(sorted(m["old_vers"]))
        pk = (hgroup(m["cat"]), m["cat"], rr, ar, m["new_ver"], ov, m["is_vu"])
        mp[pk].append(f"{m['dt']} (.{m['ext']})")

    for (hg, cat, rr, ar, nv, ov, is_vu), items in sorted(mp.items()):
        old_r = ", ".join(f"{s}-{e}" for s, e in rr)
        note = ""
        if is_vu:
            note = f"cross-version: absorbs {','.join(ov)}"
        else:
            mixed = [v for v in ov if v != nv]
            if mixed:
                note = f"cross-version: absorbs {','.join(mixed)}"
        types_str = ", ".join(sorted(set(items)))
        print(f"  [{hg}] | {cat} | {types_str} | {old_r} | {ar[0]}-{ar[1]} [{nv}] | {note}")

    # Version upgrades
    print("=== VERSION UPGRADES ===")
    vup = defaultdict(list)
    for vu in version_upgrades_list:
        vk = (hgroup(vu["cat"]), vu["cat"], tuple(sorted(vu["old_vers"])), vu["new_ver"])
        rr = [(s, e) for s, e, v in vu["rem_ranges"]]
        vup[vk].append(f"{vu['dt']} (.{vu['ext']}): {', '.join(f'{s}-{e}' for s, e in rr)} -> {vu['add_range'][0]}-{vu['add_range'][1]}")
    for (hg, cat, ov, nv), items in sorted(vup.items()):
        print(f"  [{hg}] {cat}: {','.join(ov)} -> {nv}")
        for i in sorted(items):
            print(f"    {i}")

    # Frontier / new data pruned from MDBX
    print("=== NEW DATA PRUNED FROM MDBX ===")
    fg = defaultdict(list)
    for f in frontier:
        fg[(hgroup(f["cat"]), f["cat"])].append(f)
    for (hg, cat), items in sorted(fg.items()):
        items_sorted = sorted(items, key=lambda x: (x["ext"], x["dt"], x["s"], x["e"], x["ver"]))
        print(f"  [{hg}] {cat}: {len(items)} files")
        for item in items_sorted:
            print(f"    {item['fname']}")

    # Totals
    print(f"=== TOTALS: removed={len(removed)} added={len(added)} hash_changes={len(hash_changes)} merges={len(merges)} vu={len(version_upgrades_list)} frontier={len(frontier)} unexpected={len(unexpected)} ===")


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <diff_file>", file=sys.stderr)
        sys.exit(1)

    diff_file = sys.argv[1]
    removed, added = parse_diff(diff_file)
    hash_changes, merges, version_upgrades_list, frontier, unexpected = classify(removed, added)
    print_report(removed, added, hash_changes, merges, version_upgrades_list, frontier, unexpected)


if __name__ == "__main__":
    main()
