#!/bin/bash
#
#   Retrieve service credentials for an existing network
#   Description: In a clean environment,if you want to reuse one existing blockchian service,you need to firstly get
#				 the service credentials,including network.json/service.json/connection profiles
#   Dependencies: serviceid of the existing network

# environment
log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/retrieveCreds.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./retrieve_creds.sh [OPTIONS]"
	log ""
	log "Options:"
	log "-s | --serviceid  :   The serviceid of the existing network"
	log "-c | --config  :   The path of hfrd configuratioin file"
}

# Parse the input arguments
Parse_Arguments() {
	while [ $# -gt 0 ]; do
		case $1 in
			--serviceid | -s)
				shift
				serviceid=$1
				;;
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
	if [[ -z "$configPath" || -z "$serviceid" ]];then
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
fi

source $configPath
PROG="[hfrd-retrieve-credentials]"
cleanOldCPs(){
	rm -rf $HOME/results/creds
}

# STEP 1. REQUEST Service credentials#
log "1. Using Service ID = $serviceid To Request service credentials"
CP_requestid=$(curl --silent --include \
	"$apiserverbaseuri/service/$serviceid/profile?env=$env" | awk -v FS="Request-Id: " 'NF>1{print $2}')
CP_requestid=${CP_requestid%$'\r'}
if [ -z "$CP_requestid" ]; then
	printf "\n"
	log "Error: Request ID invalid, exit 1."
	exit 1
fi
log "\t1.a) Server acknowledged with Request ID = $CP_requestid"

# STEP 2. RETRIEVE Service credentials #
log "2. Using Request ID = $CP_requestid and Service ID = $serviceid To Retrieve Service credentials"
next_wait_time=0
while [ $next_wait_time -lt 30 ]; do
	statuscode=$(curl --silent --output /dev/null --write-out %{http_code} \
		"$apiserverbaseuri/service/$serviceid/profile?requestid=$CP_requestid&env=$env")
	if [ "$statuscode" -eq 200 ]; then
		# connection profile package is ready, retrieve from server and add to package.tar
		curl --silent "$apiserverbaseuri/service/$serviceid/profile?requestid=$CP_requestid&env=$env" >| "$HOME/results/tmp/package.tar"
		tar xf $HOME/results/tmp/package.tar -C $HOME/results/
		cleanOldCPs
		mv $HOME/results/workdir/results $HOME/results/creds
		printf "\n"
		break
	elif [ "$statuscode" -eq 404 ] || [ "$statuscode" -eq 202 ]; then
		printf "."
		sleep $((next_wait_time++))
	else
		printf "\n"
		log "Error: unable to retrieve service credentials."
		exit 2
	fi
done
log "\t2.a)Service credentials retrieved."