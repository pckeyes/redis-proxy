#!/bin/bash

./src/redis-proxy/scripts/setup-env.sh

go test ./src/redis-proxy/cmd/server/main_test.go

./server &
./client
