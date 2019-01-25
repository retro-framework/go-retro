PACKAGE?=github.com/retro-framework/go-retro/framework/...
GOTESTFLAGS?=
GOTESTTAGS?=test

test: test-units test-integration

test-units:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go test $(GOTESTFLAGS) -tags "unit $(GOTESTTAGS)" $(PACKAGE)

test-integration:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go test $(GOTESTFLAGS) -tags "integration $(GOTESTTAGS)" $(PACKAGE)

fmt:
	docker run --rm --network none --volume ${PWD}:/go/src/github.com/retro-framework/go-retro golang:1.11.4-stretch go fmt $(PACKAGE)

.PHONY: test-units test-integration fmt

