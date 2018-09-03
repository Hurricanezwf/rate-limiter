#!/bin/bash

function types(){
    protoc -I=./ --go_out=../  ./types.proto
}

function meta(){
    protoc -I=./ --go_out=../../meta/ ./meta.proto
}

function api(){
    protoc -I=./ --go_out=../../proto/ ./api.proto
}


case $1 in
    types)
        types
        ;;
    meta)
        meta
        ;;
    api)
        api
        ;;
    all)
        types
        meta
        api
        ;;
    *)
        echo "Usage: ./gen.sh [types | meta | api | all]"
        exit -1
        ;;
esac
exit 0
