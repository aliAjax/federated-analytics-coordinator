.PHONY: fmt vet test race build smoke count
fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')
vet:
	go vet ./...
test:
	go test ./...
race:
	go test -race ./...
build:
	go build ./cmd/coordinator ./cmd/worker
smoke:
	./scripts/smoke.sh
count:
	./scripts/count-go-lines.sh
