.PHONY: all
all: format test vet

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

