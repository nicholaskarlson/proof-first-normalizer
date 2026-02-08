# proof-first-normalizer

Deterministic CSV normalizer + validator (schema-driven).

![ci](https://github.com/nicholaskarlson/proof-first-normalizer/actions/workflows/ci.yml/badge.svg)
![license](https://img.shields.io/badge/license-MIT-blue.svg)

> **Book:** *The Deterministic Finance Toolkit*
> This repo is **Project 3 of 4**. The exact code referenced in the manuscript is tagged **[`book-v1`](https://github.com/nicholaskarlson/proof-first-normalizer/tree/book-v1)**.

## Toolkit navigation

- **[proof-first-recon](https://github.com/nicholaskarlson/proof-first-recon)** — deterministic CSV reconciliation (matched/unmatched + summary JSON)
- **[proof-first-auditpack](https://github.com/nicholaskarlson/proof-first-auditpack)** — deterministic audit packs (manifest.json + sha256 + verify)
- **[proof-first-normalizer](https://github.com/nicholaskarlson/proof-first-normalizer)** — deterministic CSV normalize + validate (schema → normalized.csv/errors.csv/report.json)
- **[proof-first-finance-calc](https://github.com/nicholaskarlson/proof-first-finance-calc)** — proof-first finance calc service (Amortization v1 API + demo)

## What it does

- Validates CSVs against a simple JSON schema
- Produces deterministic outputs:
  - `normalized.csv`
  - `errors.csv`
  - `report.json`

## Quick start

Requirements:
- Go **1.22+**
- GNU Make (optional, but recommended)

```bash
# One-command proof gate
make verify

# Portable proof gate (no Makefile)
go test -count=1 ./...
go run ./cmd/normalizer demo --out ./out
```


## Usage

```bash
# Print version
go run ./cmd/normalizer version

# Demo: recomputes fixture cases and verifies outputs match goldens
go run ./cmd/normalizer demo --out ./out
```

## Output artifacts (high level)

- `normalized.csv` — canonicalized headers + normalized fields
- `errors.csv` — row-level validation failures (if any)
- `report.json` — counts, schema name, and deterministic summary stats

## Determinism contract

This project is intentionally “boring” in the best way: the same inputs must produce the same outputs.

See: **[`docs/CONVENTIONS.md`](docs/CONVENTIONS.md)** (rounding, ordering, LF, atomic writes, stable JSON, etc.).


## Handoff / maintenance

See: **[`docs/HANDOFF.md`](docs/HANDOFF.md)** (acceptance gates, troubleshooting, and “what to change (and what not to)”).


## License

MIT (see `LICENSE`).

