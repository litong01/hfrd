#!/bin/bash
set -o pipefail
mkdir -p $WORKDIR/results

heliosUser='broker'
GET_CREDS_API=$heliosUrl'/api/sid/'

cd /opt/src/scripts/onPrem
. utils.sh

numOfOrgs=${serviceid#*-}
serviceid=${serviceid%-*}
# service instance ids used to create multi organizations
for((i = 0; i<$numOfOrgs;i++))
do
    serviceids[$i]=$serviceid'-'$i
done

# "Get service creadentials from Helios Server"
getServiceCredentials

rm -f temp.json
mv network.json $WORKDIR/results/network.json
serviceid=$serviceid'-'$numOfOrgs
echo '{' > $WORKDIR/results/service.json
echo '  "serviceid": "'$serviceid'"' >> $WORKDIR/results/service.json
echo '}' >> $WORKDIR/results/service.json

