#!/bin/bash

ssh-keygen -b 4048 -f ~/.ssh/id_rsa -t rsa -N "" -C $1

yum -y update
yum install -y git-all nano man man-pages curl
yum clean all

#https://docs.docker.com/engine/installation/linux/centos/#install-with-the-script

curl -fsSL https://get.docker.com/ | sh
sudo systemctl enable docker.service
sudo systemctl start docker