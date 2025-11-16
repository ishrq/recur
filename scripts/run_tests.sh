
#!/bin/bash

echo "Running integration tests..."
cd tests/integration
go test -v -tags fts5 ./...
