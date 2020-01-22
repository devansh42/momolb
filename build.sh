#!/bin/sh
#Script to build momo docker image
sudo -s
apt update
apt install -y docker.io
git clone -b testing https://github.com/devansh42/momo.git
cd momo
docker build -t devansh42/momo .
docker login
docker push devansh42/momo

