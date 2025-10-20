GOPATH := $(shell go env GOPATH)

# Load environment variables from .env file
include .env
export

.PHONY: up
up:
	docker compose -f deploy/docker-compose.yaml up --detach --wait --build

.PHONY: down
down:
	docker compose -f deploy/docker-compose.yaml down

.PHONY: dev
dev:
	docker compose -f deploy/docker-compose.yaml up --detach --wait db alloy
	${MAKE} migrate-up
	${MAKE} air

.PHONY: air
air:
	go install github.com/air-verse/air@latest
	${GOPATH}/bin/air -c .air.toml

.PHONY: build
build:
	go build -o ./bin/smauth ./cmd/main.go

.PHONY: migrate-create
migrate-create: build
	./bin/smauth migrate create "$(name)"

.PHONY: migrate-up
migrate-up: build
	./bin/smauth migrate up

.PHONY: migrate-down
migrate-down: build
	./bin/smauth migrate down

.PHONY: migrate-status
migrate-status: build
	./bin/smauth migrate status

.PHONY: template-generate
template-generate: 
	go install github.com/a-h/templ/cmd/templ@latest
	${GOPATH}/bin/templ generate

.PHONY: aql-generate
aql-generate:
	docker run --rm -u $(id -u ${USER}):$(id -g ${USER}) --volume `pwd`/internal/openehr/aql:/work antlr/antlr4 -Dlanguage=Go AQL.g4 -o gen -package gen