FROM ubuntu:18.04 AS git
RUN apt-get -y update && apt-get -y install git
RUN git clone https://github.com/georgzerb/Prometheus-Operator.git /prometheus-operator 

FROM golang:1.15.2-alpine AS build
COPY --from=git /prometheus-operator  /go/src/github.com/coreos/prometheus-operator 
RUN	CGO_ENABLED=0 go build -a -tags netgo -o /build/operator github.com/coreos/prometheus-operator/cmd/operator

FROM ubuntu:18.04
COPY --from=build /build/operator /bin/
ENTRYPOINT ["/bin/operator"]
