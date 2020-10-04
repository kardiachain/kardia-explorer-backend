.EXPORT_ALL_VARIABLES:

COMPOSE_PROJECT_NAME=kardia-explorer
DOCKER_FILE=./deployments/godev.Dockerfile
# Variable for filename for store running processes id
PID_FILE = /tmp/my-app.pid
# We can use such syntax to get main.go and other root Go files.
GO_FILES = $(wildcard *.go)

all: build
build:
	docker-compose --file ./deployments/docker-compose.yml build
run-db:
	docker-compose --file ./deployments/docker-compose.yml up -d db
run-clean-db:
	docker-compose --file ./deployments/docker-compose.yml exec db psql -U postgres -c "DROP SCHEMA IF EXISTS public CASCADE"
	docker-compose --file ./deployments/docker-compose.yml exec db psql -U postgres -c "CREATE SCHEMA public"
serve-backend:
	docker-compose --file ./deployments/docker-compose.yml up backend
serve-grabber:
	docker-compose --file ./deployments/docker-compose.yml up grabber
run-grabber:
	docker-compose --file ./deployments/docker-compose.yml up grabber
run-backend:
	docker-compose --file ./deployments/docker-compose.yml up backend
utest:
	go test ./internal/... -cover -covermode=count -coverprofile=cover.out -coverpkg=./internal/...
	go tool cover -func=cover.out
itest:
	docker-compose --file ./deployments/docker-compose.yml exec app sh -c "cd ./features && godog ."
list-service:
	docker-compose --file ./deployments/docker-compose.yml ps
exec-service:
	docker-compose --file ./deployments/docker-compose.yml exec $(service) bash
logs:
	docker-compose --file ./deployments/docker-compose.yml logs -f $(service)
destroy:
	docker-compose --file ./deployments/docker-compose.yml down
run-backend-bg:
	docker-compose --file ./deployments/docker-compose.yml up -d backend
