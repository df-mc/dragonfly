.PHONY: lint vuln

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
