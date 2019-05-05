FROM golang:1.12.4-alpine
LABEL MAINTAINER="Patrick Bucher <patrick.bucher@stud.hslu.ch>"
RUN apk add git ca-certificates
WORKDIR /app
COPY go.mod resource.go /app/
RUN GO111MODULE=on go build -o /app/resource
EXPOSE 8000
CMD ["/app/resource"]
