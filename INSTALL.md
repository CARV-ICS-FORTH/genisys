# Prerequisites (IMPORTANT)
    
```
OS: Centos 7 (Full Installation Guide) or any other Linux Distro if Kubernetes and NFS are present
    
Docker: 20.10+

Kubernetes: 1.18 -> 1.21

Docker Registry        (For multi-node setup)

NFS                    (For multi-node setup)
```

# Gensisys + Virtual Clusters Deployment

<details>

# Initialize Setup Parameters

<details>
<summary>Details</summary> 

## Initialize NFS Path (MANDATORY for multi-node setups)
In order to share executables and datasets across all the Virtual Cluster's Nodes we have to initialize the NFS_PATH file 
with the path of the NFS filesystem. 

If single node Kubernetes is available then use a system path.

```
sudo nano Build/NFS_PATH
```

eg.
```
/var/nfs_share_dir/
```


## Initialize Docker Registry IP (OPTIONAL for Local Builds ONLY)
If you wish to modify Genisys or Virtual Cluster Dockerfiles and build your own modified images then a Docker Registry is needed.
In order to pull and push Docker Images we have to initialize Docker Registry IP in REGISTRY file with IP of the Docker Registry we installed earlier.

```
sudo nano Build/REGISTRY
```

eg.
```
REGISTRY_IP:5000
```



## Initialize Kernel Version (OPTIONAL Infiniband Support) 

The Infinband Virtual Cluster Docker image supplied on Dockerhub may not be compatible with some Infiniband setups, if this is the case you have to build the Docker Image from scratch.
   
In order for the Virtual Cluster Nodes to support Infiniband we have to initialize the KERNEL file 
with the version of the Kernel installed on the worker nodes.

```
sudo nano Build/KERNEL
```

eg.
```
3.10.0-1160.45.1
```



## Initialize Centos Version (OPTIONAL Infiniband Support) 
    
The Infinband Virtual Cluster Docker image supplied on Dockerhub may not be compatible with some Infiniband setups, if this is the case you have to build the Docker Image from scratch.  
    
Initialize Centos Version for the Virtual Cluster Docker Image base.

```
sudo nano Build/CENTOS
```

eg.
```
7.9.2009
```

## Initialize Mellanox OFED Version (OPTIONAL Infiniband Support) 
    
The Infinband Virtual Cluster Docker image supplied on Dockerhub may not be compatible with some Infiniband setups, if this is the case you have to build the Docker Image from scratch.      
    
Initialize Mellanox OFED Version for the Virtual Cluster Nodes Infiniband support. Edit the OFED file with the OFED version installed on your system.

```
ofed_info
sudo nano Build/OFED
```

eg.
```
4.9-2.2.4.0
```

## Initialize REPO (OPTIONAL Infiniband Support) 

The Infinband Virtual Cluster Docker image supplied on Dockerhub may not be compatible with some Infiniband setups, if this is the case you have to build the Docker Image from scratch.      
    
Initialize REPO that the Dockerfile is going to use to fetch the Kernel headers of your system. Edit the REPO file with the REPO that is going to be used.

```
sudo nano Build/REPO
```

eg.
```
http://mirror.centos.org/centos/7/updates/x86_64/Packages
```
</details>    


# Genisys Scheduler Deployemt

In this section we are going to configure and deploy Genisys.

<details>
<summary>Details</summary> 

## Deploy Genisys 

Run:
```
(cd Genisys/ && make install-dockerhub)
```

## If you wish to build your own modified image with different scheduling parameters

To configure Genisys modify the Genisys initialization command to the "Genisys/genisys_init_command.sh" script.

