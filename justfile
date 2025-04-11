watch:
    go tool templ generate --watch --proxy=http://localhost:8080 --open-browser=false

dev: generate
  #!/usr/bin/env -S parallel --shebang --ungroup --jobs {{ num_cpus() }}
  just watch
  air

fmt:
    go tool gci write --custom-order --skip-generated {{ invocation_directory() }} -s standard -s default -s blank -s dot -s alias -s "prefix(github.com/zaibon/shortcut)" 

lint: fmt
    go tool golangci-lint run ./...

generate:
    go generate ./...

build: generate fmt
    CGO_ENABLED=0 go build -o bin/shortcut cmd/*.go

build-linux: generate fmt
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/shortcut cmd/*.go

build-dev: 
    CGO_ENABLED=0 go build -tags=dev -o bin/shortcut cmd/*.go

run: build
    ./bin/shortcut

db-migrate action: build
    ./bin/shortcut migrate {{ action }}

db-create-migration name type:
    go tool goose -dir=db/migrations create {{ name }} {{ type }}

db-fix:
    go tool goose -dir=db/migrations fix

test: generate
    go test -v ./...

package: build
    docker build -t zaibon/shortcut:latest .

coverage:
    go test -v -race -coverprofile=coverage.txt -covermode=atomic  ./...

enable-env kind:
    ln -sf .env-{{ kind }} .env

clean:
    rm -rf bin/*
