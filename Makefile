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

.PHONY: check-license
check-license:
	go-licenses check ./cmd/sakura-controller

.PHONY: add-license
add-license:
	addlicense -c "The distributed-mariadb-controller Authors" .
	go-licenses check ./cmd/sakura-controller

.PHONY: tool
tool:
	go install github.com/google/go-licenses@latest
	go install github.com/google/addlicense@latest
