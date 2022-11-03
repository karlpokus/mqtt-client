fmt:
	@go fmt ./...

test: fmt
	@go test ./...

run: fmt
	@go run cmd/main.go
