#!/usr/bin/env bash

cd resource && go run main.go &
cd client && go run main.go &
cd authserver && go run main.go
