#!/bin/bash

protoc -I=./ --go_out=../  ./types.proto

protoc -I=./ --go_out=../../meta/v2/ ./meta.proto
