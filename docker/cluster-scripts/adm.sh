#!/bin/bash

source config.sh


function start() {
    set -e

    for i in `seq 0 2`; do
        eval __NODE__="$"NODE${i}""
        eval __NODE_LOCAL_PORT__="$"NODE${i}_LOCAL_PORT""
        eval __NODE_LOCAL_CONF_DIR__="$"NODE${i}_LOCAL_CONF_DIR""
        eval __NODE_LOCAL_LOG_DIR__="$"NODE${i}_LOCAL_LOG_DIR""
        eval __NODE_LOCAL_META_DIR__="$"NODE${i}_LOCAL_META_DIR""

        echo 'NODE               : '"${__NODE__}"
        echo 'NODE_LOCAL_PORT    : '"${__NODE_LOCAL_PORT__}"
        echo 'NODE_LOCAL_CONF    : '"${__NODE_LOCAL_CONF_DIR__}"
        echo 'NODE_LOCAL_LOG_DIR : '"${__NODE_LOCAL_LOG_DIR__}"
        echo 'NODE_LOCAL_META_DIR: '"${__NODE_LOCAL_META_DIR__}"
        echo ""

        mkdir -p ${__NODE_LOCAL_CONF_DIR__}
        mkdir -p ${__NODE_LOCAL_LOG_DIR__}
        mkdir -p ${__NODE_LOCAL_META_DIR__}

        docker run -d \
            --name ${__NODE__} \
            -h ${__NODE__} \
            --network ${NETWORK} \
            --restart on-failure \
            -p ${__NODE_LOCAL_PORT__}:20000 \
            -v ${__NODE_LOCAL_CONF_DIR__}:/rate-limiter/conf \
            -v ${__NODE_LOCAL_LOG_DIR__}:/rate-limiter/logs \
            -v ${__NODE_LOCAL_META_DIR__}:/rate-limiter/meta \
            ${IMAGE}
    done
}


function stop() {
    for i in `seq 0 2`; do
        eval __NODE__="$"NODE${i}""
        docker stop ${__NODE__}
    done
}


function umount() {
    for i in `seq 0 2`; do
        eval __NODE__="$"NODE${i}""
        docker stop ${__NODE__}
        docker rm -v ${__NODE__}
    done
}

function clean_logs() {
    for i in `seq 0 2`; do
        eval __NODE_LOCAL_LOG_DIR__="$"NODE${i}_LOCAL_LOG_DIR""

        echo 'NODE_LOCAL_LOG_DIR : '"${__NODE_LOCAL_LOG_DIR__}"
        echo ""

        rm -rf ${__NODE_LOCAL_LOG_DIR__}
    done
}


function clean_meta() {
    for i in `seq 0 2`; do
        eval __NODE_LOCAL_META_DIR__="$"NODE${i}_LOCAL_META_DIR""

        echo 'NODE_LOCAL_META_DIR : '"${__NODE_LOCAL_META_DIR__}"
        echo ""

        rm -rf ${__NODE_LOCAL_META_DIR__}
    done
}

case $1 in
    start)
        start
        ;;
    stop)
        stop
        ;;
    umount)
        umount
        ;;
    clean_logs)
        clean_logs
        ;;
    clean_meta)
        clean_meta
        ;;
    *)
        echo "Usage: ./adm.sh [options]"
        echo "Options:"
        echo "  start                       启动集群"
        echo "  stop                        停止集群"
        echo "  umount                      停止并卸载集群"
        echo "  clean_logs                  清理日志"
        echo "  clean_meta                  清理元数据"
        exit -1
        ;;
esac
exit 0
