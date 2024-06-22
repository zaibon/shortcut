watch:
    templ generate --watch --proxy=http://localhost:8080 --open-browser=false

dev:
  #!/usr/bin/env -S parallel --shebang --ungroup --jobs {{ num_cpus() }}
  just watch
  air

build-deps:
    gci --help > /dev/null || go install github.com/daixiang0/gci@v0.13.4
    templ --help > /dev/null || go install github.com/a-h/templ/cmd/templ@latest
    sqlc --help > /dev/null || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

fmt:
    gci write --custom-order --skip-generated {{ invocation_directory() }} -s standard -s default -s blank -s dot -s alias -s "prefix(github.com/zaibon/shortcut)" 

lint: fmt
    golangci-lint run ./...

generate:
    go generate ./...

build: generate fmt
    CGO_ENABLED=0 go build -o bin/shortcut cmd/*.go

build-linux: generate fmt
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/shortcut-linux cmd/*.go

build-dev: 
    CGO_ENABLED=0 go build -tags=dev -o bin/shortcut cmd/*.go

run: build
    ./bin/shortcut

db-migrate action: build
    ./bin/shortcut migrate {{ action }}

db-create-migration name type:
    goose -dir=db/migrations create {{ name }} {{ type }}

db-fix:
    goose -dir=db/migrations fix

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
