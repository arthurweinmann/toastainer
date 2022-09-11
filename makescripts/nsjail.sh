#!/usr/bin/env bash
set -euo pipefail

BUILDDIR=$1
CURRDIR=$2

apt-get -y update && apt-get install -y \
    autoconf \
    bison \
    flex \
    gcc \
    g++ \
    git \
    libprotobuf-dev \
    libnl-route-3-dev \
    libtool \
    make \
    pkg-config \
    protobuf-compiler

cd $BUILDDIR && rm -rf tmp && mkdir tmp && cd tmp && git clone https://github.com/google/nsjail.git && cd nsjail && make && mv nsjail $BUILDDIR && cd ../../ && rm -rf tmp || rm -rf tmp