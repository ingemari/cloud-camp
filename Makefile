.PHONY: build
build:
	@go build cmd/main.go

.PHONY: run
run:
	@go run cmd/main.go

.PHONY: k6
k6:
	k6 run tests/k6/test_k6.js