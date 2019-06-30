#!/usr/bin/env bash

# Help function
Print_Help() {
	echo ""
	echo "Usage: use ./hfrd_test.sh -a 'workload' or run each step individually and in order"
	echo "./hfrd_test.sh [OPTIONS]"
	echo ""
	echo "Options:"
	echo "-a | --all  : Run all of the processes.You must supply the workload name(marbles or samplecc) that you want to run"
	echo "-n | --createNetwork  : Create a new network.Will return network.json and service.json"
	echo "-r | --retrieveCreds  : (Can't be used together with option '-n')To use existing network for measurements,we need to get the service credentials.You can use this option to retrieve credentials belonged to an existing network"
	echo "-c | --channels  : Create and join channels based on the channels configuration file"
	echo "-p | --profiles  : Fetch connection profiles"
	echo "-e | --enroll  : Enroll and upload user certs"
	echo "-s | --scfile  : Get PTE SCFile from connection files"
	echo "-i | --install  : Install SAR and NMON on network containers"
	echo "-w | --workload  : Set the PTE workload that you want to drive.You must supply the workload name(marbles/samplecc) that you want to run"
	echo "-d | --delete  : Delete network"
  exit 1
}

# Parse the input arguments
Parse_Arguments() {
	input=false
	while [ $# -gt 0 ]; do
		case $1 in
			--all | -a)
        startCmd=$startCmd'-a '
				input=true
				;;
			--createNetwork | -n)
        startCmd=$startCmd'-n '
				NEW_NETWORK=true
				input=true
				;;
			--retrieveCreds | -r)
        startCmd=$startCmd'-r '
				REUSE_NETWORK=true
				input=true
				;;
			--channels | -c)
        startCmd=$startCmd'-c '
				input=true
				;;
			--profiles | -p)
        startCmd=$startCmd'-p '
				input=true
				;;
			--enroll | -e)
        startCmd=$startCmd'-e '
				input=true
				;;
			--scfile | -s)
        startCmd=$startCmd'-s '
				input=true
				;;
			--install | -i)
        startCmd=$startCmd'-i '
				input=true
				;;
			--workload | -w)
        startCmd=$startCmd'-w '
				input=true
				;;
			--delete | -d)
        startCmd=$startCmd'-d '
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
# Environments
NEW_NETWORK=false
REUSE_NETWORK=false
# Use startCmd as the entrypoint of docker container
startCmd='./docker-entrypoint.sh '
Parse_Arguments $@
source $(pwd)/hfrd_test.cfg

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
docker build -t pte-fab --build-arg uid=$(id -u) --build-arg gid=$(id -g) .

# Start to run PTE driver by using the image built above
HOME_COTNAINER='/home/'$(whoami)
if [[ $env == 'cm' ]];then
  # Only On-Prem environment needs to add hosts volume
  docker run --rm -v $(pwd)/conf/hosts:/etc/hosts -v $RESULTS_DIR:${HOME_COTNAINER}/results/ pte-fab /bin/bash -c "$startCmd"
else
	docker run --rm -v $RESULTS_DIR:${HOME_COTNAINER}/results/ pte-fab /bin/bash -c "$startCmd"
fi