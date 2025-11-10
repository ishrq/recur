.PHONY: build
build:
	go build -tags fts5 -o recur cmd/recur/main.go

.PHONY: install
install:
	go install -tags fts5 ./cmd/recur
