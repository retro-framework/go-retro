PACKAGE?=github.com/retro-framework/go-retro/framework/...
GOTESTFLAGS?=-race -v
GOTESTTAGS?=test

test-units:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go test $(GOTESTFLAGS) -tags "$(GOTESTTAGS)" $(PACKAGE)

test-integration:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go test $(GOTESTFLAGS) -tags "$(GOTESTTAGS) integration" $(PACKAGE)

fmt:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go fmt $(PACKAGE)

.PHONY: test-units test-integration fmt

