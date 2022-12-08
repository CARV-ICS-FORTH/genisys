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
NFS_PATH=$(shell cat ../NFS_PATH)
KERNEL=$(shell cat ../KERNEL)
OFED=$(shell cat ../OFED)
CENTOS=$(shell cat ../CENTOS)
REPO=$(shell cat ../REPO)


MPI_VERSION=latest
MPI_IMAGE_TAG=$(REGISTRY_NAME)/genisys-slurm-infiniband:$(MPI_VERSION)

.PHONY: container container-push install-template

install: container container-push deploy
install-dockerhub: deploy-dockerhub

container:
	chmod -R 777 *
	docker build -f Dockerfile --network=host --build-arg CENTOS=$(CENTOS) --build-arg REPO=$(REPO) --build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) --build-arg KERNEL=$(KERNEL) --build-arg OFED=$(OFED)  -t $(MPI_IMAGE_TAG) .

container-push:
	docker push $(MPI_IMAGE_TAG)

deploy:
		export IMAGE_TAG=${MPI_IMAGE_TAG}						;\
		export NFS_PATH=${NFS_PATH}								;\
		export imagePullPolicy=Always							;\
		envsubst < permissions.yaml  | kubectl apply -f - 		;\
		envsubst < mpi-slurm.deploy.yaml  | kubectl apply -f - 	;\

deploy-dockerhub:
		export IMAGE_TAG="georgzerb/virtual-cluster-infiniband:latest"	;\
		export NFS_PATH=${NFS_PATH}										;\
		export imagePullPolicy=IfNotPresent								;\
		envsubst < permissions.yaml  | kubectl apply -f - 				;\
		envsubst < mpi-slurm.deploy.yaml  | kubectl apply -f - 			;\

delete:
		export IMAGE_TAG=${MPI_IMAGE_TAG}								;\
		export NFS_PATH=${NFS_PATH}										;\
		envsubst < permissions.yaml  | kubectl delete -f - 				;\
		envsubst < mpi-slurm.deploy.yaml  | kubectl delete -f - 		;\

