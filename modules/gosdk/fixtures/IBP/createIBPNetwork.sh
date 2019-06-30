#!/usr/bin/env bash


###############################################################################
#                                                                             #
#                              HFRD-TEST                              		  #
#                                                                             #
###############################################################################
WORKDIR=$(pwd)
. $WORKDIR/scripts/utils.sh

NETWORK_CREATE=false
RETRIEVE_CREDS=false
CHANNEL_CREATE=false
PROFILE_CREATE=false
USERCERTS_CREATE=false
SCFILE_CREATE=false
INSTALL_TOOLS=false
WORKLOAD_DRIVE=false
NETWORK_DELETE=false
workload=samplecc
# Help function
Print_Help() {
	echo ""
	echo "Usage:"
	echo "./docker-entrypoint.sh [OPTIONS]"
	echo ""
	echo "Options:"
	echo "-a | --all  : Run all of the processes.You must supply the workload name(marbles/samplecc) that you want to run"
	echo "-n | --createNetwork  : Create a new network.Will return network.json and service.json"
	echo "-r | --retrieveCreds  : (Can't be used together with option '-n')To use existing network for measurements,we need to get the service credentials.You can use this option to retrieve credentials belonged to an existing network"
	exit 1
}

# Parse the input arguments
Parse_Arguments() {
	input=false
	while [ $# -gt 0 ]; do
		case $1 in
			--all | -a)
				NETWORK_CREATE=true
				CHANNEL_CREATE=true
				PROFILE_CREATE=true
				USERCERTS_CREATE=true
				SCFILE_CREATE=true
				INSTALL_TOOLS=true
				WORKLOAD_DRIVE=true
				input=true
				;;
			--createNetwork | -n)
				NETWORK_CREATE=true
				input=true
				;;
			--retrieveCreds | -r)
				RETRIEVE_CREDS=true
				input=true
				;;
			--channels | -c)
				CHANNEL_CREATE=true
				input=true
				;;
			--profiles | -p)
				PROFILE_CREATE=true
				input=true
				;;
			--enroll | -e)
				USERCERTS_CREATE=true
				input=true
				;;
			--scfile | -s)
				SCFILE_CREATE=true
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
		echo "Must provide at least one argument"
		Print_Help
	fi

	# Conflict options check
	# Confilict options: '-n' and '-r'.
	if ${NETWORK_CREATE} && ${RETRIEVE_CREDS} ; then
		echo "Can't use option '-n' and '-r' at the same time"
		Print_Help
	fi
}

log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/hfrd.log
}

cleanEnvironment(){
	rm -rf $HOME/results/*
}
createRequiredDirs(){
	if [[ ! -f $HOME/results/tmp ]];then
		mkdir -p $HOME/results/tmp
	fi
	if [[ ! -f $HOME/results/logs ]];then
		mkdir -p $HOME/results/logs
	fi
	if [[ ! -f $HOME/results/SCFiles ]];then
		mkdir -p $HOME/results/SCFiles
	fi
}

Parse_Arguments $@

createRequiredDirs

# Environments
PROG="[hfrd]"
source $WORKDIR/hfrd_test.cfg


cd $WORKDIR/scripts
if $NETWORK_CREATE ; then
	log "---------------------------------"
	log "---------------------------------"
	log "Create ${env}-${name} Network 	  "
	log "---------------------------------"
	log "---------------------------------"
	cleanEnvironment
	createRequiredDirs
	./network_generator.sh -c $WORKDIR/hfrd_test.cfg
	if [ $? -ne 0 ]; then
		exit 2
	fi
fi