Genisys Configuration options:
```
    # Usage of genisys command line options:

    #    -Allow_MPI_Colocation string
    #            -Allow_MPI_Colocation 1 (0 disbale MPI JOB Colocation) (default "1")
    #    -Allow_Spreading_of_Tasks string
    #            -Allow_Spreading 1 ("0" Max Loaded Node Selection, "1" Least Loaded Node Selection) (default "1")
    #    -Allow_Task_Colocation string
    #            -Allow_Task_Colocation 1 (0 disable Colocation) (default "1")
    #    -Exclude_Nodelist string
    #            -Exclude_Nodelist <Node-1,Node-2..,Node-n> (eg. node1,node2,nodeN)
    #    -Include_Nodelist string
    #            -Include_Nodelist <Node-1,Node-2..,Node-n> (eg. node1,node2,nodeN)
    #    -Kube_Config_Path string
    #            Path to Kubernetes Configuration file (default "/home/master/.kube/Config")
    #    -Max_Node_Capacity string
    #            -Max_Node_Capacity 1 (100% Node Capacity) (default "1")
    #    -Namespaces string
    #            -Namespaces <namespace-1,namespace-2..,namespace-n> (default "default")
    #    -PID_Scale string
    #            -PID_Scale 5 (PID Scaler larger values accelerate resource scaling while losing accuracy) (default "3")
    #    -Scheduler_Name string
    #            -Scheduler Name (default "genisys")
    # 
```
Initialization Command Example:
```
     ./genisys -Namespaces namespace1,default -Include_Nodelist node1,node2,node3,node4
```    
    
Deploy Modified Genisys 

After we have configured the "Genisys/genisys_init_command.sh" script we are going to deploy Genisys.

Run:
```
(cd Genisys/ && make install)
```  
    
The "make install" option builds and pushes the container to the local docker registry and then deploys the modified image.
    
Check if Genisys Container is up and running.
    
Run:
``` 
kubectl get pods -o wide  
```    
    
Output:
```  
genisys-7f94f85b5b-9cxrt   1/1     Running   0          65s   10.244.1.3   minikube-m02   <none>           <none>
```      
    
</details>
    
    
# Virtual Cluster Deployment (Ethernet)

If only Ethernet adapters are available to the Cluster's Nodes
we can run Slurm HPC workloads by using Virtual Clusters (Ethernet version) 
that uses the classic TCP-UDP protocol for the containers communication.

<details>
<summary>Details</summary> 
    
This version runs on every classic Data-Center hardware setup 
as it does not require any specialized hardware.

## Virtual Cluster(Ethernet) Installation

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install-dockerhub)
``` 



## Deploy Modified Virtual-Cluster-Ethernet
    
If you wish to build your own modified image with different pre-installed libraries and dependencies modify the Dockerfile.
After you have configured the Dockerfile.

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install)
``` 
    
The "make install" option builds and pushes the container to the local docker registry and then deploys the modified image.  
    
</details>
    
# Virtual Cluster Deployment (Infiniband - RoCE)

If Infininand or RoCE adapters are available to the Cluster's Nodes
we can run Slurm HPC workloads by using Virtual Clusters (Infiniband - RoCE)
that uses the Infiniband protocol for the containers communication.

<details>
<summary>Details</summary> 
    
This version requires specialized hardware setup 
while performing significantly faster than the Ethernet counterpart.

## Virtual Cluster(Infiniband - RoCE) Installation

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Infiniband && make install-dockerhub)
```    
    
## Modified Virtual Cluster(Infiniband - RoCE) Installation
    
If the provided Demo does not run correctly and the Infiniband Virtual-Cluster deployed above does not operate correctly then it is possible that the Docker image supplied from Dockerhub is not compatible with your system. 
    
In this case you have to build the Infiniband Virtual-Cluster Image from scratch.
    
It is really IMPORTANT to initialize the KERNEL, OFED and REPO in the respective files. 
In order to use the infiniband adapters with Slurm and MPI support we have to install the Mellanox
OFED driver modules that require the kernel header files of the hosts OS kernel.    
 
Delete the previous deployment:
``` 
(cd Build/Virtual-Cluster-Slurm-Infiniband && make delete)
```      
    
Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Infiniband && make install)
```  
    
The "make install" option builds and pushes the container to the local docker registry and then deploys the modified image.      
    
</details>


# Prometheus Deployment

We are going to deploy Prometheus with a Custom-Metrics-Server in order for Genisys to pull custom performance metrics from the applications running.

<details>
<summary>Details</summary> 

## Deploy Prometheus Operator

Run:
```
(cd Build/Prometheus/prometheus-operator/ && make install-dockerhub)
```

## Deploy Prometheus

Run:
```
(cd Build/Prometheus/prometheus/ && make install-dockerhub)
```

## Deploy Custom Metric Server

Run:
```
(cd Build/Prometheus/metrics-server/ && make install-dockerhub)
```
</details>       

</details>    

# HPC Demos
<details>

# Single Node Demo (Only Kubernetes needed)

In this demo we are going to deploy Genisys + Virtual Clusters (Ethernet) on a single Node Kubernetes environment.

