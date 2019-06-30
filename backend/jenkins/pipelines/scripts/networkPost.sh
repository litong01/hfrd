#!/bin/bash
# The script to create fabric network
# This file will require the following parameters
# endpoint, apikey, org, space, service, serviceplan and serviceid
# This script produces a network.json file being placed in $WORKDIR/results
# directory
set -o pipefail
MAX_RETRY=5
COUNTER=0
clean() {
	if [[ $env == *-ep ]]; then
		# Enterprise Plan
		echo "cleaning up Enterprise Plan services"
		for id in ${serviceids[@]}
		do
			bx cf dsk $id $id"-key" -f
			bx cf ds $id -f
			if [ $? -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
				COUNTER=` expr $COUNTER + 1`
				sleep 5
				clean
			fi
		done
	else
		# Starter Plan
		echo "cleaning up Starter Plan service"
		rm $WORKDIR/results/network.json
		bx cf dsk $serviceid $servicekey -f
		bx cf ds $serviceid -f
		if [ $? -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
			COUNTER=` expr $COUNTER + 1`
			sleep 5
			clean
		fi
	fi
}

runWithRetry(){
	retrytime=0
	funcName=$1
	if [ -z $funcName ]; then
			echo "Pass the func name as parameter; runWithRetry funcName [params for funcName]"
			break
	fi
   	shift
   	params=$@
	while [ $retrytime -le $MAX_RETRY ]; do
    	((retrytime++))
		# Use first parameter as the function
		echo "Executing $funcName with retry: $retrytime, MAX_RETRY=$MAX_RETRY"
		$funcName $params
    	if [ $? -ne 0 ]; then
        	if [ $MAX_RETRY -eq $retrytime ]; then
       			echo "Already reach the maximum number of attempts.Still failed to finish the job"
       			if [ $funcName == "clean" ]; then
       				echo "clean failure... Need to manually clean up this service"
       				exit 1
       			else
       				clean
       				exit 1
       			fi
    		else
            	echo "$funcName failed,will retry"
				sleep 2s
        	fi
    	else
    		echo "$funcName executed successfully"
        	break
    	fi
	done
}

# parameters: networkId, email, company_name, key, token, heliosUrl
sendInvitation() {
	networkId=$1
	email=$2
	companyName=$3
	key=$4
	secret=$5
	heliosUrl=$6
	echo "[sendInvitation] params: $@"
	if [ -z $networkId ] || [ -z $email ] || [ -z $companyName ] || [ -z $key ] || [ -z $secret ] || [ -z $heliosUrl ];
	then
		echo "[sendInvitation] Invalid parameters. Should be networkId, email, company_name, key, secret, heliosUrl"
		return 1
	fi
	BODY=$(cat <<EOF
		{
  			"invites": [
			{
				"email": "${email}",
				"company_name": "${companyName}"
			}
  		]
}
EOF
)
	invitationUrl="${heliosUrl}/api/v1/networks/${networkId}/invite"
	result=$(curl -X POST \
		--header 'Content-Type: application/json' \
		--header 'Accept: application/json' \
		--basic --user ${key}:${secret} \
		--data "${BODY}" \
		${invitationUrl})
	echo "[sendInvitation] invitationUrl: $invitationUrl"
	echo "[sendInvitation] result:$result"
	if [[ $result ==  *invited* ]]; then
		return 0
	else
		return 1
	fi
}

# parameters: networkId, email, company_name, service_id, heliosUrl
joinNetwork() {
	networkId=$1
	email=$2
	companyName=$3
	serviceId=$4
	heliosUrl=$5
	echo "[joinNetwork] params: $@"
	if [ -z $networkId ] || [ -z $email ] || [ -z $companyName ] || [ -z $serviceId ] || [ -z $heliosUrl ];
	then
		echo "Invalid parameters. Should be networkId, email, company_name, service_id"
		return 1
	fi
	bx cf cs $service $serviceplan $serviceId
	bx cf create-service-key $serviceId $serviceId"-key"
	bx cf service-key $serviceId $serviceId"-key" | tail -n +5 > $WORKDIR/results/network-${serviceId}.json
	if [ $? -ne 0 ]; then
	# create service failed
		bx cf dsk $serviceId $serviceId"-key" -f
		bx cf ds $serviceId -f
		return 1
	fi
	id=$(jq -r .service_instance_id $WORKDIR/results/network-${serviceId}.json)
	token=$(jq -r .service_instance_token $WORKDIR/results/network-${serviceId}.json)
	BODY=$(cat <<EOF
	{
	  "company_name": "${companyName}",
	  "email": "${email}",
	  "peers": ${numOfPeers}
	}
EOF
)
	joinUrl="${heliosUrl}/api/v1/networks/"${networkId}"/join"
	statusCode=$(curl -X POST --silent --output /dev/null --write-out %{http_code} \
		--header 'Content-Type: application/json' \
		--header 'Accept: application/json' \
		--basic --user ${id}:${token} \
		--data "${BODY}" \
		${joinUrl})
	echo "[joinNetwork] joinUrl: ${joinUrl}; id: $id, token: $token"
	echo "[joinNetwork] status code: ${statusCode}"
	if [ $statusCode -eq 200 ]; then
		# regenerate service key
		runWithRetry bx cf dsk $serviceId $serviceId"-key" -f
		runWithRetry bx cf csk $serviceId $serviceId"-key"
		bx cf service-key $serviceId $serviceId"-key" | tail -n +5 > $WORKDIR/results/network-${serviceId}.json
		if [ $? -eq 0 ]; then
			return 0
		fi
	fi
	rm $WORKDIR/results/network-${serviceId}.json
	bx cf dsk $serviceId $serviceId"-key" -f
	bx cf ds $serviceId -f
	return 1
}

if [ -z $serviceid ]; then
	echo "no serviceid provided..."
	exit 1
fi

mkdir -p $WORKDIR/results
bx config --check-version=false
bx login -a $endpoint --apikey $apikey -o $org -s $space
if [ $? -ne 0 ]; then
	echo "bx login failed"
	exit 1
fi

# Bluemix Enterprise Plan
if [[ $env == *-ep ]]; then # specific to enterprise plan
	echo "Creating EP network env:$env, loc:$loc, numOfOrgs:$numOfOrgs, numOfPeers:$numOfPeers, ledgerType:$ledgerType"
	# Basic params check
	if [ $numOfOrgs -gt 2 ]; then
		echo "Unsupported numOfOrgs:$numOfOrgs. We only support no more than 2 Orgs for Enterprise Plan for now"
		exit 1
	fi
	if [ $numOfPeers -gt 3 ]; then
		echo "Unsupported numOfPeers:$numOfPeers. Should be no more than 3"
		exit 1
	fi
	# service instance id modified for Enterprise plan
	for ((i = 0; i<$numOfOrgs; i++))
	do
   		serviceids[$i]=$serviceid'-'$i
	done
	# emails used to join the network
	emails=("blkchnst@us.ibm.com" "xixuejia@cn.ibm.com" "sunhwei@cn.ibm.com")

	# claim a blockchain EP network with blkchnst@us.ibm.com
	runWithRetry bx cf cs $service $serviceplan ${serviceids[0]}
	runWithRetry bx cf create-service-key ${serviceids[0]} ${serviceids[0]}"-key"
	bx cf service-key ${serviceids[0]} ${serviceids[0]}"-key" | tail -n +5 > $WORKDIR/results/network-${serviceids[0]}.json
	if [ $? -ne 0 ]; then
		clean
		exit 1
	fi
	service_instance_id=$(jq -r .service_instance_id $WORKDIR/results/network-${serviceids[0]}.json)
	service_instance_token=$(jq -r .service_instance_token $WORKDIR/results/network-${serviceids[0]}.json)
	echo "service_instance_id: $service_instance_id"
	echo "service_instance_token: $service_instance_token"
	BODY=$(cat <<EOF
		{
		  "location_id": "${loc}",
		  "company_name": "IBM-0",
		  "email": "${emails[0]}",
		  "peers": ${numOfPeers},
		  "ledger_type": "${ledgerType}"
		}
EOF
	)
	# get available locations
	if [[ $env == bxstaging* ]]; then
		url="https://ibmblockchain-dev-v2.stage1.ng.bluemix.net/api/v1/network-locations/available"
		devHeliosBaseUrl="https://ibmblockchain-dev-v2.stage1.ng.bluemix.net"
	elif [[ $env == bxproduction* ]]; then
		url="https://ibmblockchain-v2.ng.bluemix.net/api/v1/network-locations/available"
		devHeliosBaseUrl="https://ibmblockchain-v2.ng.bluemix.net"
	fi
	echo "Checking whether the location id:${loc} is available"
	result=$(curl ${url})
	url=$(echo ${result} | jq -r .\"$loc\".swagger_url)
	baseurl=$(echo $url | sed 's/\(https:\/\/[^ /]*\).*/\1/g')
	if [[ $baseurl == https://* ]]; then
		echo -e "loc:$loc\nbase url: $baseurl"
	else
		echo "Unavailable location $loc"
		clean
		exit 1
	fi
	# claim network
	claimNetwork(){
	    RESULT=$(curl -X POST --header 'Content-Type: application/json' \
		--header 'Accept: application/json' \
		--basic --user ${service_instance_id}:${service_instance_token} \
		--data "${BODY}" \
		${devHeliosBaseUrl}/api/v1/networks)

		if [[ $RESULT != *\"network_id\":* ]]
		then
		    echo "Claim network error: $RESULT"
            return 1
        fi
        return 0
	}
	runWithRetry claimNetwork
	echo "create network result: $RESULT"
	# Check result
	if [[ $RESULT != *\"network_id\":* ]]
	then
		clean
		exit 1
	fi
	# recreate service key
	runWithRetry bx cf dsk ${serviceids[0]} ${serviceids[0]}"-key" -f
	runWithRetry bx cf csk ${serviceids[0]} ${serviceids[0]}"-key"
	bx cf service-key ${serviceids[0]} ${serviceids[0]}"-key" | tail -n +5 > $WORKDIR/results/network-${serviceids[0]}.json
	if [ $? -ne 0 ]; then
		clean
		exit 1
	fi
	# Enterprise plan hardcoded the first org's msp id to 'PeerOrg1'
	networkId=$(jq -r .PeerOrg1.network_id $WORKDIR/results/network-${serviceids[0]}.json)
	if [ -z $networkId ] || [ $networkId == null ]; then
		echo "empty networkid"
		clean
		exit 1
	fi
	key=$(jq -r .PeerOrg1.key $WORKDIR/results/network-${serviceids[0]}.json)
	secret=$(jq -r .PeerOrg1.secret $WORKDIR/results/network-${serviceids[0]}.json)
	# heliosUrl=$(jq -r .PeerOrg1.url $WORKDIR/results/network-${serviceids[0]}.json)
	### Invite other orgs and join the network
	for ((i = 1; i<$numOfOrgs; i++))
	do
		# sendInvitation params: networkId, email, company_name, key, token
		runWithRetry 'sendInvitation' $networkId ${emails[i]} "IBM-$i" $key $secret $devHeliosBaseUrl
		# parameters: networkId, email, company_name, service_id
		runWithRetry 'joinNetwork' $networkId ${emails[i]} "IBM-$i" ${serviceids[i]} $devHeliosBaseUrl
	done
	# successfully created network and joined orgs now
	echo '{}' > $WORKDIR/results/network.json
	for ((i = 0; i<$numOfOrgs; i++))
	do
		networkObject=$(jq -s '.[0] * .[1]' $WORKDIR/results/network.json $WORKDIR/results/network-${serviceids[i]}.json)
        echo $networkObject > $WORKDIR/results/network.json
        rm $WORKDIR/results/network-${serviceids[i]}.json
	done

# Bluemix Starter Plan
else
	echo "Creating Starter plan in env: $env"
	bx cf cs $service $serviceplan $serviceid
	bx cf create-service-key $serviceid $servicekey
	bx cf service-key $serviceid $servicekey | tail -n +5 > $WORKDIR/results/network.json
	if [ $? -ne 0 ]; then
		clean
		exit 1
	fi
fi

echo '{' > $WORKDIR/results/service.json
if [[ $env == *-ep ]]; then
	echo '  "serviceid": "'$serviceid-$numOfOrgs'"' >> $WORKDIR/results/service.json
else
	echo '  "serviceid": "'$serviceid'"' >> $WORKDIR/results/service.json
fi
echo '}' >> $WORKDIR/results/service.json
pwd
cat $WORKDIR/results/network.json
echo '........starting generating certs and other files ......'
python /opt/src/scripts/zipgen.py -n $WORKDIR/results/network.json  -o $WORKDIR/results
python /opt/src/scripts/uploadcert.py -d $WORKDIR/results/keyfiles 
rm -rf $WORKDIR/results/keyfiles
