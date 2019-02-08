#!/bin/bash -ex

docker run -it --rm \
	       -e GOPATH=/app \
	       -v "$(pwd):/app/src/db-operator-mysql" \
	       -v "$HOME/.ssh:/root/.ssh:ro" \
	       --workdir /app/src/db-operator-mysql \
	       instrumentisto/dep:0.5.0 ensure -v
