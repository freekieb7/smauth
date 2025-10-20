GOPATH := $(shell go env GOPATH)

# Load environment variables from .env file
include .env
export

.PHONY: up
up:
	docker-compose up -d

.PHONY: down
down:
	docker-compose down

.PHONY: air
air:
	go install github.com/air-verse/air@latest
	${GOPATH}/bin/air -c .air.toml

# .PHONY: build-migrator
# build-migrator:
# 	go build -o bin/migrator ./cmd/migrator


# migrate-up: build-migrator
# 	./bin/migrator up

# migrate-down: build-migrator
# 	./bin/migrator down

# migrate-status: build-migrator
# 	./bin/migrator status

migrate-create: build-migrator
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	./bin/migrator create $(name)

migrate-init: build-migrator
	./bin/migrator init

template-generate: 
	go install github.com/a-h/templ/cmd/templ@latest
	${GOPATH}/bin/templ generate

.PROXY: aql-generate
aql-generate:
	docker run --rm -u $(id -u ${USER}):$(id -g ${USER}) --volume `pwd`/internal/openehr/aql:/work antlr/antlr4 -Dlanguage=Go AQL.g4 -o gen -package gen

.PHONY: docker-compose-up
docker-compose-up:
    ${ENV_ARGS} docker-compose -f deploy/docker-compose.yaml up

.PHONY: docker-compose-down
docker-compose-down:
    ${ENV_ARGS} docker-compose -f deploy/docker-compose.yaml down