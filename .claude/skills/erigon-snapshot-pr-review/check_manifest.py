#!/usr/bin/env python3
"""Check what analyze_diff.py does not check in a snapshot manifest change.

Usage: python3 check_manifest.py <base_toml> <head_toml>

Prints one line per problem, then `=== PROBLEMS: <n> ===`. Exits 1 when n > 0.

- Every line that differs between base and head must be a `'name' = 'hex'`
  entry whose name is a snapshot file (block, state or caplin).
"""

import os
import re
import sys
from collections import Counter

here = os.path.dirname(os.path.abspath(__file__))
sys.path[:0] = [here, os.path.join(here, "..", "summarize-changes")]
from analyze_diff import parse_filename  # noqa: E402

ENTRY = re.compile(r"^'([^']+)'\s*=\s*'[a-f0-9]+'$")
SNAPSHOT_CATS = ("blocks", "caplin", "accessor", "domain", "history", "idx")


def read_lines(path):
    with open(path) as f:
        return Counter(line.rstrip() for line in f if line.strip())


def changed_line_problems(base, head):
    problems = []
    for sign, lines in (("-", base - head), ("+", head - base)):
        for line in sorted(lines):
            m = ENTRY.match(line)
            if not m:
                problems.append(f"unparsed line: {sign}{line}")
            elif parse_filename(m[1])["cat"] not in SNAPSHOT_CATS:
                problems.append(f"not a snapshot file: {sign}{m[1]}")
    return problems


def main():
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <base_toml> <head_toml>", file=sys.stderr)
        sys.exit(2)
    base, head = read_lines(sys.argv[1]), read_lines(sys.argv[2])
    problems = changed_line_problems(base, head)
    for p in problems:
        print(p)
    print(f"=== PROBLEMS: {len(problems)} ===")
    sys.exit(1 if problems else 0)


if __name__ == "__main__":
    main()
