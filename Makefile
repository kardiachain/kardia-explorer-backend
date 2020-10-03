.EXPORT_ALL_VARIABLES:

COMPOSE_PROJECT_NAME=kardia-kaistarter
DOCKER_FILE=./deployments/godev.Dockerfile

all: build-docker-compose run-db
build-docker-compose:
	docker-compose --file ./deployments/docker-compose.yml build
run-db:
	docker-compose --file ./deployments/docker-compose.yml up -d db
run-clean-db:
	docker-compose --file ./deployments/docker-compose.yml exec db psql -U postgres -c "DROP SCHEMA IF EXISTS public CASCADE"
	docker-compose --file ./deployments/docker-compose.yml exec db psql -U postgres -c "CREATE SCHEMA public"
run-app:
	docker-compose --file ./deployments/docker-compose.yml up app
run-app-test:
	docker-compose --file ./deployments/docker-compose.yml up app_test
run-app-bg:
	docker-compose --file ./deployments/docker-compose.yml up -d app
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
run-app-bg:
	docker-compose --file ./deployments/docker-compose.yml up -d app