<details>
<summary>Details</summary>
    
This demo is for demonstration purposes.

## Prerequisites (IMPORTANT)

A working Kubernetes __1.18 -> 1.21__ installation.

## Configure and Deploy Genisys (Step 1)

### Install Genisys

Run:
```
(cd Genisys/ && make install-dockerhub)
```


Check if Genisys Container is up and running.
    
Run:
``` 
kubectl get pods -o wide  
```    
    
Output:
```    
genisys-XXX-XXX   1/1     Running   0          65s   10.244.1.3   minikube-m02   <none>           <none>
```    

## Virtual Cluster Deployment (Ethernet) (Step 2)

This version runs on every classic Data-Center hardware setup 
as it does not require any specialized hardware.

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install-dockerhub)
``` 


## MPI Demo Single Node (NAS Parallel Benchmarks) (Step 3)

In this example we are going to run some MPI benchmarks from the NAS Parallel Benchmarks suite, 
using the Virtual Cluster (Execution Environment) + Genisys (Scheduler) concept.

### Connect to a Virtual Cluster's Container
In order to use a Virtual Cluster we have to connect to one of its containers.

Check if Virtual Cluster's Containers are up and running.
Run:
``` 
kubectl get pods -o wide  
``` 

This command will return a list with all the running Kubernetes pods inside our namespace.
The Virtual Cluster's containers are named genisys-slurm-ethernet-XXX or genisys-slurm-infiniband-XXX.
We choose one the containers in order to connect with a terminal:.
    
Output:
``` 
genisys-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.1.57    worker-node1   <none>           <none>
genisys-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.0.112   master-node    <none>           <none>
```     
    
This command will open a terminal inside the Virtual Cluster's selected container:  
    
Run:
``` 
kubectl exec -it genisys-slurm-ethernet-XXX bash                             
``` 

### Download "NAS Parallel Benchmarks" 

Now we are going to download and build the "NAS Parallel Benchmarks" 
inside the /nfs/ directory.

Inside the container Run:
``` 
git clone  https://github.com/wzzhang-HIT/NAS-Parallel-Benchmark.git /nfs/NAS-Parallel-Benchmark &&
cd /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI && 
mv config/make.def.template config/make.def &&
mv config/suite.def.template config/suite.def &&
sed -i 's/MPIF77 = f77/MPIF77 = mpifort/' config/make.def &&
sed -i 's/S/C/' config/suite.def &&
sed -i 's/1/4/' config/suite.def &&
mkdir bin
``` 

### Build the Benchmarks

In order to build the benchmarks 

Run:
``` 
make suite
``` 

Now the binaries are located under the "/nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin" directory

### Run a Benchmark

In order to start an MPI benchmark with Slurm 

Run:
``` 
srun --mpi=pmix -N 1 -n 4 /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin/cg.C.4
``` 

The above command is going to run the cg.C.4 benchmark on 1 Node with 4 processes in total.

</details>





# Multi Node Minikube Demo (Easy to follow, all dependencies included by Minikube)


In this demo we are going to deploy Genisys + Virtual Clusters (Ethernet) on a 2 Node Minikube Kubernetes environment.

<details>
<summary>Details</summary>
    
This demo is for demonstration purposes.


## Prerequisites

A working Minikube installation.

See below how to set up Minikube:

https://minikube.sigs.k8s.io/docs/start/

Create a directory in order to be accessible by all Virtual Cluster Pods
```
sudo mkdir /var/nfs_share_dir/
sudo chmod -R 755 /var/nfs_share_dir
```



Add Minikube Docker Registry to the REGISTRY file in order to be used by the upcoming deployments
```
echo "127.0.0.1:5000" > REGISTRY
```

## Start the Minikube environment using Docker (Step 1)
```
minikube start --driver=docker  --kubernetes-version=v1.19.7 --nodes 2 --cpus 4 --mount-string="/var/nfs_share_dir/:/var/nfs_share_dir/" --mount
```

## Deploy Genisys (Step 2)
In this section we are going to deploy Genisys


Run:
```
(cd Genisys/ && make install-dockerhub)
```


Check if Genisys Container is up and running.
    
Run:
``` 
kubectl get pods -o wide  
```  
    
Output:   
```  
genisys-7f94f85b5b-9cxrt   1/1     Running   0          65s   10.244.1.3   minikube-m02   <none>           <none>    
```  
    
## Virtual Cluster Deployment (Ethernet) (Step 3)

If only Ethernet adapters are available to the Cluster's Nodes
we can run Slurm HPC workloads by using Virtual Clusters (Ethernet version) 
that uses the classic TCP-UDP protocol for the containers communication.

This version runs on every classic Data-Center hardware setup 
as it does not require any specialized hardware.

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install-dockerhub)
``` 

