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

. utils.sh

log() {
	printf "${PROG}  ${1}\n" | tee -a ${dir_networks}logs/network_generator.log
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
		echo "ERROR: No hfrd configuration file supplied."
		Print_Help
		exit 1
	fi

}

createRequiredDirs(){
	if [[ ! -f ${results_dir}tmp ]];then
		mkdir -p ${results_dir}tmp
	fi
	if [[ ! -f ${results_dir}logs ]];then
		mkdir -p ${results_dir}logs
	fi
	if [[ ! -f ${results_dir}SCFiles ]];then
		mkdir -p ${results_dir}SCFiles
	fi
}
buildNetworkObject(){
	eval env=${network__env[${index_type}]}
	eval name=${network__planName[${index_type}]}
	if [[ $name == 'ep' ]];then
		eval loc=${network__loc[${index_type}]}
		eval numOfOrgs=${network__numOfOrgs[${index_type}]}
		eval numOfPeers=${network__numOfPeers[${index_type}]}
		eval ledgerType=${network__ledgerType[${index_type}]}
		REQUEST_BODY="{\"env\": \"$env\",\"loc\": \"$loc\",\"name\":\"$name\",\"config\":{\"numOfOrgs\": $numOfOrgs,\"numOfPeers\": $numOfPeers,\"ledgerType\": \"$ledgerType\"}}"
	else
		REQUEST_BODY="{\"env\": \"$env\"}"
	fi
}

requestNetwork(){
	requestid=$(curl --silent --include -X POST \
	"$apiserverbaseuri/service" -d "$REQUEST_BODY" | awk -v FS="Request-Id: " 'NF>1 {print $2}')
	requestid=${requestid%$'\r'}
	if [ -z "$requestid" ]; then
		log "Error: Request ID invalid, exit 1."
		return 1
	fi
	log "Server acknowledged with Request ID = $requestid"
}

retrieveServiceCredentials(){
	log "Retrieve service creadentials by using requestID $requestid"
	nextwaittime=0
	while [ $nextwaittime -le 100 ]; do
		statuscode=$(curl --silent --output /dev/null --write-out %{http_code} \
			"$apiserverbaseuri/service?requestid=$requestid&env=$env")
		if [ "$statuscode" -eq 200 ]; then
			# key package is ready, retrieve it from the server
			createRequiredDirs
			curl --silent "$apiserverbaseuri/service?requestid=$requestid&env=$env" >| "${results_dir}tmp/package.tar"
			tar xf ${results_dir}tmp/package.tar -C $results_dir
			mv ${results_dir}workdir/results ${results_dir}creds
			# extract service id from the package
			serviceid=$(jq -r '.serviceid' ${results_dir}creds/service.json)
			serviceid=${serviceid%$'\r'}
			mv ${results_dir} ${dir_networks}${type}"/"$serviceid
			printf "\n"
			break
		elif [ "$statuscode" -eq 404 ] || [ "$statuscode" -eq 202 ]; then
			printf "."
			sleep $((nextwaittime++))
		else
			printf "\n"
			log "Error: Unable to retrieve service credentials, exit 2."
			return 2
		fi
	done
	if [ -z "$serviceid" ]; then
		printf "\n"
		log "Error: Service Credentials invalid, exit 2."
		return 2
	fi
	log "Retrieved  serviceid, Service ID = $serviceid"
}

Parse_Arguments $@

# Sanity check
if [[ ! -f $configPath ]]; then
	echo "Missing hfrd configuration file, cannot continue."
	exit 1
fi

# Load the configurations
create_variables $configPath

DATESTR=$(date +%Y-%m-%d" "%H:%M:%S" "%p)
PROG="[hfrd-network-generator]"
log "$DATESTR"

apiserverbaseuri=${apiserver_apiserverhost}/${apiserver_apiversion}/${apiserver_apiuser}
# Create multiple networks based on the configurations
index_type=0
for type in ${network__type[@]}; do
	eval numberOfNetworks=${network__numberOfNetworks[${index_type}]}
	buildNetworkObject
	log "Create ${numberOfNetworks} ${type} networks"
	succ=0
	fail=0
	for index in $(seq 1 ${numberOfNetworks})
	do
		log "Create ${type} network: ${index}"
		requestNetwork
		if [ $? -ne 0 ]; then
			((fail++))
		else
			results_dir=${dir_networks}${type}"/network${index}/"
			retrieveServiceCredentials
			if [ $? -ne 0 ]; then
				((fail++))
			else
				((succ++))
			fi
		fi
	done
	log "Done on $type network creation: ${succ} succeeded , ${fail} failed"
	log " "
	((index_type++))
done

