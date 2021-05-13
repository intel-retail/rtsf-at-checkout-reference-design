# Copyright © 2019 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

.PHONY: run-portainer run-base run-vap run-full all simulator docker

DOCKERS=cv-roi device-scale reconciler loss-detector product-lookup rsp-event-handler

.PHONY: $(DOCKERS)

DOCKER_TAG=dev

docker-rm:
	-docker rm $$(docker ps -aq)

clean-docker: docker-rm
	docker volume prune -f && \
	docker network prune -f

run-portainer:
	cd ./loss-detection-app && docker-compose -f docker-compose.portainer.yml up -d

run-base:
	cd ./loss-detection-app && \
	docker-compose -f docker-compose.edgex.yml up -d && \
	docker-compose -f docker-compose.loss-detection.yml up -d

run-vap:
	cd ./loss-detection-app && \
	docker-compose -f docker-compose.vap.yml up -d

run-rsp:
	cd ./loss-detection-app && \
	docker-compose -f docker-compose.rsp.yml up -d

run-full: run-base run-vap run-rsp

down:
	cd ./loss-detection-app && \
	docker-compose -f docker-compose.vap.yml down && \
	docker-compose -f docker-compose.rsp.yml down && \
	docker-compose -f docker-compose.loss-detection.yml down && \
	docker-compose -f docker-compose.edgex.yml down

vas-down:
	cd ./loss-detection-app && \
	docker-compose -f docker-compose.vap.yml down

VAS_VERSION=v0.5.0-beta
vas:
	git clone https://github.com/intel/video-analytics-serving && \
	cd video-analytics-serving/docker && \
	git checkout ${VAS_VERSION} && \
	./build.sh 

rsp:
	git clone https://github.com/intel/rsp-sw-toolkit-im-suite-mqtt-device-service && \
	cd rsp-sw-toolkit-im-suite-mqtt-device-service && \
    docker build \
    	 --build-arg http_proxy \
    	 --build-arg https_proxy \
    	 -t rsp/mqtt-device-service:dev \
    	 .;
	git clone https://github.com/intel/rsp-sw-toolkit-installer && \
	cd rsp-sw-toolkit-installer/docker && \
	./build.sh

clean-deps:
	rm -rf rsp-sw-toolkit-im-suite-mqtt-device-service
	rm -rf rsp-sw-toolkit-installer
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

docker: $(DOCKERS)

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

rsp-event-handler:
	cd rtsf-at-checkout-rsp-controller-event-handler; \
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		-t rtsf-at-checkout/rsp-controller-event-handler:$(DOCKER_TAG) \
		.
