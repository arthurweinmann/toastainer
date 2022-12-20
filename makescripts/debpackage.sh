#!/usr/bin/env bash
set -euo pipefail

BUILDDIR=$1;
CURRDIR=$2;
VERSION=$3;
REVISION=$4;
ARCHI=$5;

DEBDIR=toastainer_$VERSION-${REVISION}_$ARCHI

cd $BUILDDIR;

rm -rf $DEBDIR
rm -f $DEBDIR.deb

mkdir -p $DEBDIR
mkdir -p $DEBDIR/DEBIAN
mkdir -p $DEBDIR/usr/local/bin
mkdir -p $DEBDIR/etc/systemd/system
mkdir -p $DEBDIR/usr/share/toastainer

cp -r toastainer nsjail $DEBDIR/usr/local/bin
cp -r web images  $DEBDIR/usr/share/toastainer

cp -r $CURRDIR/releasing/debian/* $DEBDIR/DEBIAN
cp -r $CURRDIR/releasing/systemd/toastainer.service $DEBDIR/etc/systemd/system

cp -r $CURRDIR/config_example.json $DEBDIR/usr/share/toastainer

dpkg-deb --build --root-owner-group $BUILDDIR/$DEBDIR

rm -rf $DEBDIR