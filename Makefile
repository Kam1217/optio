.PHONY: test test-unit test-integration

test: test-unit test-integration

test-unit:
	go test $$(go list ./... | grep -v /tests/integration) -cover

test-integration:
	go test ./tests/integration/...