.PHONY: build test test-verbose clean

build:
	go build -tags fts5 -o recur cmd/recur/main.go

test:
	cd tests/integration && go test -tags fts5 ./...

test-verbose:
	cd tests/integration && go test -v -tags fts5 ./...

test-coverage:
	cd tests/integration && go test -tags fts5 -coverprofile=coverage.out ./...
	go tool cover -html=tests/integration/coverage.out

clean:
	rm -f recur
	rm -f tests/integration/coverage.out

install:
	go install -tags fts5 ./cmd/recur
