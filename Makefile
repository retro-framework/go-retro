test-units:
	docker run -e CGO_ENABLED=0 --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.2-alpine3.7 go test -v github.com/retro-framework/go-retro/framework/...

test-integration:
	docker run -e CGO_ENABLED=0 --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.2-alpine3.7 go test -v -tags 'integration test' github.com/retro-framework/go-retro/framework/...

.PHONY: test-units test-integration

