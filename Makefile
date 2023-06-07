.PHONY: all
all: format test vet

.PHONY: sakura-all
sakura-all: all sakura-build

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: ci
ci: format test vet

.PHONY: vet
vet:
	go vet ./...

.PHONY: sakura-build
sakura-build:
	go build -o bin/sakura-controller ./cmd/sakura-controller
