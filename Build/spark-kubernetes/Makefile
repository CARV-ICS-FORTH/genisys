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
NAMESPACE='default'
REGISTRY_NAME=$(shell cat ../REGISTRY)
NFS_PATH=$(shell cat ../NFS_PATH)

CENTOS=$(shell cat ../CENTOS)
SPARK_VERSION=latest
SPARK_IMAGE_TAG=$(REGISTRY_NAME)/spark:$(SPARK_VERSION)


.PHONY: container container-push install-template

install: container container-push deploy
install-dockerhub: deploy-dockerhub

container:
	cd docker && docker build -f Dockerfile --network=host --build-arg CENTOS=$(CENTOS) --build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) -t $(SPARK_IMAGE_TAG) .

container-push:
	docker push $(SPARK_IMAGE_TAG)

deploy:
	export IMAGE_TAG=${SPARK_IMAGE_TAG}															;\
	export NFS_PATH=${NFS_PATH}																	;\
	export imagePullPolicy=Always																;\
	envsubst < ./kubernetes/spark-master-deployment.yaml | kubectl apply -n $(NAMESPACE) -f - 	;\
	envsubst < ./kubernetes/spark-master-service.yaml | kubectl apply -n $(NAMESPACE) -f -		;\
	envsubst < ./kubernetes/spark-worker-deployment.yaml | kubectl apply -n $(NAMESPACE) -f - 	;\
	envsubst < ./kubernetes/minikube-ingress.yaml | kubectl apply -n $(NAMESPACE) -f - 			;\

deploy-dockerhub:
	export IMAGE_TAG="georgzerb/genisys-spark:latest"											;\
	export NFS_PATH=${NFS_PATH}																	;\
	export imagePullPolicy=IfNotPresent															;\
	envsubst < ./kubernetes/spark-master-deployment.yaml | kubectl apply -n $(NAMESPACE) -f - 	;\
	envsubst < ./kubernetes/spark-master-service.yaml | kubectl apply -n $(NAMESPACE) -f -		;\
	envsubst < ./kubernetes/spark-worker-deployment.yaml | kubectl apply -n $(NAMESPACE) -f - 	;\
	envsubst < ./kubernetes/minikube-ingress.yaml | kubectl apply -n $(NAMESPACE) -f - 			;\

delete:
	export IMAGE_TAG=${SPARK_IMAGE_TAG}															;\
	export NFS_PATH=${NFS_PATH}																	;\
	envsubst < ./kubernetes/spark-master-deployment.yaml | kubectl delete -n $(NAMESPACE) -f - 	;\
	envsubst < ./kubernetes/spark-master-service.yaml | kubectl delete -n $(NAMESPACE) -f -		;\
	envsubst < ./kubernetes/spark-worker-deployment.yaml | kubectl delete -n $(NAMESPACE) -f - 	;\
	envsubst < ./kubernetes/minikube-ingress.yaml | kubectl delete -n $(NAMESPACE) -f - 		;\