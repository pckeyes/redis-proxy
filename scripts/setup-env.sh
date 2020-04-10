#!/bin/bash

go get github.com/go-redis/redis
go get github.com/gorilla/mux
go get github.com/stretchr/testify/assert

go build ./src/redis-proxy/cmd/server
go build ./src/redis-proxy/cmd/client

