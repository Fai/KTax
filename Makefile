.PHONY: test verify

test:
	go clean -testcache && go test -v -race github.com/Fai/assessment-tax/...

verify:
	gofmt ./... && go vet ./... && go mod tidy && go mod verify