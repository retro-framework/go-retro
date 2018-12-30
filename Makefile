PACKAGE?=github.com/retro-framework/go-retro/framework/...

test-units:
	docker run -e CGO_ENABLED=0 --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-alpine3.7 go test -v -tags 'test' $(PACKAGE)

test-integration:
	docker run -e CGO_ENABLED=0 --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-alpine3.7 go test -v -tags 'integration test' $(PACKAGE)

fmt:
	docker run -e CGO_ENABLED=0 --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-alpine3.7 go fmt $(PACKAGE)

.PHONY: test-units test-integration fmt

