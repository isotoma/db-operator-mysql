GO ?= go
BINDIR := $(CURDIR)/bin
GOFLAGS :=

BUILD_NUMBER=`cat build-number`

.PHONY: release build dep strip docker test dep-links 

release: build strip

build: vendor #.vendor-links.stamp
	CGO_ENABLED=0 GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) github.com/isotoma/db-operator-mysql/cmd/...

strip:
	strip bin/*

test:
	cd pkg/heartbeat && go test

vendor:
	dep ensure	

