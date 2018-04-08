#!/bin/sh

if [ ! -n "$TZ" ]; then
    export TZ="Asia/Shanghai"
fi

ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
echo $TZ > /etc/timezone && \
gotask start
