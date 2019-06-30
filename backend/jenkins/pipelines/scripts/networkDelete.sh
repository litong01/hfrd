#!/bin/bash
# The script to delete fabric network
# This file will require the following parameters
# endpoint, apikey, org, space, service, serviceplan and serviceid
mkdir -p $WORKDIR/results
set -o pipefail
MAX_RETRY=5
COUNTER=0
clean() {
	id=$1
	echo "cleaning up Bluemix IBP service $id"
	bx cf dsk $id $id"-key" -f
	bx cf ds $id -f
	ret=$?
	echo "return code of delete service $id: $ret"
	if [ $ret -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
		COUNTER=` expr $COUNTER + 1`
		echo "Delete service $id failed: $COUNTER, will retry in 5s"
		sleep 5
		clean $id
		return $?
	elif [ $ret -eq 0 ]; then
		echo "successfully deleted service: $id"
		COUNTER=0
		return 0
	else
		COUNTER=0
		return 1
	fi
}

if [[ $serviceid =~ [a-z0-9]+-[0-9] ]]; then
	# Enterprise Plan service
	echo "deleting EP service: $serviceid"
	originServiceId=${serviceid}
	numOfOrgs=$(echo $serviceid | awk 'match($0, /[0-9]+$/) {print substr($0, RSTART, RLENGTH)}')
	serviceid=$(echo $serviceid | awk 'match($0, /.*-/) {print substr($0, RSTART, RLENGTH-1)}')
	echo "service name prefix: $serviceid; numOfOrgs: $numOfOrgs"
	reNum='^[0-9]+$'  # numOfOrgs should be number
	if [ -z $numOfOrgs ] || [ -z $serviceid ] || ! [[ $numOfOrgs =~ $reNum ]]; then
		echo "invalid serviceid: $originServiceId"
		exit 1
	fi
	for ((i = 0; i<$numOfOrgs; i++))
	do
		serviceids[$i]=$serviceid'-'$i
	done
else
	# Starter Plan
	echo "deleting SP service: $serviceid"
	serviceids[0]=$serviceid
fi
bx login -a $endpoint --apikey $apikey -o $org -s $space
for id in ${serviceids[@]}
do
	clean $id
	if [ $? -ne 0 ]; then
		echo "delete $id failed"
		exit 1
	fi
done