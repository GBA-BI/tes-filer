FROM golang:1.20.5-bullseye AS builder
COPY . /go/src/github.com/GBA-BI/vetes-filer
WORKDIR /go/src/github.com/GBA-BI/vetes-filer
RUN make build-linux

FROM debian:11
RUN apt update \
     && apt install -y --no-install-recommends ca-certificates default-mysql-client curl \
     && echo "Asia/Shanghai" > /etc/timezone \
     && dpkg-reconfigure -f noninteractive tzdata \
     && apt clean \
     && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /app/log 
COPY --from=builder /go/src/github.com/GBA-BI/vetes-filer/bin/vetes-filer /app/filer
WORKDIR     /app
ENV         HOME=/app
ENTRYPOINT  ["/app/filer"]
