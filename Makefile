.PHONY: build build-linux run test test-race test-integration cover docker clean fmt vet lint deploy

BINARY := resume-analyzer
PKG := ./...
REMOTE_HOST ?= $(error Set REMOTE_HOST, e.g. make deploy REMOTE_HOST=10.0.0.1)
REMOTE_USER ?= deploy
REMOTE_BIN ?= /opt/resume-analyzer/resume-analyzer

build:
	CGO_ENABLED=0 go build -o $(BINARY) ./cmd/server

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o $(BINARY)-linux-arm64 ./cmd/server

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

lint:
	@command -v golangci-lint >/dev/null || { echo "install golangci-lint: https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run $(PKG)

docker:
	docker build -t resume-analyzer:dev .

deploy: build-linux
	@echo "Deploying to $(REMOTE_USER)@$(REMOTE_HOST)..."
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p /opt/resume-analyzer"
	scp $(BINARY)-linux-arm64 $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_BIN)
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "chmod +x $(REMOTE_BIN) && systemctl restart resume-analyzer 2>/dev/null || (pkill -f resume-analyzer; cd /opt/resume-analyzer && nohup ./resume-analyzer &)"
	@echo "Deployed. Checking health..."
	@sleep 2
	@ssh $(REMOTE_USER)@$(REMOTE_HOST) "curl -sf http://localhost:8080/health || echo 'Health check failed'"
	@echo "Done."

clean:
	rm -f $(BINARY) $(BINARY)-linux-arm64 coverage.out
