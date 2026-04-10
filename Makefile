.PHONY: build test lint vet clean

build:
	go build -o bin/prism ./cmd/prism

test:
	go test ./... -v -race

lint:
	golangci-lint run

vet:
	go vet ./...

clean:
	rm -rf bin/
