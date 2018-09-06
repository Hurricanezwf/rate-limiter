#!/bin/bash

# 项目根目录
__ROOT_DIR__=$GOPATH/src/github.com/Hurricanezwf/rate-limiter


function types(){
    protoc -I=./ --go_out=${__ROOT_DIR__}/types/  ./types.proto
}

function meta(){
    protoc -I=./ --go_out=${__ROOT_DIR__}/meta/ ./meta.proto
}

function api(){
    protoc -I=./ --go_out=${__ROOT_DIR__}/proto/ ./api.proto
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
