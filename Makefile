IMAGE_URL = lititacr.azurecr.io/notification-handler
TAG = latest

.PHONE: migrations
migrations:
	@go run cmd/migrate/main.go
	@echo Makefile: $@ target finished

.PHONE: topics
topics:
	@go run cmd/kafka/main.go
	@echo Makefile: $@ target finished

.PHONE: setup
setup:
	@docker compose -f docker-compose.dev.yml up -d
	@echo Makefile: $@ target finished

.PHONY: run
run: migrations topics
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

.PHONY: docker-login
docker-login:
	@az acr login --name lititacr
	@echo Makefile: $@ target finished

.PHONY: docker-build
docker-build:
	@docker build --platform linux/amd64 -t $(IMAGE_URL):$(TAG) .
	@echo Makefile: $@ target finished

.PHONY: docker-push
docker-push:
	@docker push $(IMAGE_URL):$(TAG)
	@echo Makefile: $@ target finished

.PHONY: docker-down
docker-down:
	@docker compose down
	@echo Makefile: $@ target finished
