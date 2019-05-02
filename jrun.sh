#!/bin/sh

./kill.sh

cd authserver
go run authserver.go &

cd ../resource
go run resource.go &

cd ../client
php -S localhost:1234 &
