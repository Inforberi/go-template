docker-help:
	@echo ""
	@echo "make up          - Запустить все сервисы"
	@echo "make up-build    - Запустить с пересборкой образов"
	@echo "make down        - Остановить и удалить контейнеры"
	@echo "make clean-up    - Удалить контейнеры и очистить данные БД"
	@echo ""

up:
	docker compose up -d

up-build:
	docker compose up -d --build

down:
	docker compose down

clean-up:
	@read -p "Очистить локальные данные БД? [y/N]: " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		docker compose down && rm -rf out/pgdata && echo "Данные очищены"; \
	else \
		echo "Отменено"; \
	fi