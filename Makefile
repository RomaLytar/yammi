.PHONY: up down restart test clean

# ============================================================
#  make up     — поднять всё (бек, фронт, графана, сокеты)
#  make down   — остановить всё
#  make restart — перезапустить всё
# ============================================================

up:
	@BUILDKIT_PROGRESS=quiet docker compose --progress quiet up --build -d
	@echo ""
	@echo "=== Yammi started ==="
	@docker compose ps --format "table {{.Name}}\t{{.Status}}"

down:
	docker compose down

restart: down up

# ============================================================
#  make test FILE=realistic_1000_users.js [LIMIT=300000] [CLEAN=1] [REDIS_PASS=your_pass]
#
#  FILE       — (обязательно) имя файла из tests/load/
#  LIMIT      — (опционально) rate limit для всех эндпоинтов на время теста
#  CLEAN      — (опционально) 1 = запустить cleanup.sh после теста
#  REDIS_PASS — (опционально) пароль Redis для cleanup
# ============================================================

FILE ?=
LIMIT ?=
CLEAN ?= 0
REDIS_PASS ?= yammi_redis

test:
ifndef FILE
	$(error FILE обязателен. Пример: make test FILE=realistic_1000_users.js)
endif
ifneq ($(LIMIT),)
	@echo "=== Rate limit → $(LIMIT) ==="
	@RATE_LIMIT_REGISTER=$(LIMIT) RATE_LIMIT_LOGIN=$(LIMIT) RATE_LIMIT_REFRESH=$(LIMIT) RATE_LIMIT_DEFAULT=$(LIMIT) \
		docker compose --progress quiet up -d api-gateway
endif
	docker run --rm --network=host -v $(PWD)/tests/load:/scripts grafana/k6 run /scripts/$(FILE)
ifneq ($(LIMIT),)
	@echo "=== Rate limit → defaults ==="
	@docker compose --progress quiet up -d api-gateway
endif
ifeq ($(CLEAN),1)
	REDIS_PASSWORD=$(REDIS_PASS) bash tests/load/cleanup.sh
endif

# ============================================================
#  make clean [REDIS_PASS=your_pass]
# ============================================================

clean:
	REDIS_PASSWORD=$(REDIS_PASS) bash tests/load/cleanup.sh