## MPI Demo Single Node (NAS Parallel Benchmarks) (Step 4)

In this example we are going to run some MPI benchmarks from the NAS Parallel Benchmarks suite, 
using the Virtual Cluster (Execution Environment) + Genisys (Scheduler) concept.

### Connect to a Virtual Cluster's Container
In order to use a Virtual Cluster we have to connect to one of its containers.

Check if Virtual Cluster's Containers are up and running.
    
Run:
    
``` 
kubectl get pods -o wide  
``` 

This command will return a list with all the running Kubernetes pods inside our namespace.
The Virtual Cluster's containers are named genisys-slurm-ethernet-XXX or genisys-slurm-infiniband-XXX.
   
Output:
    
``` 
genisys-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.1.57    worker-node1   <none>           <none>
genisys-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.0.112   master-node    <none>           <none>   
```    
 
    
    
We choose one the containers in order to connect with a terminal:.
    
This command will open a terminal inside the Virtual Cluster's selected container:  
    
Run:
``` 
kubectl exec -it genisys-slurm-ethernet-XXX bash                         
``` 

### Download "NAS Parallel Benchmarks" 

Now we are going to download and build the "NAS Parallel Benchmarks" 
inside the /nfs/ directory.

Inside the container Run:
``` 
git clone  https://github.com/wzzhang-HIT/NAS-Parallel-Benchmark.git /nfs/NAS-Parallel-Benchmark &&
cd /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI && 
mv config/make.def.template config/make.def &&
mv config/suite.def.template config/suite.def &&
sed -i 's/MPIF77 = f77/MPIF77 = mpifort/' config/make.def &&
sed -i 's/S/C/' config/suite.def &&
sed -i 's/1/4/' config/suite.def &&
mkdir bin
``` 

### Build the Benchmarks

In order to build the benchmarks 

Run:
``` 
make suite
``` 

Now the binaries are located under the "/nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin" directory

### Run a Benchmark

In order to start an MPI benchmark with Slurm 

Run:
``` 
srun --mpi=pmix -N 2 -n 4 /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin/cg.C.4
``` 

The above command is going to run the cg.C.4 benchmark on 2 Nodes with 4 processes in total.

</details>


# Complete Demo

In this example we are going to run some MPI benchmarks from the NAS Parallel Benchmarks suite, 
using the Virtual Cluster (Execution Environment) + Genisys (Scheduler) concept.

<details>
<summary>Details</summary> 
    
## Connect to a Virtual Cluster's Container (Step 1)
In order to use a Virtual Cluster we have to connect to one of its containers.

Run:
``` 
kubectl get pods -o wide
``` 

Output:
``` 
genisys-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.1.57    worker-node1   <none>           <none>
genisys-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.0.112   master-node    <none>           <none>
```
    
This command will return a list with all the running Kubernetes pods inside our namespace.
The Virtual Cluster's containers are named genisys-slurm-ethernet-XXX or genisys-slurm-infiniband-XXX.    
    
We choose one of the containers in order to connect with a terminal:

Run:
``` 
kubectl exec -it genisys-slurm-ethernet-XXX bash                         
``` 
    
This command will open a terminal inside the Virtual Cluster's selected container.       
    
## Download "NAS Parallel Benchmarks"  (Step 2)

Now we are going to download and build the "NAS Parallel Benchmarks" 
inside the NFS directory that is shared among the containers at the "/nfs" directory.

Inside the container 
    
Run:
``` 
git clone  https://github.com/wzzhang-HIT/NAS-Parallel-Benchmark.git /nfs/NAS-Parallel-Benchmark &&
cd /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI && 
mv config/make.def.template config/make.def &&
mv config/suite.def.template config/suite.def &&
sed -i 's/MPIF77 = f77/MPIF77 = mpifort/' config/make.def &&
mkdir bin
``` 

## Configure the Benchmarks  (Step 3)

