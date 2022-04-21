default: tidy fmt lint cover

cover: tidy fmt
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	gofmt -l -w -s .

lint: tidy fmt
	golangci-lint run --fix

test: tidy fmt
	go test -v -race ./...

tidy:
	go mod tidy
