VERSION = snapshot
GHRFLAGS =
# Git current tag
GIT_TAG=$(shell git tag -l --contains HEAD | sed -e "s/^v//")

.PHONY: build release

default: build

fgo-build:
	fgo -p releases -b homebrew-swipe build ${GIT_TAG}

build:
	goxc -d=releases -bc="linux,386 darwin" -pv=$(VERSION)

release:
	ghr -u swipe-io -replace $(GHRFLAGS) v$(VERSION) releases/$(VERSION)

chglog:
	git-chglog -o CHANGELOG.md

check:
	go vet ./...
	go test -v ./...

#build:	check
	#go build -o swipe ./cmd/swipe
#
#install: build
#	mv ./swipe ${GOPATH}/bin