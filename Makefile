GOMOD = go mod
GOBUILD = go build -trimpath -v
GOTEST = go test -v -cover -race

ROOT = $(shell git rev-parse --show-toplevel)
BIN = dist/grdep
CMD = "./cmd/"

.PHONY: $(BIN)
$(BIN):
	$(GOBUILD) -o $@ $(CMD)

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: init
init:
	$(GOMOD) tidy

.PHONY: lint
lint: vet vuln

.PHONY: vuln
vuln:
	go tool govulncheck ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: golden
golden:
	./cmd/test/update-golden.sh
