#!/bin/bash

killall rate-limiter

cd ../ && make

cp  rate-limiter $HOME/app/rate-limiter-2/bin/
cp  rate-limiter $HOME/app/rate-limiter-1/bin/
cp  rate-limiter $HOME/app/rate-limiter-0/bin/
