#!/bin/bash

scp www_jivecake_com.pem www_jivecake_com.key settings.js restart.sh $1:~
ssh $1 'bash -s' < install-dependencies.sh