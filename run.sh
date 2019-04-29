<<<<<<< HEAD
#!/usr/bin/env bash

cd resource && go run main.go &
cd client && go run main.go &
cd authserver && go run main.go
=======
#!/bin/sh

./kill.sh

cd authserver
go run authserver.go &

cd ../resource
go run resource.go &

cd ../client
go run client.go &
>>>>>>> e0e06c912bece7a2e02a9ff0d82c80a9d6ba6d92
