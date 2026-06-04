# BiteSense build & deploy targets.
#
# Two flows are supported:
#
# 1. Pi-Homelab (default): build images on the Pi via BuildKit (`pibuild`),
#    rsync code + Flutter web bundle from the notebook beforehand. No registry.
# 2. External registry (GHCR): use `make image-build-portable` to produce
#    standard docker images via `docker buildx`, then push wherever you want.
#
# Configurable variables (override on the command line, e.g. `make deploy TAG=$(git rev-parse --short HEAD)`):
#   PI_USER     SSH user on the Pi-Homelab        (default: marvin)
#   PI_HOST     SSH host of the Pi-Homelab        (default: pi-homelab.fritz.box)
#   PI_PATH     Working dir on the Pi             (default: /home/marvin/bitesense)
#   TAG         Image tag                          (default: short git SHA)
#   API_URL     Flutter API base URL              (default: https://bitesense.tailb969ce.ts.net)
#   KUBE_OVERLAY  k8s overlay to apply             (default: prod)

PI_USER       ?= marvin
PI_HOST       ?= pi-homelab.fritz.box
PI_PATH       ?= /home/marvin/bitesense
TAG           ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo local)
API_URL       ?= https://bitesense.tailb969ce.ts.net
KUBE_OVERLAY  ?= prod
SSH           ?= ssh
RSYNC         ?= rsync -az --delete

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-22s %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Local dev
# ---------------------------------------------------------------------------

.PHONY: backend-run
backend-run: ## Run backend locally (reads backend/.env).
	cd backend && go run ./cmd/server

.PHONY: backend-test
backend-test: ## Run backend tests.
	cd backend && go test -race ./...

.PHONY: backend-lint
backend-lint: ## Run golangci-lint.
	cd backend && golangci-lint run

.PHONY: frontend-pub-get
frontend-pub-get: ## Resolve Flutter dependencies (also generates ARB bindings).
	cd frontend && flutter pub get

.PHONY: frontend-test
frontend-test: ## Run Flutter tests.
	cd frontend && flutter test

.PHONY: frontend-build-web
frontend-build-web: frontend-pub-get ## Build Flutter Web release into frontend/build/web/.
	cd frontend && flutter build web --release --base-href / \
		--dart-define=BITESENSE_API=$(API_URL)

# ---------------------------------------------------------------------------
# Pi-Homelab deployment (BuildKit / pibuild)
# ---------------------------------------------------------------------------

.PHONY: pi-sync
pi-sync: frontend-build-web ## Rsync backend + frontend (incl. build/web/) + k8s to the Pi.
	$(SSH) $(PI_USER)@$(PI_HOST) "mkdir -p $(PI_PATH)/backend $(PI_PATH)/frontend $(PI_PATH)/k8s"
	$(RSYNC) backend/   $(PI_USER)@$(PI_HOST):$(PI_PATH)/backend/  --exclude=.env --exclude=data/
	# Frontend: ship sources + build/web/, but skip the Dart tool cache.
	$(RSYNC) frontend/  $(PI_USER)@$(PI_HOST):$(PI_PATH)/frontend/ \
		--exclude=.dart_tool/ \
		--exclude=android/ \
		--exclude=ios/ \
		--exclude=.idea/
	$(RSYNC) k8s/       $(PI_USER)@$(PI_HOST):$(PI_PATH)/k8s/

.PHONY: pi-build-backend
pi-build-backend: ## Build backend image on the Pi via pibuild.
	$(SSH) $(PI_USER)@$(PI_HOST) "pibuild $(PI_PATH)/backend bitesense-backend:$(TAG)"

.PHONY: pi-build-frontend
pi-build-frontend: ## Build frontend image on the Pi via pibuild.
	$(SSH) $(PI_USER)@$(PI_HOST) "pibuild $(PI_PATH)/frontend bitesense-frontend:$(TAG)"

.PHONY: pi-build
pi-build: pi-sync pi-build-backend pi-build-frontend ## Sync + build both images.
	@echo "Built bitesense-backend:$(TAG) + bitesense-frontend:$(TAG) on $(PI_HOST)"

.PHONY: pi-apply
pi-apply: ## Set image tags in the overlay and apply on the Pi.
	$(SSH) $(PI_USER)@$(PI_HOST) "cd $(PI_PATH)/k8s/overlays/$(KUBE_OVERLAY) && \
		kustomize edit set image bitesense-backend=bitesense-backend:$(TAG) && \
		kustomize edit set image bitesense-frontend=bitesense-frontend:$(TAG) && \
		kubectl apply -k . && \
		kubectl -n bitesense rollout status deploy/bitesense-backend && \
		kubectl -n bitesense rollout status deploy/bitesense-frontend"

.PHONY: pi-deploy
pi-deploy: pi-build pi-apply ## Full Pi deploy: sync, build, apply.
	@echo "https://bitesense.tailb969ce.ts.net"

# ---------------------------------------------------------------------------
# Portable / external registry build (GHCR)
# ---------------------------------------------------------------------------

REGISTRY ?= ghcr.io/marvinf95

.PHONY: image-build-portable
image-build-portable: ## Build multi-arch images via docker buildx (push to GHCR).
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t $(REGISTRY)/bitesense-backend:$(TAG) --push ./backend
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f frontend/Dockerfile.fullbuild \
		-t $(REGISTRY)/bitesense-frontend:$(TAG) --push ./frontend
