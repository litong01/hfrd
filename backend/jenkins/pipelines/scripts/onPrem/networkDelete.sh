#!/bin/bash
set -o pipefail
mkdir -p $WORKDIR/results

heliosUser='broker'
GET_CREDS_API=$heliosUrl'/api/sid/'
AUTH_URL='https://iam.stage1.ng.bluemix.net/oidc/token'
MAX_RETRY=2

cd /opt/src/scripts/onPrem
. utils.sh


numOfOrgs=${serviceid#*-}
serviceid=${serviceid%-*}
# service instance ids used to create multi organizations
for((i = 0; i<$numOfOrgs;i++))
do
    serviceids[$i]=$serviceid'-'$i
done

echo "Get service credentials from Helios"
RESPONSE=$(curl -k -s -w "HTTPSTATUS:%{http_code}" -H "Content-Type:application/json" \
                            -u $heliosUser:$heliosSecret \
                            -X GET \
                            ${GET_CREDS_API}${serviceids[0]}
 )

res_body=$(echo "$RESPONSE" | sed -e 's/HTTPSTATUS\:.*//g')
statuscode=$(echo "$RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
if [ $statuscode -ne 200 ]; then
    if [[ $res_body =~ 'not find network from siid' ]]; then
        echo "Network has been deleted mannually"
        exit 0
    fi
    echo "Error found when delete network :$res_body"
    exit 1
fi
networkID=$(echo $res_body | jq -r .PeerOrg1.network_id)

echo "Getting access token for cluster manager"
get_access_token
NET_DELETE=$cmUrl'/api/manager/network/'$networkID

retrytime=0
while [ $retrytime -le $MAX_RETRY ]; do
    ((retrytime++))
    deleteNetwork
    if [ $? -ne 0 ]; then
        if [ $MAX_RETRY -eq $retrytime ]; then
            echo "Already reach the maximum number of attempts.Still failed to delete network.Will exit 1 "
            exit 1
        else
            echo "Delete network failed,will retry..."
        fi
    else
        echo "Successfully delete network $networkID for service:${serviceid}-${numOfOrgs}"
        break
    fi
done