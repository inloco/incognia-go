test:
	mkdir -p coverage
	go test -coverprofile coverage/coverage.out $(shell go list ./... | grep -v /vendor/) -p 1
	go test -race -coverprofile coverage/coverage_race.out $(shell go list ./... | grep -v /vendor/) -run "TestAutoRefreshTokenProviderTestSuite|TestManualRefreshTokenProviderTestSuite" -p 1