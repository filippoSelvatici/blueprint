#!/bin/bash

set -x

pushd build
set -a
source .env
set +a
./wlgen/wlgen_proc/wlgen_proc/wlgen_proc
popd

