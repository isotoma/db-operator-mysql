.PHONY: all build deps

all: deps build

build:
	./build.sh

deps:
	./update_deps.sh
