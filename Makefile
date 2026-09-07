.PHONY: fmt test race vet build

fmt:
	gofmt -w .

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build ./...
