#!/bin/bash

get_access_token(){
    echo "Getting access token for cluster manager"
    payload='grant_type=urn%3Aibm%3Aparams%3Aoauth%3Agrant-type%3Aapikey&apikey='$apikey
    response=$(curl -k -s -H "Accept:application/json" \
                        -H "Cache-control:no-cache"   \
                        -H "Content-type:application/x-www-form-urlencoded" \
                        -d "$payload" \
                        -X POST \
                        $AUTH_URL
                )
    access_token=$(echo $response | jq -r .access_token)
    if [[ "$access_token" == null ]]; then
        return 1
    else
        return 0
    fi
}

createNetwork(){
    #Due to the network is created from scratch,Mysql, zk, kafka, peer, orderer, etc…dozens of docker containers..
    #This will take over 10 minutes
    echo "Create Networks on $loc... Have a cup of coffee.. This will take over 10 minutes..."
    statuscode=$(curl -k -o /dev/null -s -w %{http_code} -H "Accept:application/json" \
                        -H "Authorization:$access_token"   \
                        -H "Content-type:application/json" \
                        -X POST \
                        $NET_CREATE
            )
    if [ "$statuscode" -eq 200 ]; then
        return 0
    else
        return 1
    fi
}

claimeNetwork(){
    echo "Claim network"
    network_input=$(jq .network $NET_INPUT | sed -e 's@{serviceid}@'"${serviceids[0]}"'@g')
    response=$(curl -k -H "Accept:application/json" \
                        -H "Authorization:$access_token"   \
                        -H "Content-type:application/json" \
                        -d "$network_input" \
                        -X POST \
                        $NET_CLAIME
        )
    echo $response | jq -r .message._id
    NetworkID=$(echo $response | jq -r .message._id)
    if [[ "$NetworkID" == null ]]; then
        return 1
    else
        return 0
    fi
}

deleteNetwork(){
    echo "Delete network $networkID for service:${serviceid}-${numOfOrgs} "
    statuscode=$(curl -k -o /dev/null -s -w %{http_code} -H "Accept:application/json" \
                        -H "Authorization:$access_token"   \
                        -H "Content-type:application/json" \
                        -X DELETE \
                        $NET_DELETE
            )
    if [ "$statuscode" -eq 200 ]; then
        echo "Successfully deleted network for service ${serviceid}-${numOfOrgs}"
        return 0
    else
        echo "Failed to delete network $serviceid-$numOfOrgs"
        return 1
    fi
}

createPeer(){
    echo "Create peer$peer for service : $id"
    statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
                    -H "Accept:application/json" \
                    -H "Authorization:$access_token"   \
                    -H "Content-type:application/json" \
                    -d "$peer_object" \
                    -X POST \
                    $NET_PEER_CREATE
            )
        if [ "$statuscode" -eq 200 ]; then
            return 0
        else
            return 1
        fi
}

updateIBMIdDoc(){
	echo "Update ibmid_doc for network: $NetworkID"
	statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
                    -X POST \
					-H 'Content-Type: application/json' \
					-H 'Accept: application/json' \
					-u ${ORG_API_KEY}:${ORG_API_SECRET} \
					${API_ENDPOINT}/api/v1/networks/${NetworkID}/getIBMIdDoc
    	)
    if [ "$statuscode" -eq 200 ]; then
        return 0
    else
		return 1
    fi
}

joinNetwork(){
    echo "Join service $id into network :$NetworkID"
    statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
                    -H "Content-Type:application/json" \
                    -u $heliosUser:$heliosSecret \
                    -X POST \
                    --data "{\"service_id\":\"$id\"}" \
                    ${heliosUrl}${NET_ORG_JOIN}
    )
    if [ "$statuscode" -eq 200 ]; then
            return 0
    else
            return 1
    fi
}

getServiceCredentials(){
    echo "get service credentials"
    echo '{}' > network.json
    for id in ${serviceids[@]}
    do
        curl -k -H "Content-Type:application/json" \
                            -u $heliosUser:$heliosSecret \
                            -X GET \
                            ${GET_CREDS_API}$id  > temp.json
        networkObject=$(jq -s '.[0] * .[1]' network.json temp.json)
        echo $networkObject > network.json
        sleep 0.5s
    done
}

runWithRetry(){
	retrytime=0
	while [ $retrytime -le $MAX_RETRY ]; do
    	((retrytime++))
		# Use first parameter as the function
    	$1
    	if [ $? -ne 0 ]; then
        	if [ $MAX_RETRY -eq $retrytime ]; then
       			echo "Already reach the maximum number of attempts.Still failed to finish the job"
				sleep 2s
        		exit 1
    		else
            	echo "Job failed,will retry"
				sleep 2s
        	fi
    	else
    		echo "Job succeeded"
			sleep 2s
        	break 1
    	fi
	done
}