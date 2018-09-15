#!/bin/bash

cluster="127.0.0.1:20000,127.0.0.1:20001,127.0.0.1:20002"

rctype="cWNsb3VkI0VJUExpc3Q="
#rctype=""



case $1 in
        stress)
                ./ratelimiterctl --cluster ${cluster} stress --rctype ${rctype} --robot 5 --worker 24 --duration 30
                ;;
        regist)
                ./ratelimiterctl --cluster ${cluster} regist --rctype ${rctype} --quota 10 --resetInterval 3
                ;;
        rcList)
                if [ -z "${rctype}" ];then
                        ./ratelimiterctl --cluster ${cluster} rcList
                else
                        ./ratelimiterctl --cluster ${cluster} rcList --rctype ${rctype}
                fi
                ;;
        delete)
                ./ratelimiterctl --cluster ${cluster} delete --rctype ${rctype}
                ;;
        *)
                echo "./cmd.sh [regist|rcList|delete]"
                exit -1
                ;;
esac
exit 0
