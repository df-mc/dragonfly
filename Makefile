.PHONY: lint

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8 run ./...
