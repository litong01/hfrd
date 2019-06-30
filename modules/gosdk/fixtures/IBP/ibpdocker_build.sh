#!/usr/bin/env bash

# Environments
apiuser=sunhwei
apiserverhost=http://hfrdrestsrv.rtp.raleigh.ibm.com:8080
apiversion=v1
apiserverbaseuri="$apiserverhost/$apiversion/$apiuser"
serviceid={serviceid}
RESULTS_DIR=$(pwd)/results/
REUSE_NETWORK=false
# Use startCmd as the entrypoint of docker container
startCmd='./docker-entrypoint.sh '

# Sanity check
if [[ ! -d $RESULTS_DIR ]]; then
		mkdir -p $RESULTS_DIR
fi

# Build driver image
cp -f Dockerfile.in Dockerfile
OS_ARCH='amd64'
OSTYPE=$(uname)
if [[ $(uname -m) == 's390x' ]];then
    OS_ARCH='s390x'
fi

if [ "$OSTYPE" == "Darwin" ]; then
  ISED='-i "" '
else
  ISED='-i'
fi

eval sed $ISED 's/FABRIC_VERSION/1.1.0/g' Dockerfile
eval sed $ISED "s/OS_ARCH/$OS_ARCH/g" Dockerfile
eval sed $ISED "s/TEST_USER/$(whoami)/g" Dockerfile

# Build pte-fab image with same user name/id and group id
docker build -t hfrdibp --build-arg uid=$(id -u) --build-arg gid=$(id -g) .

# Start to run PTE driver by using the image built above
#HOME_COTNAINER='/home/'$(whoami)
#	docker run --rm -d -v $RESULTS_DIR:${HOME_COTNAINER}/results/ netprofiles4ibp /bin/bash 
