FROM golang:1.12.4-alpine
LABEL MAINTAINER="Patrick Bucher <patrick.bucher@stud.hslu.ch>"
RUN apk add git ca-certificates
WORKDIR /app
COPY go.mod authserver.go auth.html lock.ico /app/
RUN GO111MODULE=on go build -o /app/authserver
ENV HOST=authserver PORT=8443
EXPOSE "$PORT"
CMD ["/app/authserver"]
