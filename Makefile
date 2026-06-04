-include .env

export PATH := /usr/local/go/bin:$(PATH)
export GOFLAGS ?= -mod=mod

.PHONY: help build run run-race clean lint generate format test bench-domain security-scan bin-deps

help: ## Показать все команды
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Собрать бинарник
	@mkdir -p bin
	go build -o bin/server ./cmd/server

run: ## Запустить локально
	go run ./cmd/server

run-race: ## Запустить с race detector
	go run -race ./cmd/server

clean: ## Очистить
	rm -rf bin/

lint: ## Запустить линтер
	golangci-lint run ./... -v

generate: ## go generate
	go generate ./...

format: ## Форматировать код
	gofmt -s -w .
	goimports -w -local github.com/djalben/istok-agent-core .

test: ## Тесты
	go test ./internal/... -race -count=1

bench-domain: ## Микробенчмарки домена
	go test -bench=. -benchmem -count=1 ./internal/domain/...

security-scan: ## govulncheck + gosec
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go run github.com/securego/gosec/v2/cmd/gosec@latest -exclude-dir=vendor -fmt=text -severity=medium ./...

bin-deps: ## Установить golangci-lint, goimports и прочие CLI
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
