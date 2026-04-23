.PHONY: build run test test-race test-integration cover docker clean fmt vet

BINARY := resume-analyzer
PKG := ./...

build:
	CGO_ENABLED=0 go build -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test -count=1 $(PKG)

test-race:
	go test -race -count=1 $(PKG)

test-integration:
	go test -tags=integration -count=1 $(PKG)

cover:
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -func=coverage.out

fmt:
	go fmt $(PKG)

vet:
	go vet $(PKG)

docker:
	docker build -t resume-analyzer:dev .

clean:
	rm -f $(BINARY) coverage.out
