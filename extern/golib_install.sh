#!/bin/bash

# print every running process
set -x

BUILD_PATH=`pwd`

cd $BUILD_PATH/levigo && go clean -i ./ && CGO_CFLAGS="-I/usr/local/include" CGO_LDFLAGS="-L/usr/local/lib -lsnappy" go install ./

cd $BUILD_PATH/gorocks && go clean -i ./ && CGO_CFLAGS="-I/usr/local/include" CGO_LDFLAGS="-L/usr/local/lib -lsnappy" go install ./

cd $BUILD_PATH
