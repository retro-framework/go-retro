PACKAGE?=github.com/retro-framework/go-retro/framework/...
EXTRAFLAGS?=-race

test-units:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go test -v $(EXTRAFLAGS) -tags 'test' $(PACKAGE)

test-integration:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go test -v $(EXTRAFLAGS) -tags 'integration test' $(PACKAGE)

fmt:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go fmt $(PACKAGE)

.PHONY: test-units test-integration fmt

