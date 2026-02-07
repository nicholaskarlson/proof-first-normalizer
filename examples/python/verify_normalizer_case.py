#!/usr/bin/env python3
# SPDX-License-Identifier: MIT

from __future__ import annotations

import argparse
import csv
import hashlib
import json
from pathlib import Path


def sha256_file(p: Path) -> str:
    h = hashlib.sha256()
    with p.open("rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


def read_text_lf(p: Path) -> str:
    s = p.read_text(encoding="utf-8")
    if "\r\n" in s:
        raise AssertionError(f"CRLF found in {p}")
    return s


def load_json(p: Path):
    return json.loads(read_text_lf(p))


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out-root", default="out/demo", help="demo output root (contains <case>/...)")
    ap.add_argument("--case", default="case02_errors", help="case folder name")
    ap.add_argument("--compare-goldens", action="store_true", help="also compare output bytes to fixtures/expected")
    args = ap.parse_args()

    repo = Path(__file__).resolve().parents[2]
    case = args.case

    expected_dir = repo / "fixtures" / "expected" / case
    if (expected_dir / "error.txt").exists():
        # The Go demo handles expected-fail cases by comparing the error string.
        # This Python check is artifact-based, so we skip these cases.
        print(f"SKIP: {case} is an expected-fail case (fixtures/expected/{case}/error.txt)")
        return

    out_root = repo / args.out_root
    out_dir = out_root / case

    norm_p = out_dir / "normalized.csv"
    errs_p = out_dir / "errors.csv"
    rep_p = out_dir / "report.json"

    for p in (norm_p, errs_p, rep_p):
        assert p.exists(), f"missing output file: {p}"

    # LF determinism
    read_text_lf(norm_p)
    read_text_lf(errs_p)
    read_text_lf(rep_p)

    report = load_json(rep_p)
    assert report.get("tool") == "proof-first-normalizer", "report.json tool mismatch"

    # Input/schema hashes come from fixtures
    raw_p = repo / "fixtures" / "input" / case / "raw.csv"
    schema_p = repo / "fixtures" / "input" / case / "schema.json"
    assert raw_p.exists(), f"missing fixture raw.csv: {raw_p}"
    assert schema_p.exists(), f"missing fixture schema.json: {schema_p}"

    assert report.get("sha256_input") == sha256_file(raw_p), "sha256_input mismatch"
    assert report.get("sha256_schema") == sha256_file(schema_p), "sha256_schema mismatch"
    assert report.get("sha256_normalized") == sha256_file(norm_p), "sha256_normalized mismatch"
    assert report.get("sha256_errors") == sha256_file(errs_p), "sha256_errors mismatch"

    rows_total = int(report.get("rows_total"))
    rows_ok = int(report.get("rows_ok"))
    rows_error = int(report.get("rows_error"))
    assert rows_total == rows_ok + rows_error, "rows_total != rows_ok + rows_error"

    # Header order should match schema column order.
    schema = load_json(schema_p)
    want_cols = [c["name"] for c in schema.get("columns", [])]
    assert want_cols, "schema has no columns"

    with norm_p.open(newline="", encoding="utf-8") as f:
        r = csv.reader(f)
        header = next(r)
    assert header == want_cols, f"normalized header mismatch: got={header} want={want_cols}"

    # Optional: byte-for-byte compare to checked-in goldens.
    if args.compare_goldens:
        for name in ("normalized.csv", "errors.csv", "report.json"):
            want = (expected_dir / name).read_bytes()
            got = (out_dir / name).read_bytes()
            assert got == want, f"golden mismatch: {case}/{name}"

    print("OK: normalizer demo outputs are internally consistent.")


if __name__ == "__main__":
    main()
