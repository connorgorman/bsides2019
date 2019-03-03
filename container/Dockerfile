FROM ubuntu:18.04

RUN apt-get update && apt-get install -y gnupg ca-certificates && \
    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 4052245BD4284CDD && \
    echo "deb https://repo.iovisor.org/apt/bionic bionic main" | tee /etc/apt/sources.list.d/iovisor.list

RUN apt-get update && apt-get install -y bcc-tools libbcc-examples linux-headers-4.15.0-1026-gcp

RUN apt-get install -y inetutils-ping curl netcat && mkdir -p /var/lib/bsides

COPY ./bin/ /

COPY ./scripts/ /
