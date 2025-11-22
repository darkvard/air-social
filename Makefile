# ===================== DOCKER ======================

# Detect project name from go.mod
MODULE := $(shell grep '^module ' go.mod | sed 's/module //')
PROJECT := $(notdir $(MODULE))

IMAGE ?= $(PROJECT)
CONTAINER ?= $(PROJECT)-dev

HOST_PORT ?= 3000
CONTAINER_PORT ?= 8080


# -------- Utils --------

.PHONY: docker-stop
docker-stop:
	-@docker stop $(CONTAINER)

.PHONY: docker-logs
docker-logs:
	@docker logs -f $(CONTAINER)

.PHONY: docker-sh
docker-sh:
	@docker exec -it $(CONTAINER) sh


# -------- Clean --------

.PHONY: docker-clean-container
docker-clean-container:
	@if docker ps -a --format '{{.Names}}' | grep -Eq '^$(CONTAINER)$$'; then \
		echo "🗑 Removing container: $(CONTAINER)"; \
		docker rm -f $(CONTAINER); \
	fi

.PHONY: docker-clean-image
docker-clean-image:
	@if docker images -q $(IMAGE) | grep -q .; then \
		echo "🗑 Removing image: $(IMAGE)"; \
		docker rmi -f $(IMAGE); \
	fi


# -------- Build --------

.PHONY: docker-build
docker-build: docker-clean-image
	@echo "🔥 Building image: $(IMAGE)"
	@docker build -t $(IMAGE) .


# -------- Run (no mount) --------

.PHONY: docker-run
docker-run: docker-clean-container
	@echo "🚀 Running container (NO-MOUNT): $(CONTAINER)"
	@docker run --name $(CONTAINER) \
		-p $(HOST_PORT):$(CONTAINER_PORT) \
		$(IMAGE)


# -------- Run (mount for Air hot reload) --------

.PHONY: docker-run-mount
docker-run-mount: docker-clean-container
	@echo "🔁 Running container (MOUNT for AIR): $(CONTAINER)"
	@docker run --name $(CONTAINER) \
		-p $(HOST_PORT):$(CONTAINER_PORT) \
		-v $(PWD):/app \
		$(IMAGE)


# -------- Reload (no mount) --------

.PHONY: docker-reload
docker-reload:
	@echo "♻️ Reload: clean → build → run"
	@$(MAKE) docker-clean-container
	@$(MAKE) docker-clean-image
	@$(MAKE) docker-build
	@$(MAKE) docker-run


# -------- Reload (mount + Air) --------
.PHONY: docker-reload-mount
docker-reload-mount:
	@echo "♻️ Reload DEV: clean → build → run (mount)"
	@$(MAKE) docker-clean-container
	@$(MAKE) docker-clean-image
	@$(MAKE) docker-build
	@$(MAKE) docker-run-mount


# ===================== Air ======================
.PHONY: air-build
air-build:
	@go build -o ./tmp/main -buildvcs=false .
