# HANDOFF — proof-first-normalizer

This tool normalizes CSV inputs deterministically and produces a clean output + an error file.

## Build

```bash
go test -count=1 ./...
make build VERSION=v0.1.0
./bin/normalizer version
```

## Run (v0.1.0)

```bash
# Normalize a CSV
normalizer normalize --in raw.csv --schema schema.json --out outdir

# Validate without outputting files
normalizer validate --in raw.csv --schema schema.json

# Run the demo (runs fixture cases and verifies outputs match goldens)
normalizer demo --out outdir
```

## Definition of Done

- `go test -count=1 ./...` passes
- Fixtures + golden outputs committed
- Deterministic outputs (LF, stable ordering)
- Version embedded via ldflags

## What this tool is NOT

- No services/daemons
- No Docker requirement
- No network calls

## Optional: Python check (stdlib only)

```bash
# Run the Go demo first (writes + verifies goldens)
go run ./cmd/normalizer demo --out ./out/demo

# Then run the optional Python verifier on one case
python3 examples/python/verify_normalizer_case.py --out-root ./out/demo --case case02_errors
```
