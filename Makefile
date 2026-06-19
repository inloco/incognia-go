GO_CACHE_DIR ?= $(CURDIR)/local/go-build-cache
GO_MOD_CACHE_DIR ?= $(CURDIR)/local/go-mod-cache

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
	mkdir -p coverage $(GO_CACHE_DIR) $(GO_MOD_CACHE_DIR)
	PACKAGES="$$(GOCACHE="$(GO_CACHE_DIR)" GOMODCACHE="$(GO_MOD_CACHE_DIR)" go list ./... | grep -v /vendor/)"; \
	GOCACHE="$(GO_CACHE_DIR)" GOMODCACHE="$(GO_MOD_CACHE_DIR)" go test -coverprofile coverage/coverage.out $$PACKAGES -p 1; \
	GOCACHE="$(GO_CACHE_DIR)" GOMODCACHE="$(GO_MOD_CACHE_DIR)" go test -race -coverprofile coverage/coverage_race.out $$PACKAGES -run "TestAutoRefreshTokenProviderTestSuite|TestManualRefreshTokenProviderTestSuite" -p 1

goproxy:
	./scripts/build_goproxy.sh