Modify the config/suite.def in order to choose the parameters 
that the benchmarks are going to be compiled with.

Run:
``` 
nano config/suite.def
``` 

Example configuration:
``` 
# config/suite.def
# This file is used to build several benchmarks with a single command.
# Typing "make suite" in the main directory will build all the benchmarks
# specified in this file.
# Each line of this file contains a benchmark name, class, and number
# of nodes. The name is one of "cg", "is", "ep", mg", "ft", "sp", "bt",
# "lu", and "dt".
# The class is one of "S", "W", "A", "B", "C", "D", and "E"
# (except that no classes C, D and E for DT, and no class E for IS).
# The number of nodes must be a legal number for a particular
# benchmark. The utility which parses this file is primitive, so
# formatting is inflexible. Separate name/class/number by tabs.
# Comments start with "#" as the first character on a line.
# No blank lines.
# The following example builds 4 processor sample sizes of all benchmarks.

ep      C       4
cg      C       4

```     

    
## Build the Benchmarks  (Step 4)

In order to build the benchmarks 

Run:
``` 
make suite
``` 

Now the binaries are located under the "/nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin" directory

## Run a Benchmark  (Step 5)

In order to start an MPI benchmark with Slurm 

Run:
``` 
srun --mpi=pmix -N "NUM_OF_NODES" -n "NUM_OF_PROCESSES" /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin/"REPLACE_BENCHMARK_BINARY"
``` 

The above command is going to run the "REPLACE_BENCHMARK_BINARY" benchmark binary on "NUM_OF_NODES" nodes with "NUM_OF_PROCESSES" processes in total.

</details>

</details>  

# Prerequisites Installation Tutorials (Docker, Kubernetes, Helm, Docker Registry, NFS)

<details>
<summary>Details</summary>     

# Install Docker Engine

In this step we are going to install Docker engine (if not present) on our cluster's nodes.

<details>
<summary>Details</summary>
    
## Uninstall old versions

Run:
```
 sudo yum remove docker \
                  docker-client \
                  docker-client-latest \
                  docker-common \
                  docker-latest \
                  docker-latest-logrotate \
                  docker-logrotate \
                  docker-engine
```


## Install using the repository
Before you install Docker Engine for the first time on a new host machine, you need to set up the Docker repository. Afterward, you can install and update Docker from the repository.

### Set up the repository
Install the yum-utils package (which provides the yum-config-manager utility) and set up the repository.

Run:
```
 sudo yum install -y yum-utils
 
 sudo yum-config-manager \
    --add-repo \
    https://download.docker.com/linux/centos/docker-ce.repo
```

### Set up the repository

Run:
```
sudo yum install docker-ce docker-ce-cli containerd.io docker-compose-plugin
```

### Start Docker

Run:
```
sudo systemctl start docker
```

### Verify that Docker Engine is installed correctly by running the hello-world image.

Run:
```
sudo docker run hello-world
```
    
</details>
    
# Kubernetes Cluster Setup

Create a functional Kubernetes cluster.

<details>
<summary>Details</summary>  
  
***
Reference: https://phoenixnap.com/kb/how-to-install-kubernetes-on-centos
***

## Steps for Installing Kubernetes on CentOS 7

### Configure Kubernetes Repository
Kubernetes packages are not available from official CentOS 7 repositories. This step needs to be performed on the Master Node, and each Worker Node you plan on utilizing for your container setup. Enter the following command to retrieve the Kubernetes repositories.

Run:
```
cat >> /etc/yum.repos.d/kubernetes.repo <<EOF
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=0
EOF
```



### Install kubelet, kubeadm, and kubectl

Run:
```
sudo yum update                                            
sudo yum install -y kubelet-1.19.7 kubectl-1.19.7 kubeadm-1.19.7
```

### Set Hostname on Nodes
On master node 

Run:
```
sudo hostnamectl set-hostname master-node
```

On each worker node 

Run:
```
sudo hostnamectl set-hostname worker-node<node_number> # Where <node_number> the number of each node (worker-node1, worker-node2, worker-node3 ...)
```

In this example, the master node is now named master-node, while the worker nodes are 
named "worker-node<node_number>" (worker-node1, worker-node2, worker-node3 ...).

### Make a host entry or DNS record to resolve the hostname for all nodes

On each node run: 
```
sudo vi /etc/hosts
```

