.PHONY: lint

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run ./...
