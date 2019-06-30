#!/bin/bash

#   Worload Driver
#   Description: Used to drive traffic to blockchain network
#   Dependencies: Path of PTE SCFile
#   TODO: This should be able
#   One parameter :
#		1)path of PTE SCFile
WORKDIR=$(pwd)
. utils.sh

MAX_RETRY=5
PROG="[hfrd-pte-test-local]"
DefaultSCFile=${HOME}/fabric-sdk-node/test/PTE/SCFiles/config-chan1-TLS.json

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

msp_id=$(jq -r keys[0] $networkPath)
networkId=$(jq -r .$msp_id.network_id $networkPath)
if [[ ! -f ${HOME}/results/SCFiles/config-net-${networkId}.json ]]; then
	log "Missing SCFile,must get the PTE SCFile before continue."
	exit 1
fi

# # Copy PTE SCFile to PTE/SCFiles
cp -f ${HOME}/results/SCFiles/config-net-${networkId}.json ${HOME}/fabric-sdk-node/test/PTE/SCFiles/config-chan1-TLS.json

# Copy and replace the userInputs with msp id
rm -rf ${HOME}/fabric-sdk-node/test/PTE/userInputs
cp -rf ${WORKDIR}/../workloads/$workload ${HOME}/fabric-sdk-node/test/PTE/userInputs

source ${HOME}/.nvm/nvm.sh
cd ${HOME}/fabric-sdk-node/test/PTE

# Install chaincodes
./pte_driver.sh userInputs/runCases.install.txt | tee -a $HOME/results/logs/workload.log
sleep 5

# Instantiate chaincodes
./pte_driver.sh userInputs/runCases.instantiate.txt | tee -a $HOME/results/logs/workload.log
sleep 5

# Drive traffic to blockchian network
./pte_driver.sh userInputs/runCases.run.txt | tee -a $HOME/results/logs/workload.log
