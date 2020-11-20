VERSION = snapshot
GHRFLAGS =

.PHONY: build release

default: build

fgo-build:
	fgo -p releases -b homebrew-swipe build ${NEXT_TAG}

build:
	goxc -d=releases -bc="linux,386 darwin" -pv=$(VERSION)

release:
	ghr -u swipe-io -replace $(GHRFLAGS) v$(VERSION) releases/$(VERSION)

chglog:
	git-chglog -o CHANGELOG.md

check:
	go vet ./...
	go test -v ./...