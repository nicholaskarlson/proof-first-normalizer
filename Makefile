.PHONY: test fmt build demo verify clean

VERSION ?= dev

test:
	go test -count=1 ./...

fmt:
	gofmt -w cmd internal tests

build:
	mkdir -p bin
	go build -ldflags "-X main.version=$(VERSION)" -o bin/normalizer ./cmd/normalizer

demo:
	go run ./cmd/normalizer demo --out ./out

verify: test demo

clean:
	rm -rf ./bin ./out
