BINARY := timebox
PKG := ./cmd/timebox

.PHONY: build

build: ## Compile the timebox binary
	go build -o $(BINARY) $(PKG)
