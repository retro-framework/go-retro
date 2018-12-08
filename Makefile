test-units:
			docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.2-alpine3.7 go test -v github.com/retro-framework/go-retro/framework/depot/...

test-integration:
			docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.2-alpine3.7 go test -v -tags 'integration test' github.com/retro-framework/go-retro/depot/...

.PHONY: test-units test-integration

