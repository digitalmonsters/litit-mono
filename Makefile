IMAGE_URL := "lititacr.azurecr.io/configurator:v2"

export ENVIRONMENT=dev

.PHONE: migrations
migrations:
	@go run cmd/migrate/main.go
	@echo Makefile: $@ target finished

.PHONE: setup
setup:
	@docker compose -f docker-compose.dev.yml up -d
	@echo Makefile: $@ target finished

.PHONY: run
run: migrations
	@go run cmd/main.go
	@echo Makefile: $@ target finished

.PHONY: build
build:
	@go build -o bin/configurator cmd/main.go
	@echo Makefile: $@ target finished

.PHONY: docker-run
docker-run:
	@docker compose up -d
	@echo Makefile: $@ target finished
