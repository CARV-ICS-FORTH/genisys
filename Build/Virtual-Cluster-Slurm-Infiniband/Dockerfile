FROM golang:1.17.2-alpine AS build

ENV CGO_ENABLED=0

WORKDIR /home

# Separate downloading dependencies from build
COPY slurm_kubernetes_connector_plugin/go.mod ./
RUN go mod download

# Build
COPY . .
RUN (cd slurm_kubernetes_connector_plugin && chmod 777 * && go mod tidy && go build -o ./slurm_k8s_plugin && cp -r slurm_k8s_plugin / && cd ..)



ARG CENTOS=7.7.1908
FROM centos:7.7.1908

# Install kubectl
ARG KUBECTL_VERSION=v1.19.8
RUN curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

WORKDIR /root

# Install kernel, development tools, and Mellanox OFED
ARG CENTOS=7.7.1908
ARG KERNEL=3.10.0-957.10.1
ARG OFED=4.7-3.2.9.0
ARG REPO=4.7-3.2.9.0

RUN yum install -y ${REPO}/kernel-${KERNEL}.el7.x86_64.rpm && \
    yum install -y ${REPO}/kernel-headers-${KERNEL}.el7.x86_64.rpm && \
    yum install -y ${REPO}/kernel-devel-${KERNEL}.el7.x86_64.rpm && \
    yum install -y ${REPO}/kernel-debug-devel-${KERNEL}.el7.x86_64.rpm && \
    yum groupinstall -y "Development Tools" && \
    yum install -y cmake python-devel pciutils numactl-libs valgrind-devel tcsh libdb-devel lsof zlib-devel ethtool numactl-devel libmnl-devel binutils-devel iptables-devel openssl-devel libselinux-devel libnl3 elfutils-devel libnl3-devel && \
    curl -L -o MLNX_OFED_SRC-${OFED}.tgz http://www.mellanox.com/downloads/ofed/MLNX_OFED-${OFED}/MLNX_OFED_SRC-${OFED}.tgz && \
    tar -xvzf MLNX_OFED_SRC-${OFED}.tgz && \
    ./MLNX_OFED_SRC-${OFED}/install.pl --all && \
    echo "1 s" | mpi-selector-menu && \
    rm -rf MLNX_OFED_SRC-${OFED} && \
    rm -f MLNX_OFED_SRC-${OFED}.tgz && \
    yum remove -y kernel-devel kernel-debug-devel linux-firmware && \
    rpm -e --nodeps kernel && \
    rpm -qa | grep "^kmod" | xargs rpm -e --nodeps && \
    yum install -y kmod kmod-libs && \
    yum clean all \
    && rm -rf /var/cache/yum \
        /tmp/* \
        /var/tmp/* \
        /usr/share/man \
        /usr/share/doc \
        /usr/share/doc-base

#Install wget tool
RUN yum install -y wget

#Install cgroups resource limiting drivers
RUN yum install -y libcgroup libcgroup-tools

#Install python3
RUN yum install -y python3

#Add genisys connector files
ADD job_test.c /root/
COPY --from=build /slurm_k8s_plugin /root/

# Install PMIX
ARG PMIX=2.2.4
RUN yum install -y python3 hwloc hwloc-devel libevent libevent-devel && \
    yum groupinstall -y "Development Tools" && \
    curl -LO https://github.com/openpmix/openpmix/releases/download/v${PMIX}/pmix-${PMIX}.tar.bz2 && \
    tar -jxvf pmix-${PMIX}.tar.bz2 && \
    (cd pmix-${PMIX} && ./configure --prefix=/usr --libdir=/usr/lib64 --sysconfdir=/etc --localstatedir=/var && make && make install) && \
    rm -rf pmix-${PMIX} && \
    rm -f pmix-${PMIX}.tar.bz2 

# Install SLURM
ARG SLURM=20-11-3-1
RUN wget https://github.com/SchedMD/slurm/archive/refs/tags/slurm-${SLURM}.zip && \
    unzip slurm-${SLURM}.zip && \
    mv /root/job_test.c slurm-slurm-${SLURM}/src/plugins/select/cons_common/ && \
    (cd slurm-slurm-${SLURM} && ./configure --prefix=/usr --libdir=/usr/lib64 --sysconfdir=/etc --localstatedir=/var && make && make install) && \
    rm -f slurm-${SLURM}.zip

RUN chmod +x slurm_k8s_plugin && mv slurm_k8s_plugin /usr/bin/ 

# Install SSH
RUN yum -y install openssh-server openssh-clients && \
    yum clean all \
    && rm -rf /var/cache/yum \
        /tmp/* \
        /var/tmp/* \
        /usr/share/man \
        /usr/share/doc \
        /usr/share/doc-base

RUN TINI_VERSION="0.18.0" && \
    yum install -y https://github.com/krallin/tini/releases/download/v${TINI_VERSION}/tini_${TINI_VERSION}-amd64.rpm

# Install nano
RUN yum -y install nano

ADD start.sh /root/

# Create Slurm log directory
RUN mkdir -p /tmp/slurm_nums/

ENTRYPOINT [ "/usr/bin/tini", "--" ]
CMD [ "/root/start.sh" ]