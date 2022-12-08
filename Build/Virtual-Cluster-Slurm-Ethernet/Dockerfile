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

#Install wget tool
RUN yum install -y wget

#Install cgroups resource limiting drivers
RUN yum install -y libcgroup libcgroup-tools

#Install python3
RUN yum install -y python3

#Add genisys connector files
ADD job_test.c /root/
COPY --from=build /slurm_k8s_plugin /root/

#Add mpi_init script to the profile.d directory
ADD mpi_init.sh /etc/profile.d/

# Instal openMPI
RUN wget https://download.open-mpi.org/release/open-mpi/v4.1/openmpi-4.1.4.tar.gz && \
    tar -xvzf openmpi-4.1.4.tar.gz && \
    yum install -y perl gcc gcc-c++ make gcc-gfortran && \
    (cd openmpi-4.1.4 && ./configure --with-slurm CC=gcc CXX=g++ FC=gfortran && make all install) && \
    rm -f openmpi-4.1.4.tar.gz


#RUN yum install -y openmpi-devel



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