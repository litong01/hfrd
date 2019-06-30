#!/bin/bash
#
#   Connection Profiles Generator
#   Description: Used to generate connection profiles.
#   Dependencies: hfrd_test.cfg , service.json

# environment
log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/connectionProfiles.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./connectionProfile_generator.sh [OPTIONS]"
	log ""
	log "Options:"
	log "-c | --config  :   The path of hfrd configuratioin file"
	log "-s | --service  :   The path of service.json "
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
	if [[ -z "$configPath" || -z "$servicePath" ]];then
		log "ERROR: No enough parameters supplied."
		Print_Help
		exit 1
	fi

}

Parse_Arguments $@

# Sanity check
if [[ ! -f $configPath ]]; then
	log "Missing hfrd configuration file, cannot continue."
	exit 1
elif [[ ! -f $servicePath ]]; then
	log "Missing service.json, cannot continue."
	exit 1
fi

source $configPath
PROG="[hfrd-connectionProfiles-generator]"

serviceid=$(jq -r '.serviceid' $servicePath)
if [[ "$serviceid" == null ]]; then
	log "Missing serviceid,cannot continue"
	exit 1
fi

cleanOldCPs(){
	rm -rf $HOME/results/creds/ConnectionProfile_*
}

# STEP 1. REQUEST Connection Profiles#
log "1. Using Service ID = $serviceid To Request Connection Profile"
CP_requestid=$(curl --silent --include \
	"$apiserverbaseuri/service/$serviceid/profile?env=$env" | awk -v FS="Request-Id: " 'NF>1{print $2}')
CP_requestid=${CP_requestid%$'\r'}
if [ -z "$CP_requestid" ]; then
	printf "\n"
	log "Error: Request ID invalid, exit 1."
	exit 1
fi
log "\t1.a) Server acknowledged with Request ID = $CP_requestid"

# STEP 2. RETRIEVE Connection Profiles #
log "2. Using Request ID = $CP_requestid and Service ID = $serviceid To Retrieve Connection Profile"
next_wait_time=0
while [ $next_wait_time -lt 30 ]; do
	statuscode=$(curl --silent --output /dev/null --write-out %{http_code} \
		"$apiserverbaseuri/service/$serviceid/profile?requestid=$CP_requestid&env=$env")
	if [ "$statuscode" -eq 200 ]; then
		# connection profile package is ready, retrieve from server and add to package.tar
		curl --silent "$apiserverbaseuri/service/$serviceid/profile?requestid=$CP_requestid&env=$env" >| "$HOME/results/tmp/package.tar"
		tar xf $HOME/results/tmp/package.tar -C $HOME/results/
		cleanOldCPs
		mv $HOME/results/workdir/results/* $HOME/results/creds/
		printf "\n"
		break
	elif [ "$statuscode" -eq 404 ] || [ "$statuscode" -eq 202 ]; then
		printf "."
		sleep $((next_wait_time++))
	else
		printf "\n"
		log "Error: unable to retrieve connection profile."
		exit 2
	fi
done
log "\t2.a)Connection Profiles retrieved."