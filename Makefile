chglog:
	git-chglog -o CHANGELOG.md

check:
	go vet ./...
	go test -v ./...

build:	check
	go build -o swipe ./cmd/swipe

install: build
	mv ./swipe ${GOPATH}/bin