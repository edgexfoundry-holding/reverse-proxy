#
# Copyright (c) 2018 Dell Technologies, Inc
#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: build run test

build: edgex-proxy
	go build ./...

edgex-proxy:
	go build -o ./edgex-proxy

run:
	cd bin && ./proxy-launch.sh

test:
	go test ./...
	go vet ./...
