# Copyright SecureKey Technologies Inc.
#
# SPDX-License-Identifier: Apache-2.0

# Namespace for the docker images
DOCKER_OUTPUT_NS   ?= docker.pkg.github.com
DOCKER_IMAGE_NAME ?= trustbloc/hub-router/hub-router
WEBHOOK_IMAGE_NAME ?= trustbloc/hub-router/sample-webhook

# Tool commands (overridable)
ALPINE_VER ?= 3.12
GO_VER ?= 1.15

.PHONY: all
all: checks unit-test bdd-test

.PHONY: checks
checks: license lint

.PHONY: lint
lint:
	@scripts/check_lint.sh

.PHONY: license
license:
	@scripts/check_license.sh

.PHONY: unit-test
unit-test:
	@scripts/check_unit.sh

.PHONY: bdd-test
bdd-test: clean test-keys docker sample-webhook-docker
	@scripts/check_integration.sh

.PHONY: test-keys
test-keys: clean
	@mkdir -p -p test/bdd/fixtures/keys/tls
	@docker run -i --rm \
		-v $(abspath .):/opt/workspace/hub-router \
		--entrypoint "/opt/workspace/hub-router/scripts/generate_test_keys.sh" \
		frapsoft/openssl

.PHONY: docker
docker:
	@echo "Building hub-router docker image"
	@docker build -f ./images/hub-router/Dockerfile --no-cache \
	   -t $(DOCKER_OUTPUT_NS)/$(DOCKER_IMAGE_NAME):latest \
	   --build-arg ALPINE_VER=$(ALPINE_VER) \
	   --build-arg GO_VER=$(GO_VER) .

.PHONY: hub-router
hub-router:
	@echo "Building hub-router"
	@mkdir -p ./.build/bin
	@cd cmd/hub-router && go build -o ../../.build/bin/hub-router main.go

.PHONY: sample-webhook
sample-webhook:
	@echo "Building sample webhook server"
	@mkdir -p ./build/bin
	@go build -o ./build/bin/webhook-server test/bdd/cmd/webhook/main.go

.PHONY: sample-webhook-docker
sample-webhook-docker:
	@echo "Building sample webhook server docker image"
	@docker build -f ./images/mocks/webhook/Dockerfile --no-cache -t $(DOCKER_OUTPUT_NS)/$(WEBHOOK_IMAGE_NAME):latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) \
	--build-arg GOPROXY=$(GOPROXY) .

.PHONY: clean
clean:
	@rm -Rf ./.build
	@rm -Rf ./test/bdd/fixtures/keys/tls
	@rm -Rf ./test/bdd/docker-compose.log


