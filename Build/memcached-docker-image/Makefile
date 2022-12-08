# Copyright [2020] [FORTH-ICS]
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

KUBECTL_VERSION?=v1.19.8

REGISTRY_NAME=$(shell cat ../REGISTRY)

CENTOS=$(shell cat ../CENTOS)



MEMCACHED_VERSION=latest
MEMCACHED_IMAGE_TAG=$(REGISTRY_NAME)/memcached:$(MEMCACHED_VERSION)

.PHONY: container container-push install-template


install: container container-push deploy
install-dockerhub: deploy-dockerhub

container:
	chmod -R 777 *
	docker build -f Dockerfile --network=host --build-arg CENTOS=$(CENTOS) --build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) -t $(MEMCACHED_IMAGE_TAG) .

container-push:
	docker push $(MEMCACHED_IMAGE_TAG)

deploy:
	export MEMCACHED_IMAGE_TAG=${MEMCACHED_IMAGE_TAG}   ;\
	export imagePullPolicy=Always						;\
	envsubst < memcached.yaml | kubectl apply -f -

deploy-dockerhub:
	export MEMCACHED_IMAGE_TAG="georgzerb/genisys-memcached:latest" 	;\
	export imagePullPolicy=IfNotPresent							;\
	envsubst < memcached.yaml | kubectl apply -f -

delete:
	export MEMCACHED_IMAGE_TAG=${MEMCACHED_IMAGE_TAG}; \
	envsubst < memcached.yaml | kubectl delete -f -