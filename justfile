watch:
    templ generate --watch .

generate:
    go generate ./...

build: generate
    go build -o bin/shortcut cmd/main.go 

run: build
    ./bin/shortcut

clean:
    rm -rf bin/*