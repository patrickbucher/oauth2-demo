#!/bin/sh

cd authserver
go run authserver.go &

cd ../resource
go run resource.go &

cd ../client
go run client.go &
