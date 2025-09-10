.PHONY: test test-unit test-integration formatting

test: test-unit test-integration formatting

test-unit:
	go test $$(go list ./... | grep -v /tests/integration) -cover

test-integration:
	go test ./tests/integration/...

formatting:
	test -z $$(go fmt ./...)