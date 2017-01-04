#!/bin/bash

#Install boilerplate
yum -y update
yum install -y git-all nano man man-pages curl
yum clean all

#https://docs.docker.com/engine/installation/linux/centos/#install-with-the-script
#Install docker

curl -fsSL https://get.docker.com/ | sh
sudo systemctl enable docker.service
sudo systemctl start docker

#Test docker
sudo docker run --rm hello-world
docker rmi hello-world