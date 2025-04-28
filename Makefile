.PHONY: build
build:
	@go build cmd/main.go

.PHONY: run
run:
	@go run cmd/main.go

.PHONY: k6
k6:
	@k6 run tests/k6/test_k6.js

.PHONY: docker-redis
docker-redis:
	@docker run --name my-redis -d -p 6379:6379 redis:alpine

.PHONY: redis
redis:
	@docker exec -it my-redis redis-cli