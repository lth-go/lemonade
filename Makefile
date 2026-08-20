VERSION=$(shell git describe --tags 2>/dev/null || echo dev)

GOFLAGS := -trimpath
LDFLAGS := -s -w

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o ./bin/lemonade .

install:
	go install .

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .
	goimports -w .

release:
	@command -v gox >/dev/null 2>&1 || { echo "gox is required: go install github.com/mitchellh/gox@latest"; exit 1; }
	gox --arch 'amd64' --os 'windows linux' --output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -ldflags "$(LDFLAGS)"

clean:
	rm -rf dist/ bin/

.PHONY: build install test vet fmt release clean
