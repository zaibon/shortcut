watch:
    templ generate --watch .

build-deps:
    gci --help > /dev/null || go install github.com/daixiang0/gci@v0.13.4
    templ --help > /dev/null || go install github.com/a-h/templ/cmd/templ@latest

fmt:
  gci write --custom-order --skip-generated {{ invocation_directory() }} -s standard -s default -s blank -s dot -s alias -s "prefix(github.com/zaibon/shortcut)" 

lint: fmt
    golangci-lint run ./...

generate:
    go generate ./...

build: generate fmt
    go build -o bin/shortcut cmd/main.go

build-linux: generate fmt
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/shortcut-linux cmd/main.go

build-dev: generate
    go build -o bin/shortcut cmd/main.go

run: build
    ./bin/shortcut

test: generate
    go test -v ./...

coverage:
    go test -v -race -coverprofile=coverage.txt -covermode=atomic  ./...

clean:
    rm -rf bin/*
