FROM golang:1.15

RUN apt-get -y update && apt-get install -y jq uuid-runtime
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.32.2

ENV GO111MODULE=on
ENV GOOS=linux
ENV CGOGOARCH=amd64
ENV PRJ_SRC_PATH=/go/src/github.com/oscarpfernandez/imudbcc

RUN mkdir -p ${PRJ_SRC_PATH}
WORKDIR ${PRJ_SRC_PATH}
COPY  . .

RUN ./run_ci.sh
RUN ./run_itests.sh