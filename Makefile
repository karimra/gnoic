GO      ?= go
BINARY  ?= gnoic
PKGS    := ./...

.PHONY: all build test race cover lint fmt fmt-check vet staticcheck tidy-check vuln clean

all: lint test build

build:
	CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o $(BINARY) .

test:
	$(GO) test -count=1 $(PKGS)

race:
	$(GO) test -race -count=1 $(PKGS)

cover:
	$(GO) test -race -count=1 -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -func=coverage.out | tail -1

lint: fmt-check vet staticcheck tidy-check

fmt:
	gofmt -w .

fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "files not gofmt'd:"; echo "$$out"; exit 1; fi

vet:
	$(GO) vet $(PKGS)

staticcheck:
	$(GO) run honnef.co/go/tools/cmd/staticcheck@latest $(PKGS)

tidy-check:
	$(GO) mod tidy -diff

vuln:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKGS)

clean:
	rm -f $(BINARY) coverage.out
