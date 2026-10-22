COMPOSE := docker compose --env-file .env -f infra/docker-compose.yml
PY_SERVICES := auth portfolio analytics risk ai notifications
GO_MODULES := libs/go-common services/gateway services/brokerage-sync

.PHONY: up down smoke ps logs build test test-py test-go lint lint-py lint-go fmt migrate seed

.env:
	cp .env.example .env

up: .env ## Build and start the whole stack, waiting until every service is healthy
	$(COMPOSE) up -d --build --wait

down:
	$(COMPOSE) down

smoke: ## /ready on every service from inside the compose network
	sh scripts/smoke.sh

ps:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs -f --tail=100

build: .env
	$(COMPOSE) build

test: test-py test-go

test-py:
	uv run pytest

test-go:
	@for m in $(GO_MODULES); do (cd $$m && go test -race ./...) || exit 1; done

lint: lint-py lint-go

lint-py:
	uv run ruff check .
	uv run ruff format --check .
	uv run mypy libs/py-common $(addprefix services/,$(PY_SERVICES))

lint-go:
	@for m in $(GO_MODULES); do (cd $$m && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2 run ./...) || exit 1; done

fmt:
	uv run ruff format .
	uv run ruff check --fix .
	@for m in $(GO_MODULES); do (cd $$m && gofmt -w .); done

migrate: ## Runs every service's migrations (Alembic / golang-migrate); first tables arrive in Phase 2
	@for s in $(PY_SERVICES); do \
		if [ -f services/$$s/alembic.ini ]; then $(COMPOSE) run --rm $$s alembic upgrade head || exit 1; fi; \
	done

seed: ## Mock-broker demo data (Phase 3)
	@echo "seed: mock broker arrives in Phase 3"
