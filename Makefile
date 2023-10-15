test:
	go test -v --race ./...

lint:
	golangci-lint run