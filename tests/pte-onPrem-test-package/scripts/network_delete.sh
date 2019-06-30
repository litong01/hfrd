#!/bin/bash
#
#   Network Deleter
#   Description: Used to delete network by service id
#   Dependencies: hfrd_test.cfg , service.json

. utils.sh

MAX_RETRY=5
PROG="[hfrd-delete]"
DefaultCredsDir=$HOME/results/creds

log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/networkDelete.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./network_delete.sh [OPTIONS]"
	log ""
	log "Options:"
    log "-c | --config  :   The path of hfrd configuratioin file"
	log "-s | --service :   The path of service.json which contains service id"
}

# Parse the input arguments
Parse_Arguments() {
	while [ $# -gt 0 ]; do
		case $1 in
			--config | -c)
				shift
				configPath=$1
				;;
			--service | -s)
				shift
				servicePath=$1
				;;
			--help | -h)
				Print_Help
				;;
		esac
		shift
	done

	# If no components are passed, print error
	if [[ -z "$servicePath" ]]; then
		log "ERROR: No enough parameters supplied."
		Print_Help
		exit 1
	fi
}

Parse_Arguments $@

# Sanity check
if [[ ! -f $servicePath ]]; then
	log "Missing hfrd configuration file, cannot continue."
	exit 1
fi

deleteService(){
	log "Delete service: $serviceid"
	NET_DELETE=$apiserverbaseuri'/service/'$serviceid'?env='$env
	statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
	 					-X DELETE \
						$NET_DELETE)
	if [ "$statuscode" -eq 202 ]; then
		return 0
 	else
		return 1
    fi
}

source $configPath
serviceid=$(jq -r '.serviceid' $servicePath)
runWithRetry 'deleteService'




