FROM golang:1.12.4-alpine
LABEL MAINTAINER="Patrick Bucher <patrick.bucher@stud.hslu.ch>"
RUN apk add git ca-certificates
WORKDIR /app
COPY go.mod resource.go /app/
RUN GO111MODULE=on go build -o /app/resource
ENV HOST=resource PORT=8000 AUTHSERVER_HOST=authserver AUTHSERVER_PORT=8443
EXPOSE "$PORT"
CMD ["/app/resource"]
