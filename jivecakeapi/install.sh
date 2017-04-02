#!/bin/bash

scp star_jivecake_com.jks settings.yml docker-compose.yml $1:~
scp id_rsa $1:~/.ssh
scp id_rsa.pub $1:~/.ssh
ssh $1 "chmod 600 ~/.ssh/id_rsa ~/.ssh/id_rsa.pub"
ssh $1 'bash -s' < install-dependencies.sh