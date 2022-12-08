FROM debian:buster-slim

ENV NODE_VERSION 10.14.1

RUN apt-get -y update 
RUN apt-get -y update --fix-missing
RUN apt-get -y install vim nano

# install dependencies
RUN apt-get install -y iproute2 curl

# create a large random file to upload
RUN head -c 2M </dev/urandom > /upload.data

#install wondersharper nc network limiter with the network limiting script
RUN apt-get -y install wondershaper
COPY net-limiter /net-limiter

#install node js (needed for the metrics server)
RUN apt-get -y install nodejs

ENV YARN_VERSION 1.12.3


# add our user and group first to make sure their IDs get assigned consistently, regardless of whatever dependencies get added
RUN groupadd --system --gid 11211 memcache && useradd --system --gid memcache --uid 11211 memcache

# ensure SASL's "libplain.so" is installed as per https://github.com/memcached/memcached/wiki/SASLHowto
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
	libsasl2-modules \
	; \
	rm -rf /var/lib/apt/lists/*


#install Memcached
ENV MEMCACHED_VERSION 1.5.6
ENV MEMCACHED_SHA1 5d3b5af3ce0a1483d655017db7228bcaeff10d47

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y \
      memcached=${MEMCACHED_VERSION}* \
 && sed 's/^-d/# -d/' -i /etc/memcached.conf \
 && rm -rf /var/lib/apt/lists/*
 
##########################################################################
COPY docker-entrypoint.sh /usr/local/bin/
RUN ln -s usr/local/bin/docker-entrypoint.sh /entrypoint.sh # backwards compat
ENTRYPOINT ["docker-entrypoint.sh"]

#server.js is node js server that read the real time metric from a file and serves it to prometheus
COPY server.js /server.js

#memcache-top-master is an utility that reports the current memcached performance 
COPY memcache-top-master /memcache-top-master
COPY report.txt /report.txt
COPY metric-exporter /metric-exporter

##########################################################################
RUN [ "apt-get", "-q", "update" ]
RUN [ "apt-get", "-qy", "--force-yes", "upgrade" ]
RUN [ "apt-get", "-qy", "--force-yes", "dist-upgrade" ]
RUN [ "apt-get", "install", "-qy", "--force-yes", \
	"perl", \
	"build-essential", \
	"cpanminus" ]
RUN [ "apt-get", "clean" ]
RUN [ "rm", "-rf", "/var/lib/apt/lists/*", "/tmp/*", "/var/tmp/*" ]
##########################################################################

##########################################################################
#install cpu limter
RUN  apt-get install -y wget

RUN  apt-get install -y unzip

RUN  apt-get install -y cpulimit

COPY cpu-limiter cpu-limiter

#install cgroup drivers
RUN apt-get install -y cgroup-tools
##########################################################################



RUN apt-get update && \
	apt-get -y install sudo

RUN apt-get -y update && apt-get -y install procps
RUN adduser --disabled-password --gecos '' docker
RUN adduser docker sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers




USER docker
EXPOSE 11211
EXPOSE 8080
COPY server server
COPY memcached memcached
COPY exporter exporter
COPY start start
CMD ./start

