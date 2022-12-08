# Copyright [2021] [FORTH-ICS]
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
REGISTRY_NAME=$(shell cat ../Build/REGISTRY)

genisys_VERSION=$(shell cat VERSION)
genisys_IMAGE_TAG=$(REGISTRY_NAME)/genisys:$(genisys_VERSION)

.PHONY: container container-push local

install: container container-push deploy
install-dockerhub: deploy-dockerhub

container:
	docker build --build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) -t $(genisys_IMAGE_TAG) .

container-push:
	docker push $(genisys_IMAGE_TAG)

deploy:
	export genisys_IMAGE_TAG=${genisys_IMAGE_TAG} 		;\
	export imagePullPolicy=Always						;\
	envsubst < genisys.yaml | kubectl apply -f - 		;\
	envsubst < permissions.yaml | kubectl apply -f -	;\
	envsubst < components.yaml | kubectl apply -f -		;\

deploy-dockerhub:
	export genisys_IMAGE_TAG="georgzerb/genisys:latest" ;\
	export imagePullPolicy=IfNotPresent					;\
	envsubst < genisys.yaml | kubectl apply -f - 		;\
	envsubst < permissions.yaml | kubectl apply -f -	;\
	envsubst < components.yaml | kubectl apply -f -		;\
	
delete:
	export genisys_IMAGE_TAG=${genisys_IMAGE_TAG}		 ;\
	envsubst < genisys.yaml | kubectl delete -f - 		 ;\
	envsubst < permissions.yaml | kubectl delete -f - 	 ;\
	envsubst < components.yaml | kubectl delete -f -	 ;\

local:
	(cd scheduler && go build)
