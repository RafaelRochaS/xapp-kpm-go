FROM nexus3.o-ran-sc.org:10002/o-ran-sc/bldr-ubuntu22-c-go:1.0.0 AS build-app

RUN apt update && apt install -y iputils-ping net-tools curl sudo ca-certificates

RUN wget --content-disposition https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/rmr_4.7.0_amd64.deb/download.deb && dpkg -i rmr_4.7.0_amd64.deb && rm -rf rmr_4.7.0_amd64.deb
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/rmr-dev_4.7.0_amd64.deb/download.deb && dpkg -i rmr-dev_4.7.0_amd64.deb && rm -rf rmr-dev_4.7.0_amd64.deb

RUN mkdir -p /go/src/app
COPY . /go/src/app

WORKDIR "/go/src/app"

ENV GO111MODULE=on GO_ENABLED=0 GOOS=linux

RUN go build -a -installsuffix cgo -o app ./cmd


FROM ubuntu:22.04

ENV CFG_FILE=config/config-file.json
ENV RMR_SEED_RT=config/uta_rtg.rt

RUN mkdir /config

COPY --from=build-mec-app /go/src/app/app /
COPY --from=build-mec-app /go/src/app/config/* /config/
COPY --from=build-mec-app /usr/local/lib /usr/local/lib

RUN ldconfig

RUN chmod 755 /app
CMD /app
