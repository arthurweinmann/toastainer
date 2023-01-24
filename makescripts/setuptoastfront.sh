#!/usr/bin/env bash
set -euo pipefail

BUILDDIR=$1

cd $BUILDDIR
mkdir tmp
cd tmp

git clone https://github.com/toastate/toastfront.git
cd toastfront/cmd/toastfront
go build
mv -f toastfront $BUILDDIR


rm -rf $BUILDDIR/tmp