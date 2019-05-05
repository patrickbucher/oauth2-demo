FROM golang:1.12.4-alpine
LABEL MAINTAINER="Patrick Bucher <patrick.bucher@stud.hslu.ch>"
RUN apk add git ca-certificates
WORKDIR /app
COPY go.mod client.go index.html gossip.html gossip.ico /app/
RUN GO111MODULE=on go build -o /app/client
ENV HOST=client PORT=1234 RESOURCE_HOST=resource RESOURCE_PORT=8000
EXPOSE "$PORT"
CMD ["/app/client"]
