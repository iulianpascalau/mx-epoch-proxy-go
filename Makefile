SHELL := $(shell which bash)

.DEFAULT_GOAL := help

.PHONY: clean-test test build run

help:
	@echo -e ""
	@echo -e "Make commands:"
	@grep -E '^[a-zA-Z_-]+:.*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":"}; {printf "\t\033[36m%-30s\033[0m\n", $$1}'
	@echo -e ""

# #########################
# Base commands
# #########################

cmd_dir = cmd/proxy
binary = proxy

clean-tests:
	go clean -testcache

tests: clean-tests
	go test ./...

build:
	cd ${cmd_dir} && \
		go build -v \
		-o ${binary} \
		-ldflags="-X main.appVersion=$(shell git describe --tags --long --dirty) -X main.commitID=$(shell git rev-parse HEAD)"

run: build
	cd ${cmd_dir} && \
		./${binary} --log-level="*:DEBUG"

redis-tests: clean-tests
	@docker compose -f docker-compose.yml build
	@docker compose -f docker-compose.yml up & go test ./storage/... -v -timeout 20m -tags redis
	@docker compose -f docker-compose.yml down -v
