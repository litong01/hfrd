#!/bin/bash
#
#   Network Generator
#   Description: Used to generate networks.
#   Dependencies: hfrd_test.cfg
# 	Note: hfrd_test.cfg decribes the network info,including:
#		Section for API Server
#			apiuser: The user account you want to use.
#			apiserverhost: The API server host,like ' http://hfrdrestsrv.rtp.raleigh.ibm.com:8080 '
#			apiversion: The API  version
#			apiserverbaseuri="$apiserverhost/$apiversion/$apiuser"
#		Section for network topology
#			env: Where the network will be created. Currently we support 'bxstaging','bxproduction','cm'
#			name: The type of this network. You can specify 'sp'(starter plan) or 'ep'(enterprise plan) for this
#			For ep networks:
#			loc: The location of IBP Cluster you want to create the network
#			numOfOrgs:The number of orgs you want to create
#			numOfPeers:The number of peers per org you want to create
#			ledgerType:The worldstate ledger you want to use

log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/network_generator.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./network_generator.sh [OPTIONS]"
	log ""
	log "Options:"
	log "-c | --config  :   The path of hfrd configuratioin file"
}

# Parse the input arguments
Parse_Arguments() {
	while [ $# -gt 0 ]; do
		case $1 in
			--config | -c)
				shift
				configPath=$1
				;;
			--help | -h)
				Print_Help
				;;
		esac
		shift
	done

	# If no components are passed, print error
	if [ -z "$configPath" ]; then
		log "ERROR: No hfrd configuration file supplied."
		Print_Help
		exit 1
	fi

}

Parse_Arguments $@

# Sanity check
if [[ ! -f $configPath ]]; then
	log "Missing hfrd configuration file, cannot continue."
	exit 1
fi

# environment
source $configPath
DATESTR=$(date +%Y-%m-%d" "%H:%M:%S" "%p)
PROG="[hfrd-network-generator]"
# STEP 0. API SERVER ENDPOINT DETAILS & SETUP #
log "$DATESTR"
log "1. Starting HFRD API  TEST Script"

# STEP 1. REQUEST A  PLAN NETWORK FROM IBP #
if [[ $name == 'ep' ]];then
	REQUEST_BODY="{\"env\": \"$env\",\"loc\": \"$loc\",\"name\":\"$name\",\"config\":{\"numOfOrgs\": $numOfOrgs,\"numOfPeers\": $numOfPeers,\"ledgerType\": \"$ledgerType\"}}"
else
	REQUEST_BODY="{\"env\": \"$env\"}"
fi

requestid=$(curl --silent --include -X POST \
	"$apiserverbaseuri/service" -d "$REQUEST_BODY" | awk -v FS="Request-Id: " 'NF>1 {print $2}')
requestid=${requestid%$'\r'}
if [ -z "$requestid" ]; then
	log "Error: Request ID invalid, exit 1."
	exit 1
fi
log "\t1.a) Server acknowledged with Request ID = $requestid"

# STEP 2. RETRIEVE serviceid#
log "2. Using Request ID = $requestid To Retrieve Service Credentials"
nextwaittime=0
while [ $nextwaittime -le 100 ]; do
	statuscode=$(curl --silent --output /dev/null --write-out %{http_code} \
		"$apiserverbaseuri/service?requestid=$requestid&env=$env")
	if [ "$statuscode" -eq 200 ]; then
		# key package is ready, retrieve it from the server
		curl --silent "$apiserverbaseuri/service?requestid=$requestid&env=$env" >| "$HOME/results/tmp/package.tar"
		tar xf $HOME/results/tmp/package.tar -C $HOME/results/
		mv $HOME/results/workdir/results $HOME/results/creds
		# extract service id from the package
		serviceid=$(jq -r '.serviceid' $HOME/results/creds/service.json)
		serviceid=${serviceid%$'\r'}
		printf "\n"
		break
	elif [ "$statuscode" -eq 404 ] || [ "$statuscode" -eq 202 ]; then
		printf "."
		sleep $((nextwaittime++))
	else
		printf "\n"
		log "Error: Unable to retrieve service credentials, exit 2."
		exit 2
	fi
done
if [ -z "$serviceid" ]; then
	printf "\n"
	log "Error: Service Credentials invalid, exit 2."
	exit 2
fi
log "\t2.a) Retrieved  serviceid, Service ID = $serviceid"
