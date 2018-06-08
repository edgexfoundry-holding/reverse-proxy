#!/bin/bash
#
# Copyright (c) 2018
# Dell Technologies, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#

###
# Launches the EdgeX Go reverse proxy (must be previously built).
###

DIR=$PWD

# Kill edgex-proxy if running
function cleanup {
	pkill edgex-proxy
}

###
# Support logging
###
cd ..
# Add `edgex-` prefix on start, so we can find the process family
exec -a edgex-proxy ./edgex-proxy &

trap cleanup EXIT

while : ; do sleep 1 ; done