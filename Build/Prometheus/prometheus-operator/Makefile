REGISTRY_NAME=$(shell cat ../../REGISTRY)
IMAGE?=prometheus-operator
TEMP_DIR:=$(shell mktemp -d)
ARCH?=amd64
ALL_ARCH=amd64 arm arm64
ML_PLATFORMS=linux/amd64,linux/arm,linux/arm64

VERSION?=latest

ifeq ($(ARCH),amd64)
	BASEIMAGE?=busybox
endif
ifeq ($(ARCH),arm)
	BASEIMAGE?=arm32v7/busybox
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE?=arm64v8/busybox
endif

all: build

install: container container-push deploy
install-dockerhub: deploy-dockerhub

container:
	cp Dockerfile $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i "s|BASEIMAGE|$(BASEIMAGE)|g" Dockerfile
	docker build -t $(REGISTRY_NAME)/$(IMAGE)-$(ARCH):$(VERSION) $(TEMP_DIR)
	rm -r $(TEMP_DIR)

container-push:
	$(MAKE) ARCH=$(ARCH) container
	docker push $(REGISTRY_NAME)/$(IMAGE)-$(ARCH):$(VERSION)

deploy:
	export IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE)-$(ARCH):$(VERSION) 	;\
	export imagePullPolicy=Always						            ;\
	envsubst < prometheus-operator.yaml | kubectl apply -f -		;\
	kubectl apply -f metrics_service.yaml							;\

deploy-dockerhub:
	export IMAGE_TAG="georgzerb/genisys-prometheus-operator:latest" ;\
	export imagePullPolicy=IfNotPresent								;\
	envsubst < prometheus-operator.yaml | kubectl apply -f -		;\
	kubectl apply -f metrics_service.yaml	

delete:
	export IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE)-$(ARCH):$(VERSION)	;\
	envsubst < prometheus-operator.yaml | kubectl delete -f -		;\
	kubectl delete -f metrics_service.yaml							;\

