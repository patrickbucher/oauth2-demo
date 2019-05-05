FROM golang:1.12.4-alpine
LABEL MAINTAINER="Patrick Bucher <patrick.bucher@stud.hslu.ch>"
RUN apk add git ca-certificates
WORKDIR /app
COPY go.mod client.go gossip.html gossip.ico /app/
RUN GO111MODULE=on go build -o /app/client
EXPOSE 1234
CMD ["/app/client"]
