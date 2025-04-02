.PHONY: all
all: format test vet build

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

.PHONY: build
build:
	go build -o bin/db-controller ./cmd/db-controller

.PHONY: check-license
check-license:
	go-licenses check ./cmd/db-controller

.PHONY: add-license
add-license:
	addlicense -c "The distributed-mariadb-controller Authors" .
	go-licenses check ./cmd/db-controller

.PHONY: tool
tool:
	go install github.com/google/go-licenses@latest
	go install github.com/google/addlicense@latest
