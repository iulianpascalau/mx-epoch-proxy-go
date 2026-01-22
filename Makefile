SHELL := $(shell which bash)

.DEFAULT_GOAL := help

.PHONY: clean-test test build run-backend run-frontend run-solution

help:
	@echo -e ""
	@echo -e "Make commands:"
	@grep -E '^[a-zA-Z_-]+:.*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":"}; {printf "\t\033[36m%-30s\033[0m\n", $$1}'
	@echo -e ""

# #########################
# Base commands
# #########################

binary = epoch-proxy-server

clean-tests:
	go clean -testcache

tests: clean-tests
	go test ./...

slow-tests: clean-tests
	@docker compose -f docker/docker-compose.yml build
	@docker compose -f docker/docker-compose.yml up -d
	@go test ./integrationTests/... -v -timeout 40m -tags slow
	@docker compose -f docker/docker-compose.yml down -v

build:
	cd ./services/proxy && go build -v \
	-o ${binary} \
	-ldflags="-X main.appVersion=$(shell git describe --tags --long --dirty) -X main.commitID=$(shell git rev-parse HEAD)"

run-backend: build
	cd ./services/proxy && \
		./${binary} --log-level="*:DEBUG"

run-frontend:
	cd frontend && \
	export NVM_DIR="$$HOME/.nvm" && \
	[ -s "$$NVM_DIR/nvm.sh" ] && \. "$$NVM_DIR/nvm.sh" && \
	nvm exec 22.12.0 npm run dev

run-solution:
	$(MAKE) -j2 run-backend run-frontend

binary-crypto-payment = crypto-payment-server

build-crypto-payment:
	cd ./services/crypto-payment && go build -v \
	-o ${binary-crypto-payment} \
	-ldflags="-X main.appVersion=$(shell git describe --tags --long --dirty) -X main.commitID=$(shell git rev-parse HEAD)"

run-crypto-payment: build-crypto-payment
	cd ./services/crypto-payment && \
		./${binary-crypto-payment} --log-level="*:DEBUG"

lint-install:
ifeq (,$(wildcard test -f bin/golangci-lint))
	@echo "Installing golint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s
endif

run-lint:
	@echo "Running golint"
	bin/golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 --timeout=2m
