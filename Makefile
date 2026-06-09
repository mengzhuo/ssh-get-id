.PHONY: test build lint cover clean vet

test:
	go test ./... -count=1

cover:
	go test ./... -coverprofile=coverage.txt -covermode=atomic
	go tool cover -func=coverage.txt

lint:
	golangci-lint run ./...

vet:
	go vet ./...

build:
	go build -o ssh-get-id ./cmd/ssh-get-id/

clean:
	rm -f ssh-get-id coverage.txt
