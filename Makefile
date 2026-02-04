.PHONY: test fmt build demo clean

VERSION ?= dev

test:
	go test -count=1 ./...

fmt:
	gofmt -w cmd internal tests

build:
	mkdir -p bin
	go build -ldflags "-X main.version=$(VERSION)" -o bin/normalizer ./cmd/normalizer

demo: build
	./bin/normalizer demo --out ./out/demo

clean:
	rm -rf ./bin ./out
