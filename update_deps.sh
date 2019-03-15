#!/bin/bash -ex

docker run -it --rm \
	       -e GOPATH=/app \
	       -v "$(pwd):/app/src/github.com/isotoma/db-operator-mysql" \
	       -v "$HOME/.ssh:/root/.ssh:ro" \
	       --workdir /app/src/github.com/isotoma/db-operator-mysql \
	       instrumentisto/dep:0.5.0 ensure -v -update
