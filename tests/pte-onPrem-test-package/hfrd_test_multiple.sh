#!/usr/bin/env bash

. scripts/utils.sh
# Help function
Print_Help() {
	echo ""
	echo "Usage: use ./hfrd_test_multiple.sh -n to create multiple networks"
	echo "./hfrd_test_multiple.sh [OPTIONS]"
	echo ""
	echo "Options:"
	echo "-n | --createNetwork  : Create a new network.Will return network.json and service.json for each network"
  exit 1
}

# Parse the input arguments
Parse_Arguments() {
	input=false
	while [ $# -gt 0 ]; do
		case $1 in
			--createNetwork | -n)
        startCmd=$startCmd'-n '
				NEW_NETWORK=true
				input=true
				;;
			--help | -h)
				input=true
				Print_Help
				;;
		esac
		shift
	done

	if ! $input; then
		echo "Must provide one argument"
		Print_Help
	fi
	# Conflict options check
	# Confilict options: '-n' and '-r'.
	if ${NEW_NETWORK} && ${REUSE_NETWORK} ; then
		echo "Can't use option '-n' and '-r' at the same time"
		Print_Help
	fi
}
# Default settings
NEW_NETWORK=false
REUSE_NETWORK=false
eval configPath=$(pwd)/conf/networks.yaml

# Use startCmd as the entrypoint of docker container
startCmd='./scripts/multiple-support/docker-entrypoint-multiple.sh '
Parse_Arguments $@

# Load the configurations
create_variables $configPath

# Sanity check
if [[ ! -d ${dir_ROOT} ]]; then
		mkdir -p ${dir_ROOT}
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
docker build -t pte-fab --build-arg uid=$(id -u) --build-arg gid=$(id -g) .

# Start to run PTE driver by using the image built above
if [[ $env == 'cm' ]];then
  # Only On-Prem environment needs to add hosts volume
  docker run --rm -v $(pwd)/conf/hosts:/etc/hosts -v ${dir_ROOT}:${dir_ROOT} pte-fab /bin/bash -c "$startCmd"
else
	docker run --rm -v ${dir_ROOT}:${dir_ROOT} pte-fab /bin/bash -c "$startCmd"
fi