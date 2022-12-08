REGISTRY_NAME=$(shell cat ../../REGISTRY)
IMAGE?=prometheus
TEMP_DIR:=$(shell mktemp -d)
ARCH?=amd64
ALL_ARCH=amd64 arm arm64
ML_PLATFORMS=linux/amd64,linux/arm,linux/arm64

IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE)-$(ARCH)

PROMARCH=$(ARCH)
ifeq ($(ARCH),amd64)
	BASEIMAGE=busybox
endif
ifeq ($(ARCH),arm)
	PROMARCH=armv7
	BASEIMAGE=arm32v7/busybox
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE=arm64v8/busybox
endif

VERSION_SEMVER=2.2.1
VERSION?=v$(VERSION_SEMVER)
URL?=https://github.com/prometheus/prometheus/releases/download/$(VERSION)/prometheus-$(VERSION_SEMVER).linux-$(PROMARCH).tar.gz

all: build


install: container container-push deploy
install-dockerhub: deploy-dockerhub

container:
	curl -sSL $(URL) | tar -xz -C $(TEMP_DIR) --strip-component=1
	cp Dockerfile $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i "s|BASEIMAGE|$(BASEIMAGE)|g" Dockerfile

	docker build -t $(IMAGE_TAG):$(VERSION) $(TEMP_DIR)
	rm -r $(TEMP_DIR)

push-%:
	$(MAKE) ARCH=$* container
	docker push $(REGISTRY_NAME)/$(IMAGE)-$*:$(VERSION)

container-push: ./manifest-tool $(addprefix push-,$(ARCH))
	./manifest-tool push from-args --platforms $(ML_PLATFORMS) --template $(IMAGE_TAG):$(VERSION) --target $(REGISTRY_NAME)/$(IMAGE):$(VERSION)

deploy:
	export IMAGE_TAG=$(IMAGE_TAG) 										;\
	envsubst < sample-prometheus-instance.yaml | kubectl apply -f -

deploy-dockerhub:
	export IMAGE_TAG="georgzerb/genisys-prometheus"						;\
	envsubst < sample-prometheus-instance.yaml | kubectl apply -f -

delete:
	export IMAGE_TAG=$(IMAGE_TAG)										;\
	envsubst < sample-prometheus-instance.yaml | kubectl delete -f -

./manifest-tool:
	curl -sSL https://github.com/estesp/manifest-tool/releases/download/v0.7.0/manifest-tool-linux-amd64 > manifest-tool
	chmod +x manifest-tool
