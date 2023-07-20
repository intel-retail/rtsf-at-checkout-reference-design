# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

.PHONY: run-portainer run-base run-vap run-full all simulator docker test lint

REPOS=cv-region-of-interest device-scale event-reconciler loss-detector product-lookup
GOREPOS=device-scale event-reconciler loss-detector product-lookup

.PHONY: $(REPOS)

DOCKER_TAG=dev

docker-rm:
	-docker rm $$(docker ps -aq)

clean-docker: docker-rm
	docker volume prune -f && \
	docker network prune -f

run-portainer:
	cd ./loss-detection-app && docker compose -f docker-compose.portainer.yml up -d

run-base:
	cd ./loss-detection-app && \
	docker compose -f docker-compose.edgex.yml up -d && \
	docker compose -f docker-compose.loss-detection.yml up -d

run-vap: models run-base
	cd ./loss-detection-app && \ul
	docker compose -f docker-compose.vap.yml up -d

run-full: run-vap

down:
	cd ./loss-detection-app && \
	docker compose -f docker-compose.vap.yml down && \
	docker compose -f docker-compose.loss-detection.yml down && \
	docker compose -f docker-compose.edgex.yml down

vap-down:
	cd ./loss-detection-app && \
	docker compose -f docker-compose.vap.yml down && \
	docker compose -f docker-compose.loss-detection.yml down && \
	docker compose -f docker-compose.edgex.yml down

models:
	if [ ! -d pipeline-server ] ; then git clone https://github.com/dlstreamer/pipeline-server; fi && \
	cd pipeline-server && \
	git checkout 2022.2.0 && \
	mkdir -p ./loss-detection-app/models && \
	./tools/model_downloader/model_downloader.sh --model-list $(shell pwd)/loss-detection-app/models.yml --output $(shell pwd)/loss-detection-app

clean-deps:
	rm -rf video-analytics-serving

all: simulator docker

simulator:
	cd rtsf-at-checkout-event-simulator; \
	go build -o event-simulator

clean: down clean-deps
	rm -f rtsf-at-checkout-event-simulator/event-simulator && \
	docker rmi $$(docker images | grep rtsf-at-checkout | awk '{print $$3}') && \
	docker volume prune -f && \
    docker network prune -f

docker: 
	for repo in ${REPOS}; do \
		echo rtsf-at-checkout-$$repo; \
		cd rtsf-at-checkout-$$repo; \
		docker build --rm \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		-t rtsf-at-checkout/$$repo:$(DOCKER_TAG) \
		. || exit 1; \
		cd ..; \
	done

go-test: 
	for repo in ${GOREPOS}; do \
		echo rtsf-at-checkout-$$repo; \
		cd rtsf-at-checkout-$$repo; \
		go test -tags no_zmq -coverprofile=coverage.out ./... || exit 1; \
		cd ..; \
	done

go-lint: go-tidy
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run make install-lint"
	@which golangci-lint >/dev/null ;  echo "running golangci-lint"; golangci-lint version; go version; 
	for repo in ${GOREPOS}; do \
		echo rtsf-at-checkout-$$repo; \
		cd rtsf-at-checkout-$$repo; \
		golangci-lint run --config ../.github/.golangci.yml --out-format=line-number >> ../goLintResults.txt ; \
		cd ..; \
	done

go-tidy: 
	for repo in ${GOREPOS}; do \
		echo rtsf-at-checkout-$$repo; \
		cd rtsf-at-checkout-$$repo; \
		go mod tidy || exit 1; \
		cd ..; \
	done

install-go-lint:
	sudo curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b $$(go env GOPATH)/bin v1.51.2

hadolint: 
	docker run --rm -v $(pwd):/repo -i hadolint/hadolint:latest-alpine sh -c "cd /repo && hadolint -f json ./**/Dockerfile" > go-hadolint.json

cv-roi:
	cd rtsf-at-checkout-cv-region-of-interest; \
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f ./Dockerfile \
		-t rtsf-at-checkout/cv-region-of-interest:$(DOCKER_TAG) \
		.

device-scale:
	cd rtsf-at-checkout-device-scale; \
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		-t rtsf-at-checkout/device-scale:$(DOCKER_TAG) \
		.

reconciler:
	cd rtsf-at-checkout-event-reconciler; \
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		-t rtsf-at-checkout/event-reconciler:$(DOCKER_TAG) \
		.

loss-detector:
	cd rtsf-at-checkout-loss-detector; \
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		-t rtsf-at-checkout/loss-detector:$(DOCKER_TAG) \
		.

product-lookup:
	cd rtsf-at-checkout-product-lookup; \
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		-t rtsf-at-checkout/product-lookup:$(DOCKER_TAG) \
		.
