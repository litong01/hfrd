#!/usr/bin/env bash

# ROOTDIR is an environment variable in runnerbox container
# ROOTDIR is the working directory on Jenkins host
if [ -z $ROOTDIR ]; then
	TOP=$(pwd)/workdir
else
	TOP=${ROOTDIR}/workdir
fi
echo "***************************************************"

echo "THIS IS TOP DIRECTORY FROM HOST"

echo $TOP

echo "***************************************************"
source $TOP/hfrd_test.cfg
# $(pwd)/workdir is a directory in runnerbox container
# this entrypoint shell script will be exected inside runnerbox
docker build -t pte-hfrd --build-arg user=$(whoami) --build-arg uid=$(id -u) --build-arg gid=$(id -g) -f Dockerfile-jks .

# ${TOP}/results should be a directory on Jenkins host
if [[ $env == 'cm' ]]; then
	docker run --rm -v ${TOP}/hosts:/etc/hosts -v ${TOP}/results:/home/$(whoami)/results pte-hfrd ./docker-entrypoint-jks.sh
else
	docker run --rm -v ${TOP}/results:/home/$(whoami)/results pte-hfrd ./docker-entrypoint-jks.sh
fi
