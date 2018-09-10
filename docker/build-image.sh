#!/bin/bash

set -e

APPName="rate-limiter"
VERSION=`git branch | grep '*' | awk '{print $2}'`

pushd ./ >/dev/null
cd ../
make release
popd > /dev/null
mv ../$APPName ./

docker build -t $APPName:$VERSION .
rm ./$APPName
