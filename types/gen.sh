#!/bin/bash

__src_dir__=./src

__dst_dir__=./pbwrappers


mkdir -p ${__dst_dir__}

protoc -I=${__src_dir__} --go_out=${__dst_dir__} ${__src_dir__}/*.proto
