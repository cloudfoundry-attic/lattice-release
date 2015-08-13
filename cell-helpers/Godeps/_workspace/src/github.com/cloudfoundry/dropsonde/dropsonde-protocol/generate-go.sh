#!/bin/bash

dir_resolve()
{
cd "$1" 2>/dev/null || return $?  # cd to desired directory; if fail, quell any error messages but return exit status
echo "`pwd -P`" # output full, link-resolved path
}

set -e

TARGET=`dir_resolve $1`
if [ -z "$TARGET" ]; then
    echo 'USAGE: `generate-go.sh TARGET_PATH`'
    echo ''
    echo 'TARGET_PATH is where you would like the control and events packages to be generated.'
    exit 1
fi

go get github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto}

pushd events
mkdir -p $TARGET/events
protoc --plugin=$(which protoc-gen-gogo) --gogo_out=$TARGET/events --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf:. *.proto
popd

pushd control
mkdir -p $TARGET/control
protoc --plugin=$(which protoc-gen-gogo) --gogo_out=$TARGET/control --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf:. *.proto
popd
