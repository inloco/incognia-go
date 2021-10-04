test:
	mkdir -p coverage
	go test -coverprofile coverage/coverage.out $(shell go list ./... | grep -v /vendor/) -p 1