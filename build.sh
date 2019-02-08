#!/bin/bash

docker build . -t db-operator-mysql --build-arg SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)"
