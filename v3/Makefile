VERSION = snapshot
GHRFLAGS =

.PHONY: build release

default: build

build:
	go mod download
	go build -o $(GOPATH)/bin/swipe3 cmd/swipe/main.go

chglog:
	git-chglog -o CHANGELOG.md

check:
	go vet ./...
	go test -v ./...