With the entry:
```
192.168.x.xxx master-node
192.168.x.xxx worker-node1
192.168.x.xxx worker-node2
192.168.x.xxx worker-node3
192.168.x.xxx worker-node4
```
Where we map each Node Name to the Node's IP.

### Disable Firewall
On each Node 

Run:
```
sudo systemctl stop firewalld.service
```

### Update Iptables Settings

Run:
```
sudo su
sudo cat <<EOF > /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system
```

### Disable SELinux
The containers need to access the host filesystem. SELinux needs to be set to permissive mode, which effectively disables its security functions.

Run:
```
sudo setenforce 0
sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
```

### Disable SWAP
Lastly, we need to disable SWAP to enable the kubelet to work properly:

Run:
```
sudo sed -i '/swap/d' /etc/fstab
sudo swapoff -a
```

## How to Deploy a Kubernetes Cluster

### Create Cluster with kubeadm

Run:
```
sudo kubeadm init --pod-network-cidr=10.244.0.0/16
```

### Manage Cluster as Regular User

Run:
```
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

### Set Up Pod Network

Run:
```
sudo kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```

### Check Status of Cluster

Run:
```
kubectl get nodes
```

Once a pod network has been installed, you can confirm that it is working by checking that the CoreDNS pod is running by typing:

Run:
```
kubectl get pods --all-namespaces
```

### Join Worker Node to Cluster
Use the kubeadm join command on each worker node to connect it to the cluster.

Run:
```
kubeadm join --discovery-token cfgrty.1234567890jyrfgd --discovery-token-ca-cert-hash sha256:1234..cdef 1.2.3.4:6443
```

Replace the codes with the ones from your master server. Repeat this action for each worker node on your cluster.

</details>

# Install Helm

<details>
<summary>Details</summary>

***
Reference: https://www.cyberithub.com/steps-to-install-helm-kubernetes-package-manager-on-linux/
***

## Prerequisites

a) You should have a running Linux Server.

b) You should have wget and tar utility installed in your Server.

c) You should have sudo or root access to run privileged commands.

## Download Helm

Run:
```
wget https://get.helm.sh/helm-v3.9.4-linux-amd64.tar.gz  
```

## Extract Helm 

Run:
```
tar -xvf helm-v3.9.4-linux-amd64.tar.gz  
```

## Move Binary 

Run:
```
mv linux-amd64/helm /usr/local/bin/helm
```

## Check Helm Version

Run:
```
helm version
```
    
</details>
    
# Install Docker Registry

<details>
<summary>Details</summary> 
   
***
Reference: https://faun.pub/install-a-private-docker-container-registry-in-kubernetes-7fb25820fc61
***

## Create a PersistentVolume
Create a PersistentVolume using a local storage-class, mounted on the "MASTER_TEMP" node (replace MASTER_TEMP with your Kubernetes master's actual hostname).

Run:
```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: docker-registry-pv
spec:
  capacity:
    storage: 60Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Delete
  storageClassName: docker-registry-local-storage
  local:
    path: /mnt/container-registry
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - MASTER_TEMP
EOF
```

## Verify the persistent volume was properly created

Run:
```
kubectl get pv container-registry-pv
```

## Create a docker-registry namespace

Run:
```
kubectl create namespace container-registry
```

## Create a PersistentVolumeClaim using a local storage-class, mounted on the MASTER_TEMP node

Run:
```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: docker-registry-pv-claim
  namespace: container-registry
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 60Gi
  storageClassName: docker-registry-local-storage
EOF
```

## Verify the persistent volume claim was properly created and it is bounded to

Run:
```
kubectl get pvc docker-registry-pv-claim --namespace container-registry
```

## Generate a user & password using htpasswd

Run:
```
bash <<'EOF'
   
# Change these credentials to your own
export REGISTRY_USER=admin
export REGISTRY_PASS=registry
export DESTINATION_FOLDER=${HOME}/temp/registry-creds
   
# Backup credentials to local files (in case you'll forget them later on)
mkdir -p ${DESTINATION_FOLDER}
echo ${REGISTRY_USER} >> ${DESTINATION_FOLDER}/registry-user.txt
echo ${REGISTRY_PASS} >> ${DESTINATION_FOLDER}/registry-pass.txt
   	
docker run --entrypoint htpasswd registry:2.7.0 \
    -Bbn ${REGISTRY_USER} ${REGISTRY_PASS} \
    > ${DESTINATION_FOLDER}/htpasswd
      
