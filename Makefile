IMAGE_URL := "lititacr.azurecr.io/configurator"

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

.PHONE: verify-tag
verify-tag:
ifndef tag
	$(error tag is undefined)
endif

.PHONY: docker-image
docker-image: verify-tag
ifndef github_token
	$(error github_token is undefined)
endif
	@docker build --build-arg GITHUB_TOKEN="$(github_token)" --platform linux/amd64 -t $(IMAGE_URL):$(tag) .
	@echo Makefile: $@ target finished

.PHONY: docker-push
docker-push: verify-tag
	@docker push $(IMAGE_URL):$(tag)
	@echo Makefile: $@ target finished
