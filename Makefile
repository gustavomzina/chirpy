include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the application
.PHONY: run/api
run/api:
	go run ./cmd/chirpy

## db/generate: generate models using sqlc
.PHONY: run/api
db/generate:
	sqlc generate

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${CHIRPY_DB_DSN}

## db/migrations/status: check migration status
.PHONY: db/migrations/check
db/migrations/check: 
	goose -dir sql/schema postgres ${CHIRPY_DB_DSN} status

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	goose -dir sql/schema create ${name} sql

## db/migrations/up-by-one: apply one up database migration
.PHONY: db/migrations/up-by-one
db/migrations/up-by-one: confirm
	@echo 'Running up by one migration...'
	goose -dir sql/schema postgres ${CHIRPY_DB_DSN} up-by-one

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	goose -dir sql/schema postgres ${CHIRPY_DB_DSN} up

## db/migrations/up: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migration...'
	goose -dir sql/schema postgres ${CHIRPY_DB_DSN} down

