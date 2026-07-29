WRAPPER_VERSION_FILE ?= WRAPPER_VERSION
WRAPPER_VERSION := $(strip $(shell test -f $(WRAPPER_VERSION_FILE) && tr -d '[:space:]' < $(WRAPPER_VERSION_FILE)))

.PHONY: ci/lint ci/test ci/build lint test goproxy

ci/lint: lint

ci/test: test

ci/build: goproxy

lint:
	@UNFORMATTED="$$(gofmt -l $$(git ls-files '*.go'))"; \
	if [ -z "$$UNFORMATTED" ]; then \
		echo "Files formatted properly!"; \
		exit 0; \
	fi; \
	echo "The following files are not properly formatted:"; \
	echo "$$UNFORMATTED"; \
	exit 1

test:
	mkdir -p coverage
	go test -coverprofile coverage/coverage.out $(shell go list ./... | grep -v /vendor/) -p 1
	go test -race -coverprofile coverage/coverage_race.out $(shell go list ./... | grep -v /vendor/) -run "TestAutoRefreshTokenProviderTestSuite|TestManualRefreshTokenProviderTestSuite" -p 1

goproxy:
	@test -n "$(WRAPPER_VERSION)" || (echo "$(WRAPPER_VERSION_FILE) must contain a version" >&2; exit 1)
	./scripts/build_goproxy.sh "$(WRAPPER_VERSION)"
