#!/bin/bash

#   Worload Driver
#   Description: Used to drive traffic to blockchain network
#   Dependencies: 
#   TODO: 
WORKDIR=$(pwd)
. utils.sh

MAX_RETRY=5
PROG="[hfrd-Caliper-test-local]"

log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/workload.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./workloads.sh [OPTIONS]"
	log ""
	log "Options:"
	log "-w | --workload  : Set the PTE workload that you want to drive"
	log "-n | --network  :   The path of network.json which contains all of the service credentials in blockchain network,including msp_id,networkId,API key/secret,"
}

# Parse the input arguments
Parse_Arguments() {
	while [ $# -gt 0 ]; do
		case $1 in
			--workload | -w)
				shift
				workload=$1
				;;
			--network | -n)
				shift
				networkPath=$1
				;;
			--help | -h)
				Print_Help
				;;
		esac
		shift
	done

	# If no components are passed, print error
	if [[ -z "$workload" || -z "$networkPath" ]];then
		log "ERROR: No enough parameters supplied."
		Print_Help
		exit 1
	fi
}

Parse_Arguments $@

# Sanity check
if [[ ! -f $networkPath ]]; then
	log "Missing networkPath,cannot continue."
	exit 1
fi

#  Copy Caliper network file to caliper/benchmark/
cp -f ${HOME}/caliper/network/fabric/ibpep/ibpep_fabric.json ${HOME}/caliper/benchmark/simple

source ${HOME}/.nvm/nvm.sh
cd ${HOME}/caliper/benchmark/simple

# Install chaincodes

# Instantiate chaincodes

# Drive traffic to blockchian network
node main.js -c ibpep_config.json | tee -a $HOME/results/logs/workload.log
