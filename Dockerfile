
FROM golang:alpine AS go_base
MAINTAINER devansh42

#Installing required packages
RUN apk update && apk add libpcap-dev 
RUN mkdir -p /srv/momo
COPY . /srv/momo
WORKDIR /srv/momo
#downloading required go modules
RUN go mod tidy
RUN go build -o momo .

FROM alpine
WORKDIR /momo/bin
COPY --from=go_base /srv/momo/momo /momo/bin/    
CMD ["--help"]
ENTRYPOINT ./momo