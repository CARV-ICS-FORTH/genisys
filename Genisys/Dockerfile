FROM golang:1.17.2-alpine AS build

ENV CGO_ENABLED=0

WORKDIR /home

# Separate downloading dependencies from build
COPY scheduler/go.mod ./
RUN apk add git
RUN go mod download

# Build
COPY . .
RUN (cd scheduler && chmod 777 * && go build -o genisys)
RUN (cp -r scheduler/*  /home) 
RUN (rm -rf scheduler && chmod 777 *)


ARG KUBECTL_VERSION=v1.19.8
RUN wget -O /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

CMD ["/usr/local/bin/genisys"]
