# Define required macros here
SHELL = /bin/sh

.PHONY: build, test
default: build

deploy:
	docker build --platform linux/amd64 -f ./hosting/Dockerfile -t artifact-flow-api:latest .
	docker compose -f ./local-development/docker-compose-stack.yml  --env-file ./local-development/development.env up

build:
	go build ./cmd/server/main.go