include .env
export

export PROJECT_ROOT=${shell pwd}

help:
	@echo "make docker-help - Узнать доступные команды запуска проекта"
	@echo "make migrate-help - Узнать доступные команды миграций"

include mk/docker.mk
include mk/migrate.mk