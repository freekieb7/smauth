up:
	docker-compose up -d

down:
	docker-compose down

air:
	~/go/bin/air

templ:
	~/go/bin/templ generate

build-migrator:
	go build -o bin/migrator ./cmd/migrator

migrate-up: build-migrator
	./bin/migrator up

migrate-down: build-migrator
	./bin/migrator down

migrate-status: build-migrator
	./bin/migrator status

migrate-create: build-migrator
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	./bin/migrator create $(name)

migrate-init: build-migrator
	./bin/migrator init

clean:
	rm -f bin/migrator

help:
	@echo "Available commands:"
	@echo "  run                Run the application with docker-compose"
	@echo "  build-migrator     Build the migration CLI tool"
	@echo "  migrate-create     Create a new migration (requires name=<migration_name>)"
	@echo "  migrate-up         Run all pending migrations"
	@echo "  migrate-down       Rollback the most recent migration"
	@echo "  migrate-status     Show migration status"
	@echo "  migrate-init       Initialize the migration tracking table"
	@echo "  clean              Remove built binaries"
	@echo "  help               Show this help message"