.PHONY: build
build:
	docker-compose build

.PHONY: run
run:
	docker-compose up

.PHONY: swag
swag:
	swag fmt
	swag init -g cmd/apiserver/main.go

.DEFAULT_GOAL := build