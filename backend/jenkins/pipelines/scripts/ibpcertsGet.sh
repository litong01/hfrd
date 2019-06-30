#!/bin/bash
# The script to create fabric network
# This file will require the following parameters
# endpoint, apikey, org, space, service, serviceplan and serviceid
# This script produces a network.json file being placed in $WORKDIR/results
# directory
set -o pipefail
MAX_RETRY=5
COUNTER=0
echo $WORKDIR
echo "args list:$@ , the 1arg :$1" 
mkdir -p $WORKDIR/results

pwd
python /opt/src/scripts/zipgen.py -n $1  -o $WORKDIR/results
python /opt/src/scripts/uploadcert.py -d $WORKDIR/results/keyfiles 
rm -rf $WORKDIR/results/keyfiles
