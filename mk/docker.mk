docker-help:
	@echo ""
	@echo "make up          - Запустить все сервисы"
	@echo "make up-build    - Запустить с пересборкой образов"
	@echo "make down        - Остановить и удалить контейнеры"
	@echo "make clean-up    - Удалить контейнеры и очистить данные БД"
	@echo ""
	@echo "make db-up        - Запустить только Postgres и port-forwarder"
	@echo "make db-down      - Остановить только Postgres и port-forwarder"
	@echo ""

up:
	docker compose up -d

up-build:
	docker compose up -d --build

down:
	docker compose down

db-up:
	docker compose up -d template-db port-forwarder

db-down:
	docker compose stop template-db port-forwarder

clean-up:
	@read -p "Очистить локальные данные БД? [y/N]: " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		docker compose down && rm -rf out/pgdata && echo "Данные очищены"; \
	else \
		echo "Отменено"; \
	fi