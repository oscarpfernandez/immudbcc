FROM golang:1.14

RUN apt-get -y update && apt-get install -y jq

ENV GO111MODULE=on
ENV OS=linux
ENV ARCH=amd64
ENV PRJ_SRC_PATH=/go/src/github.com/oscarpfernandez/imudbcc

RUN mkdir -p ${PRJ_SRC_PATH}
WORKDIR ${PRJ_SRC_PATH}
COPY  . .

RUN CGOGOARCH=${ARCH} GOOS=${OS} go build \
    -ldflags "-s -w" \
    -mod vendor -v \
    ${PRJ_SRC_PATH}/cmd/immudb-doc/...

RUN ./run_itests.sh