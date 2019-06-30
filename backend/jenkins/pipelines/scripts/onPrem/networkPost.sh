#!/bin/bash
# The script to create fabric network in env "OnPrem"
# This script produces a network.json file being placed in $WORKDIR/results
# directory

# TODO:Currently just support to create one network once.
# In the future, need to support create multiple networks

set -o pipefail
mkdir -p $WORKDIR/results

cd /opt/src/scripts/onPrem
. utils.sh

FABRIC_RELEASE=1.1
NET_INPUT='cmNet.json'
NET_CREATE=$cmUrl'/api/manager/networkgen/'$loc'/'$plan'/full/1/'$FABRIC_RELEASE
NET_CLAIME=$cmUrl'/api/manager/network_alpha2'
AUTH_URL='https://iam.stage1.ng.bluemix.net/oidc/token'
heliosUser='broker'
GET_CREDS_API=$heliosUrl'/api/sid/'
MAX_RETRY=5

if [ -z $serviceid ]; then
	echo "no serviceid provided..."
	exit 1
fi

# service instance ids used to create multi organizations
for((i = 0; i<$numOfOrgs;i++))
do
    serviceids[$i]=$serviceid'-'$i
done

# Get access_token for cluster manager
get_access_token
if [ $? -ne 0 ]; then
    echo "Authentication failed,exit..."
    exit 1
fi

# Firstly we can try to claime the network if there exists free networks.
# If failed we need to create one , then claim it.
claimeNetwork
if [ $? -ne 0 ]; then
    echo "Failed to claime network.Now try to create a new network"
    # Create network based on the default input and dynamic parameters
    retrytime=0
    while [ $retrytime -le $MAX_RETRY ]; do
        ((retrytime++))
        createNetwork
        if [ $? -ne 0 ]; then
            if [ $MAX_RETRY -eq $retrytime ]; then
                echo "Already reach the maximum number of attempts.Still failed to create network.Will exit 1 "
                exit 1
            else
                echo "Create network failed,will retry..."
            fi
        else
            echo "Successfully created one network"
            claimeNetwork
            if [ $? -ne 0 ]; then
                echo "Failed to claime the new network.Someone has taken it"
                exit 1
            else
                echo "Successfully claimed network $NetworkID"
                break
            fi
        fi
    done
else
    echo "Successfully claimed network $NetworkID"
fi

echo "Create Peers for $NetworkID"
NET_PEER_CREATE=$cmUrl'/api/manager/network/'$NetworkID'/peer_alpha2'

for id in ${serviceids[@]}
do
    peer_object=$(jq '.peer' $NET_INPUT | sed -e 's@{serviceid}@'"$id"'@g')
    peer_object=$(echo $peer_object | sed -e 's@{ledgertype}@'"$ledgerType"'@g')
    for peer in $(seq 1 $numOfPeers)
    do
        retrytime=0
        while [ $retrytime -le $MAX_RETRY ]; do
            ((retrytime++))
            createPeer
            if [ $? -ne 0 ]; then
                if [ $MAX_RETRY -eq $retrytime ]; then
                    echo "Already reach the maximum number of attempts.Still failed to create peer.Will exit 1 "
                    exit 1
                else
                    echo "Create peer failed,will retry..."
                fi
            else
                echo "Successfully created one peer for service: $id"
                break 1
            fi
        done
    done
done

# Get service creadentials from Helios Server
getServiceCredentials

# Update ibmid_doc to add new key/secret for the new network
API_ENDPOINT=$(jq -r '.PeerOrg1.url' network.json)
ORG_API_KEY=$(jq -r '.PeerOrg1.key' network.json)
ORG_API_SECRET=$(jq -r '.PeerOrg1.secret' network.json)
runWithRetry 'updateIBMIdDoc'

# Before join organizations into network, we must firstly run 'getServiceCredentials' to create key/secret for network in ibmids doc.
echo "Join Organizations into network $NetworkID via helios API (Update system channel to contain new orgs)"
NET_ORG_JOIN='/api/v1/networks/'$NetworkID'/joinNetwork'
for id in ${serviceids[@]}
do
    joinNetwork
    if [ $? -ne 0 ]; then
        echo "Failed to join service $id into network $NetworkID"
        exit 1
    else
        echo "Successfully join service $id into network $NetworkID"
    fi
    # sleep 5s to make sure the systemUpdate is finished
    sleep 5s
done

echo "Network:$NetworkID creation succeeded"


# move network.json to $WORKDIR/results
rm -f temp.json
mv network.json $WORKDIR/results/network.json
serviceid=$serviceid'-'$numOfOrgs
echo '{' > $WORKDIR/results/service.json
echo '  "serviceid": "'$serviceid'"' >> $WORKDIR/results/service.json
echo '}' >> $WORKDIR/results/service.json

pwd
python /opt/src/scripts/zipgen.py -n $WORKDIR/results/network.json  -o $WORKDIR/results
python /opt/src/scripts/uploadcert.py -d $WORKDIR/results/keyfiles
rm -rf $WORKDIR/results/keyfiles