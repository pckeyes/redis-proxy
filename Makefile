.PHONY: test
test:
	docker-compose run --rm server  ./src/redis-proxy/scripts/test.sh

.PHONY: up
up:
	docker-compose up -d server

.PHONY: sh
sh: up
	docker exec -it goprojects_server_1 bash

