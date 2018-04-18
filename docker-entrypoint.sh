#!/bin/sh

if [ ! -n "$TZ" ]; then
    export TZ="Asia/Shanghai"
fi

# set timezone
ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
echo $TZ > /etc/timezone 

# k8s config  switch
if [ -f "/usr/local/gotask/config/.gotask.yaml" ]; then
    ln -s  /usr/local/gotask/config/.gotask.yaml /usr/local/gotask/.gotask.yaml
fi

# apply config
echo "===start==="
cat /usr/local/gotask/.gotask.yaml
echo "====end===="

# run command
if [ ! -z "$1" ]; then
    /usr/local/gotask/gotask $@
else
    /usr/local/gotask/gotask start
fi
