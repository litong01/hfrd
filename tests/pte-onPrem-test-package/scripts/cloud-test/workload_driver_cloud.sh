#!/bin/bash

#   Worload Driver
#   Description: Used to drive traffic to blockchain network
#   Dependencies: Path of PTE SCFile
#   TODO: This should be able
#   One parameter :
#		1)path of PTE SCFile
. utils.sh

MAX_RETRY=5
PROG="[hfrd-pte-test-cloud]"

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
	log "-n | --network  :   The path of network.json which contains all of the service credentials in blockchain network,including msp_id,networkId,API key/secret"
	log "-c | --config  :   The path of hfrd_test.cfg "
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
	if [[ -z "$workload" || -z "$networkPath" || -z "$configPath" ]];then
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
elif [[ ! -f $configPath ]]; then
	log "Missing hfrd test configuration file, cannot continue."
	exit 1
fi

source $configPath

msp_id=$(jq -r keys[0] $networkPath)
networkId=$(jq -r .$msp_id.network_id $networkPath)
if [[ ! -f ${HOME}/results/SCFiles/config-net-${networkId}.json ]]; then
	log "Missing SCFile,must get the PTE SCFile before continue."
	exit 1
fi

# STEP 1. Package the PTE Test
log "1. Package PTE Test"
mkdir -p $HOME/pte-fab
cp -f $HOME/hfrd_test.cfg $HOME/pte-fab/hfrd_test.cfg
cp -f $HOME/scripts/cloud-test/test-entrypoint-jks.sh  $HOME/pte-fab/test-entrypoint-jks.sh
cp -f $HOME/scripts/cloud-test/docker-entrypoint-jks.sh  $HOME/pte-fab/docker-entrypoint-jks.sh
cp -f $HOME/scripts/cloud-test/Dockerfile-jks $HOME/pte-fab/Dockerfile-jks
cp -rf $HOME/results/creds $HOME/pte-fab/creds
cp -rf $HOME/workloads/$workload $HOME/pte-fab/userInputs
cp -f $HOME/results/SCFiles/config-net-${networkId}.json $HOME/pte-fab/config-chan1-TLS.json

if [[ $env == 'cm' ]]; then
	cp -f $HOME/conf/hosts $HOME/pte-fab/hosts
fi
cd $HOME/pte-fab
tar zcfP /tmp/${networkId}.tar.gz test-entrypoint-jks.sh docker-entrypoint-jks.sh Dockerfile-jks config-chan1-TLS.json creds userInputs
if [ $? -ne 0 ]; then
	log "Package pte tests failed.Will exit 2"
	exit 2
fi

MD5SUM=$(md5sum /tmp/${networkId}.tar.gz | cut -f1 -d " ")
log "\t5.a) PTE Test Package created in /tmp/${networkId}.tar.gz"
log "\t5.b) MD5 Sum = $MD5SUM"

# STEP 2. UPLOAD PTE TEST PACKAGE TO HTTP SERVER #
log "2. Upload PTE Test Package"
printf "Upload /tmp/${networkId}.tar.gz to an HTTP server so that the test runner can retrieve it.\n"
sshpass -p $packageServerSecret scp -o StrictHostKeyChecking=no /tmp/${networkId}.tar.gz $packageServerUser@$packageServerHost:$packageDir
if [ $? -ne 0 ]; then
	log "Upload PTE Test Package failed.Will exit 3"
	exit 3
fi
PTE_URL=$testPackageServer"/${networkId}.tar.gz"
# STEP 3. START THE PTE TEST RUN
log "3. Starting PTE Test Run"
## TODO: retrieve test_requestid from the following request
test_requestid=$(curl -X POST --silent --include "$apiserverbaseuri/test" \
	-d '{"url":"'$PTE_URL'","hash":"'$MD5SUM'","startcmd":"test-entrypoint-jks.sh"}' | awk -v FS="Request-Id: " 'NF>1 {print $2}')
test_requestid=${test_requestid%$'\r'}
if [ -z $test_requestid ]; then
	printf "Error getting request id for test.Will exit 4"
	exit 4
else
	log "\t3.a) Point your browser to $apiserverbaseuri/test?requestid=$test_requestid to observe your run."
	sleep 5
fi


