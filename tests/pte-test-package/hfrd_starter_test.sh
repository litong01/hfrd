#!/usr/bin/env bash
#. slack.sh

###############################################################################
#                                                                             #
#                              HFRD-STARTER-TEST                              #
#                                                                             #
###############################################################################

# Parse CLI args
DELETE_SERVICE=false
POSITIONAL=()
while [[ $# -gt 0 ]]; do
	key="$1"
	case $key in
	-d | --delete)
		DELETE_SERVICE=true
		shift # past argument
		shift # past value
		;;
	*) # unknown option
		POSITIONAL+=("$1") # save it in an array for later
		shift              # past argument
		;;
	esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

# cleanup previous runs
rm -rf logs
mkdir ./logs
rm -rf tmp
mkdir ./tmp

# environment
source './hfrd_starter_test.cfg'
DATESTR=$(date +%Y-%m-%d" "%H:%M:%S" "%p)
PROG="[hfrd-starter-test]"
LOGFILE="$(pwd)/logs/hfrd_starter_test.log"

# Helper functions
function log() {
	printf "$PROG\t$1\n" | tee -a $LOGFILE
}

# STEP 0. API SERVER ENDPOINT DETAILS & SETUP #
log "$DATESTR"
log "0. Starting HFRD API STARTER TEST Script"

# STEP 1. REQUEST A STARTER PLAN NETWORK FROM IBP #
log "1. Sending Request To Provision A Starter Network"
# capture the request id in the response headers 
requestid=$(curl --silent --include -X POST \
	"$apiserverbaseuri/service" -d "{\"env\": \"$env\"}" | awk -v FS="Request-Id: " 'NF>1 {print $2}')
requestid=${requestid%$'\r'}
if [ -z "$requestid" ]; then
	log "Error: Request ID invalid, exit 1."
	exit 1
fi
log "\t1.a) Server acknowledged with Request ID = $requestid"

# STEP 2. RETRIEVE STARTER PLAN INSTANCE SERVICE KEY #
log "2. Using Request ID = $requestid To Retrieve Service Credentials"
nextwaittime=0
while [ $nextwaittime -le 30 ]; do
	statuscode=$(curl --silent --output /dev/null --write-out %{http_code} \
		"$apiserverbaseuri/service?requestid=$requestid&env=$env")
	if [ "$statuscode" -eq 200 ]; then
		# key package is ready, retrieve it from the server
		curl --silent "$apiserverbaseuri/service?requestid=$requestid&env=$env" >| "./tmp/package.tar"
		# extract service id from the package
		serviceid=$(tar xfO "./tmp/package.tar" "workdir/results/service.json" | jq -r '.serviceid')
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
log "\t2.a) Retrieved service key package, Service ID = $serviceid"

# STEP 3. REQUEST CONNECTION PROFILE #
log "3. Using Service ID = $serviceid To Request Connection Profile"
CP_requestid=$(curl --silent --include \
	"$apiserverbaseuri/service/$serviceid/profile?env=$env" | awk -v FS="Request-Id: " 'NF>1{print $2}')
CP_requestid=${CP_requestid%$'\r'}
if [ -z "$CP_requestid" ]; then
	printf "\n"
	log "Error: Request ID invalid, exit 3."
	exit 3
fi
log "\t3.a) Server acknowledged with Request ID = $CP_requestid"

# STEP 4. RETRIEVE CONNECTION PROFILE #
log "4. Using Request ID = $CP_requestid and Service ID = $serviceid To Retrieve Connection Profile"
next_wait_time=0
while [ $next_wait_time -lt 30 ]; do
	statuscode=$(curl --silent --output /dev/null --write-out %{http_code} \
		"$apiserverbaseuri/service/$serviceid/profile?requestid=$CP_requestid&env=$env")
	if [ "$statuscode" -eq 200 ]; then
		# connection profile package is ready, retrieve from server and add to package.tar
		curl --silent "$apiserverbaseuri/service/$serviceid/profile?requestid=$CP_requestid&env=$env" >| "./tmp/package.tar"
		networkId=$(tar xfO "./tmp/package.tar" "workdir/results/ConnectionProfile_org1.json" | jq -r '.["x-networkId"]')
		networkId=${networkId%$'\r'}
		printf "\n"
		break
	elif [ "$statuscode" -eq 404 ] || [ "$statuscode" -eq 202 ]; then
		printf "."
		sleep $((next_wait_time++))
	else
		printf "\n"
		log "Error: unable to retrieve connection profile, exit 4."
		exit 4
	fi
