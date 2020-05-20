#!/usr/bin/env bash
set -ex

PACKAGE="oreka-api"

die(){
 echo $1;
 exit 1
}


SED_CMD="sed"
if [ "$(uname)" == "Darwin" ]; then
    SED_CMD="gsed"
fi

function replacePlaceHolders() {
    file="$1"
    $SED_CMD -i -e "s/_PACKAGE_/$PACKAGE/g" $file
}



BUILD_ROOT=$(mktemp -d)
VERSION=$(date +%s)
cp -r debian/* $BUILD_ROOT/

mkdir -p $BUILD_ROOT/usr/local/bin/


cp ../bin/oreka-api  $BUILD_ROOT/usr/local/bin/oreka-api


#replacing constants
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/prerm"
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/postrm"
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/postinst"
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/control"

$SED_CMD -i "s/_VERSION_/$VERSION/g" $BUILD_ROOT/DEBIAN/control

rm -f $PACKAGE.deb
dpkg-deb --build $BUILD_ROOT $PACKAGE.deb

rm -rf $BUILD_ROOT