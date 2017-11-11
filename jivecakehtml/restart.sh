#!/bin/bash

NAME="jivecakehtml"

ids=$(docker ps --filter="name=$NAME" -qa)

if [ "$ids" != "" ]; then
    docker stop "$ids"
    docker rm "$ids"
fi

docker run \
    -d \
    -e BRANCH=master \
    -e REPOSITORY=https://github.com/troisio/jivecakehtml.git \
    -e COMMIT=HEAD \
    -p 80:80 \
    -p 443:443 \
    --name=$NAME \
    -v /root/www_jivecake_com.key:/root/www_jivecake_com.key \
    -v /root/www_jivecake_com.pem:/root/www_jivecake_com.pem \
    -v /root/settings.js:/root/settings.js \
    jivecakehtml