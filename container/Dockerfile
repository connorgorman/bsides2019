FROM ubuntu:18.04

COPY ./bin/fsmonitor /fsmonitor

COPY ./scripts/filewriter.sh /filewriter.sh

RUN apt-get update && apt-get install -y gnupg ca-certificates && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 4052245BD4284CDD && \
    echo "deb https://repo.iovisor.org/apt/bionic bionic main" | tee /etc/apt/sources.list.d/iovisor.list

RUN apt-get update && apt-get install -y bcc-tools libbcc-examples linux-headers-4.15.0-1026-gcp
