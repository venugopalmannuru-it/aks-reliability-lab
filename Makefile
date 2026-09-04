.DEFAULT_GOAL := help

IMAGE ?= aks-reliability-lab:local
CHART := deploy/helm/aks-reliability-lab

.PHONY: help fmt fmt-check test vet build run docker-build docker-run compose-up compose-down helm-lint helm-template k8s-validate schema-validate validate

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format Go source
	gofmt -w ./cmd ./internal

fmt-check: ## Fail if Go source is not formatted
	@test -z "$$(gofmt -l ./cmd ./internal)" || (gofmt -l ./cmd ./internal && exit 1)

test: ## Run unit tests with the race detector
	go test -race -coverprofile=coverage.out ./...

vet: ## Run Go static analysis
	go vet ./...

build: ## Build the local binary
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -o bin/platform-api ./cmd/api

run: ## Run the service locally
	go run ./cmd/api

docker-build: ## Build the hardened container image
	docker build --build-arg VERSION=local -t $(IMAGE) .

docker-run: ## Run the container locally
	docker run --rm --read-only --cap-drop=ALL --security-opt=no-new-privileges:true -p 8080:8080 $(IMAGE)

compose-up: ## Start the local Compose stack
	docker compose up --build -d

compose-down: ## Stop the local Compose stack
	docker compose down --remove-orphans

helm-lint: ## Lint all supported Helm value profiles
	helm lint $(CHART)
	helm lint $(CHART) -f $(CHART)/values-development.yaml
	helm lint $(CHART) -f $(CHART)/values-production.yaml

helm-template: ## Render production Helm resources
	helm template reliability $(CHART) -f $(CHART)/values-production.yaml > /tmp/aks-reliability-lab-rendered.yaml

k8s-validate: ## Render raw Kubernetes resources with Kustomize
	kubectl kustomize deploy/kubernetes/base > /tmp/aks-reliability-lab-kustomize.yaml

schema-validate: helm-template k8s-validate ## Validate schemas when kubeconform is installed
	@if command -v kubeconform >/dev/null 2>&1; then \
		kubeconform -strict -summary -kubernetes-version 1.35.0 \
			/tmp/aks-reliability-lab-rendered.yaml /tmp/aks-reliability-lab-kustomize.yaml; \
	else \
		echo "kubeconform not installed; schema validation is enforced in CI"; \
	fi

validate: fmt-check vet test helm-lint schema-validate ## Run local quality gates
