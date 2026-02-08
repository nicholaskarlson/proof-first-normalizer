# proof-first-normalizer

Deterministic CSV normalizer + validator.

## What it does (v0.1.0)

- Validates CSVs against a simple JSON schema
- Produces deterministic outputs:
  - `normalized.csv`
  - `errors.csv`
  - `report.json`

## Canonical commands

```bash
# Proof gate (one command)
make verify

# Proof gates (portable, no Makefile)
go test -count=1 ./...
go run ./cmd/normalizer demo --out ./out
```

## Quickstart

```bash
go test -count=1 ./...

go run ./cmd/normalizer version
go run ./cmd/normalizer demo --out ./out
```

## Handoff

See [docs/HANDOFF.md](docs/HANDOFF.md).