done
log "\t4.a)Connection Profile retrieved. Network ID = $networkId"

# STEP 5. PREPARE PTE TEST PACKAGE TO TARGET REMOTE NETWORK #
log "5. Prepare PTE Test Package"
if [ ! -d "../pte-test-package" ]; then
	log "Error: Cannot locate PTE Test Package Directory, exit 5."
	exit 5
fi

if [ ! -d "../pte-test-package/creds" ]; then
        log "Error: Creds directory does not exist. creating creds directory"
        mkdir creds
fi
tar xf tmp/package.tar
cp workdir/results/ConnectionProfile_org1.json ../pte-test-package/creds/org1ConnectionProfile.json
cp workdir/results/ConnectionProfile_org2.json ../pte-test-package/creds/org2ConnectionProfile.json
cp workdir/results/network.json ../pte-test-package/creds/apikeys.json
cd ../pte-test-package
tar zcf /tmp/pte.tar.gz docker-entrypoint.sh userInputs creds scripts test-entrypoint.sh Dockerfile
MD5SUM=$(md5sum /tmp/pte.tar.gz | cut -f1 -d " ")
log "\t5.a) PTE Test Package created in /tmp/pte.tar.gz"
log "\t5.b) MD5 Sum = $MD5SUM"

# STEP 6. UPLOAD PTE TEST PACKAGE TO HTTP SERVER #
log "6. Upload PTE Test Package"
printf "Upload /tmp/pte.tar.gz to an HTTP server so that the test runner can retrieve it.\n"
printf "When this is complete, enter the URL to the remote pte.tar.gz here:\n"
read -r PTE_URL

# STEP 7. START THE PTE TEST RUN
log "7. Starting PTE Test Run"
## TODO: retrieve test_requestid from the following request
test_requestid=$(curl -X POST --silent --include "$apiserverbaseuri/test" \
	-d '{"url":"'$PTE_URL'","hash":"'$MD5SUM'","startcmd":"test-entrypoint.sh"}' | awk -v FS="Request-Id: " 'NF>1 {print $2}')
test_requestid=${test_requestid%$'\r'}
if [ -z $test_requestid ]; then
	printf "Error getting request id for test"
else
	log "\t7.a) Point your browser to $apiserverbaseuri/test?requestid=$test_requestid to observe your run."
	sleep 5
fi

# STEP 99. DELETE REMOTE NETWORK SERVICE #
if [ "$DELETE_SERVICE" = true ]; then
	log "99. It looks like you wish to delete service $serviceid."
	read -p "Are you sure? Enter Y to continue " -n 1 -r
	if [[ $REPLY =~ ^[Yy]$ ]]; then
		log "\t99.a) Using Service ID = $serviceid to Delete the Remote Service\n"
		retry=0
		while [ $retry -lt 5 ]; do
			del_requestid=$(curl -X DELETE --silent --include \
				"$apiserverbaseuri/service/$serviceid?env=$env" | awk -v FS="Request-Id: " 'NF>1{print $2}')
				del_requestid=${del_requestid%$'\r'}
			log "deleting service returned request id: $del_requestid"
			next_wait_time=0
			while [ $next_wait_time -lt 30 ] || [ "$statuscode" -eq 202 ]; do
				statuscode=$(curl --silent --output /dev/null --write-out %{http_code} \
					"$apiserverbaseuri/service?requestid=$del_requestid&env=$env")
				if [ "$statuscode" -eq 200 ]; then
					printf "\n"
					log "\t99.a) Successfully deleted service"
					exit 0
				elif [ "$statuscode" -eq 404 ] || [ "$statuscode" -eq 202 ]; then
					printf "."
					sleep $((next_wait_time++))
				else
					printf "\n"
					break
				fi
			done
			log "Round $retry failed to delete service"
			((++retry))
		done
	else
		log "\t99.a) Aborting delete service, have a nice day!"
		exit 0
	fi
	log "\t99.a) Delete service FAILED!!!!!!"
	exit 1
fi
exit 0
