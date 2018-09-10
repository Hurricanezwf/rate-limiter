#!/bin/bash

# 镜像名
IMAGE="rate-limter:v0.1.1"

# Docker网络名
NETWORK="cmp"


## ----------------- 【结点0配置】 -------------- ##
# 结点名
NODE0="rate-limiter-0"

# 本地端口
NODE0_LOCAL_PORT=20000

# 本地配置目录
NODE0_LOCAL_CONF_DIR="/home/zwf/app/rate-limiter-cluster/node0/conf"

# 本地日志目录
NODE0_LOCAL_LOG_DIR="/home/zwf/app/rate-limiter-cluster/node0/logs"

# 本地元数据目录
NODE0_LOCAL_META_DIR="/home/zwf/app/rate-limiter-cluster/node0/meta"




## ----------------- 【结点1配置】 -------------- ##
# 结点名
NODE1="rate-limiter-1"

# 本地端口
NODE1_LOCAL_PORT=20001

# 本地配置目录
NODE1_LOCAL_CONF_DIR="/home/zwf/app/rate-limiter-cluster/node1/conf"

# 本地日志目录
NODE1_LOCAL_LOG_DIR="/home/zwf/app/rate-limiter-cluster/node1/logs"

# 本地元数据目录
NODE1_LOCAL_META_DIR="/home/zwf/app/rate-limiter-cluster/node1/meta"





## ----------------- 【结点2配置】 -------------- ##
# 结点名
NODE2="rate-limiter-2"

# 本地端口
NODE2_LOCAL_PORT=20002

# 本地配置目录
NODE2_LOCAL_CONF_DIR="/home/zwf/app/rate-limiter-cluster/node2/conf"

# 本地日志目录
NODE2_LOCAL_LOG_DIR="/home/zwf/app/rate-limiter-cluster/node2/logs"

# 本地元数据目录
NODE2_LOCAL_META_DIR="/home/zwf/app/rate-limiter-cluster/node2/meta"
