.DEFAULT_GOAL := dev

.PHONY: dev
dev: ## dev build
dev: clean generate vet fmt lint test mod-tidy

.PHONY: ci
ci: ## CI build
ci: clean vet fmt test mod-tidy diff

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	rm -rf dist || true
	rm -rf bin || true

.PHONY: install
install: ## go install tools
	$(call print-target)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.46.2

.PHONY: generate
generate: ## go generate
	$(call print-target)
	go generate ./...

.PHONY: vet
vet: ## go vet
	$(call print-target)
	go vet ./...

.PHONY: fmt
fmt: ## go fmt
	$(call print-target)
	go fmt ./...

.PHONY: lint
lint: ## golangci-lint
	$(call print-target)
	golangci-lint run

.PHONY: test
test: ## go test with race detector and code covarage
	$(call print-target)
	mkdir -p bin
	go test -race -covermode=atomic -coverprofile=bin/coverage.out `go list ./... | grep -v mocks`
	go tool cover -html=bin/coverage.out -o bin/coverage.html

.PHONY: mod-tidy
mod-tidy: ## go mod tidy
	$(call print-target)
	go mod tidy

.PHONY: diff
diff: ## git diff
	$(call print-target)
	git diff --exit-code
	RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

.PHONY: build
build: ## goreleaser --snapshot --skip-publish --rm-dist
	$(call print-target)
	curl -sfL https://goreleaser.com/static/run | bash --snapshot --skip-publish --rm-dist

.PHONY: run
run: ## go run
	@go run -race .

.PHONY: go-clean
go-clean: ## go clean build, test and modules caches
	$(call print-target)
	go clean -r -i -cache -testcache -modcache

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef