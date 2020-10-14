.EXPORT_ALL_VARIABLES:

COMPOSE_PROJECT_NAME=kardia-explorer
DOCKER_FILE=.Dockerfile
# Variable for filename for store running processes id
PID_FILE = /tmp/explorer_backend.pid
# We can use such syntax to get main.go and other root Go files.
GO_FILES = $(wildcard *.go)

all: env build
env:
	if test ! -f .env ; \
    then \
         cp .env.sample .env ; \
    fi;
env_dev:
	cp .env ./features/.env ; \
	cp .env ./cmd/api/.env ; \
	cp .env ./cmd/grabber/.env ; \
	cp .env ./server/db/.env ;
build:
	docker-compose build
run-grabber:
	docker-compose up grabber
run-backend:
	docker-compose up backend
utest:
	go test ./... -cover -covermode=count -coverprofile=cover.out -coverpkg=./internal/...
	go tool cover -func=cover.out
list-service:
	docker-compose ps
exec-service:
	docker-compose exec $(service) bash
logs:
	docker-compose logs -f $(service)
destroy:
	docker-compose down
deploy:
	docker-compose up -d