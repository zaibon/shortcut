watch:
    templ generate --watch .

fmt:
  gci write --custom-order --skip-generated {{ invocation_directory() }} -s standard -s default -s blank -s dot -s alias -s "prefix(github.com/zaibon/shortcut)" 

generate:
    go generate ./...

build: generate fmt
    go build -o bin/shortcut cmd/main.go

build-dev: generate
    go build -o bin/shortcut cmd/main.go

run: build
    ./bin/shortcut

clean:
    rm -rf bin/*