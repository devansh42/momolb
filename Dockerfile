FROM golang:latest AS go_base
MAINTAINER devansh42

#Installing required packages
RUN apt update && apt install -y libpcap-dev
RUN mkdir -p /srv/momo
COPY . /srv/momo
WORKDIR /srv/momo
#downloading required go modules
RUN go build --ldflags '-linkmode external -extldflags "-static"' -o momo .
FROM alpine
WORKDIR /momo/bin
COPY --from=go_base /srv/momo/momo /momo/bin/    
CMD ["--help"]
RUN ["pwd"]
ENTRYPOINT ["/momo/bin/momo"]
