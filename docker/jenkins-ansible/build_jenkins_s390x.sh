#!/bin/bash
rm -rf docker
git clone https://github.com/jenkinsci/docker.git

cd docker
sed -i 's/openjdk/s390x\/openjdk/g' Dockerfile
docker build -t jenkins-s390x:latest .

cd ..
docker build -t hfrd/jenkins-s390x:latest -f Dockerfile.s390x .