unset REGISTRY_USER REGISTRY_PASS DESTINATION_FOLDER
   
EOF
```

## Label MASTER_TEMP node with node-type=master

Run:
```
kubectl label nodes MASTER_TEMP node-type=master
```

## Add the twuni/docker-registry Helm repository successor of previous stable/docker-registry

Run:
```
helm repo add twuni https://helm.twun.io
```

## Update local Helm chart repository cache

Run:
```
helm repo update
```

## Search for latest twuni/docker-registry Helm chart version

Run:
```
helm search repo docker-registry
```

## Install the docker-registry Helm chart using the version from previous step:

Run:
```
helm upgrade --install docker-registry \
    --namespace container-registry \
    --set replicaCount=1 \
    --set secrets.htpasswd=$(cat $HOME/temp/registry-creds/htpasswd) \
    twuni/docker-registry \
    --version 1.10.1
```

## Verify installation
Make sure docker-registry pod is running

Run:
```
kubectl get pods --namespace container-registry | grep docker-registry
```

## Get Registry IP

Run:
```
kubectl get svc --namespace container-registry | grep docker-registry
```

The output will look like:
```
docker-registry   ClusterIP   REGISTRY_IP   <none>        5000/TCP   35m
```

## Add Registry to insecure-registries

On each Kubernetes node Run:
```
sudo nano /etc/docker/daemon.json
```

Add the entry below:
```
{"insecure-registries":["REGISTRY_IP:5000"]}
```
Where REGISTRY_IP is the IP of the registry

Restart docker afterwards

Run:
```
sudo systemctl restart docker
```

## Login to the Registry 

Run:
```
docker login    -u $(cat $HOME/temp/registry-creds/registry-user.txt)    -p $(cat $HOME/temp/registry-creds/registry-pass.txt)    https://REGISTRY_IP:5000
```
Where REGISTRY_IP is the IP of the registry.


## Create Kubernetes secret to access the registry.

Run:
```
kubectl create secret docker-registry docker-registry --docker-server=http://REGISTRY_IP:5000 --docker-username=admin --docker-password=registry -n default
```
By default --docker-username is 'admin'  and --docker-password is 'registry' .

## Path your serviceaccount with the secret

Run:
```
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "docker-registry"}]}'
```

</details>
    
# Setup NFS Server-Client in Centos 7

<details>
<summary>Details</summary> 
    
***
Reference: https://dev.to/prajwalmithun/setup-nfs-server-client-in-linux-and-unix-27id
***

## Setup NFS-Server

### Installing nfs-utils

Run:
```
sudo su 
yum install nfs-utils
```

### Choose the directory to share. If not present create one.

Run:
```
mkdir /var/nfs_share_dir
```

### Add permissions and ownwership privilages to the shared directory.

Run:
```
chmod -R 755 /var/nfs_share_dir
chown nfsnobody:nfsnobody /var/nfs_share_dir
```

### Start the nfs services.

Run:
```
systemctl enable rpcbind
systemctl enable nfs-server
systemctl enable nfs-lock
systemctl enable nfs-idmap
systemctl start rpcbind
systemctl start nfs-server
systemctl start nfs-lock
systemctl start nfs-idmap
```

### Configuring the exports file for sharing.

Run:
```
nano /etc/exports
```

Fill in the the file-shared path and clients details in /etc/exports.

Run:
```
/var/nfs_share_dir    CLIENT_IP1(rw,sync,no_root_squash)
/var/nfs_share_dir    CLIENT_IP2(rw,sync,no_root_squash)
```

### Restart the service

Run:
```
systemctl restart nfs-server
```

## Setup NFS-Client

### Installing nfs-utils

Run:
```
sudo su 
yum install nfs-utils
```

### Create a mount point

Run:
```
mkdir -p /var/nfs_share_dir
```

### Mounting the filesystem

Run:
```
mount -t nfs SERVER_IP:/var/nfs_share_dir /var/nfs_share_dir/
```

### Verify if mounted

Run:
```
df -kh
```

### Mounting permanently
Now if the client is rebooted, we need to remount again. So, to mount permanently,we need to configure /etc/fstab file.
Append this to "/etc/fstab".

```
SERVER_IP:/var/nfs_share_dir /var/nfs_share_dir nfs defaults 0 0
```

</details>
    
    
</details>
