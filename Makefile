.PHONY: build test vet fmt fmt-check lint ci

build:
	go build ./...

test:
	go test ./... -race -cover

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@fmtout=$$(gofmt -l .); \
	if [ -n "$$fmtout" ]; then \
		echo "gofmt needed on:"; echo "$$fmtout"; exit 1; \
	fi

lint:
	golangci-lint run ./...

ci: fmt-check vet test
