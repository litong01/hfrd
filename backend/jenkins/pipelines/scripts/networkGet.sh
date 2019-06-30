#!/bin/bash
# The script to query fabric network
# This file will require the following parameters
# endpoint, apikey, org, space, service, serviceplan and serviceid
set -o pipefail
mkdir -p $WORKDIR/results

if [[ $serviceid =~ [a-z0-9]+-[0-9] ]]; then
	# Enterprise Plan
	echo "Get service key for EP with service: $serviceid"
	originServiceId=${serviceid}
	numOfOrgs=${serviceid#*-}
	serviceid=${serviceid%-*}
	reNum='^[0-9]+$'  # numOfOrgs should be number
	if [ -z $numOfOrgs ] || [ -z $serviceid ] || ! [[ $numOfOrgs =~ $reNum ]]; then
		echo "invalid serviceid: $originServiceId"
		exit 1
	fi
	for((i = 0; i<$numOfOrgs;i++))
	do
    	serviceids[$i]=$serviceid'-'$i
	done
	echo '{}' > network.json
	bx login -a $endpoint --apikey $apikey -o $org -s $space
	for id in ${serviceids[@]}
    do
        bx cf service-key $id $id"-key" | tail -n +5  > temp.json
        if [ $? -ne 0 ]; then
        	echo "failed to get service-key for $id"
        	exit 1
        fi
        networkObject=$(jq -s '.[0] * .[1]' network.json temp.json)
        echo $networkObject > network.json
    done
    rm -f temp.json
	mv network.json $WORKDIR/results/network.json
	echo "network.json:"
	jq . $WORKDIR/results/network.json
	echo '{' > $WORKDIR/results/service.json
	  echo '  "serviceid": "'$originServiceId'"' >> $WORKDIR/results/service.json
	echo '}' >> $WORKDIR/results/service.json
else
	# Starter plan
	echo "Get service key for SP with service: $serviceid"
	bx login -a $endpoint --apikey $apikey -o $org -s $space
	bx cf service-key $serviceid $servicekey | tail -n +5 > $WORKDIR/results/network.json
	if [ $? -ne 0 ]; then
		echo "Unable to get service key for $serviceid"
		rm $WORKDIR/results/network.json
		exit 1
	fi
	echo '{' > $WORKDIR/results/service.json
	  echo '  "serviceid": "'$serviceid'"' >> $WORKDIR/results/service.json
	echo '}' >> $WORKDIR/results/service.json
fi