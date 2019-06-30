#!/usr/bin/env bash

printf "

███████████  ████████████      ████████       ████████
███████████  ███████████████   █████████     █████████
   █████        ████   █████     ████████   ████████
   █████        ███████████      ████ ████ ████ ████
   █████        ███████████      ████  ███████  ████
   █████        ████   █████     ████   █████   ████
███████████  ███████████████   ██████    ███    ██████
███████████  ████████████      ██████     █     ██████

  _____  ___    ___   __ _             _            
  \_   \/ __\  / _ \ / _\ |_ __ _ _ __| |_ ___ _ __ 
   / /\/__\// / /_)/ \ \| __/ _' | '__| __/ _ \ '__|
/\/ /_/ \/  \/ ___/  _\ \ || (_| | |  | ||  __/ |   
\____/\_____/\/      \__/\__\__,_|_|   \__\___|_|   
                                                    
                 _     ___  _____  __               
  __ _ _ __   __| |   / _ \/__   \/__\              
 / _' | '_ \ / _' |  / /_)/  / /\/_\                
| (_| | | | | (_| | / ___/  / / //__                
 \__,_|_| |_|\__,_| \/      \/  \__/  vem              
                                          
"

PROG="[PTE-Test-2o2pp1c]"

####################
# Helper Functions #
####################
get_pem() {
	awk '{printf "%s\\n", $0}' creds/org"$1"admin/msp/signcerts/cert.pem
}
log() {
	printf "${PROG}  ${1}\n" | tee -a run.log
}


#################
# Sanity Checks #
#################
if [[ ! -f ${HOME}/creds/apikeys.json ]]; then
	echo "Missing helios api keys, cannot continue."
	exit 1
elif [[ ! -f ${HOME}/creds/org1ConnectionProfile.json ]]; then 
	echo "Missing org1 connection profile, cannot continue."
	exit1
elif [[ ! -f ${HOME}/creds/org2ConnectionProfile.json ]]; then 
	echo "Missing org2 connection profile, cannot continue."
	exit1
fi


#################
# Environmental #
#################
API_ENDPOINT=$(jq -r .org1.url creds/apikeys.json)
NETWORK_ID=$(jq -r .org1.network_id creds/apikeys.json)
ORG1_API_KEY=$(jq -r .org1.key creds/apikeys.json)
ORG2_API_KEY=$(jq -r .org2.key creds/apikeys.json)
ORG1_API_SECRET=$(jq -r .org1.secret creds/apikeys.json)
ORG2_API_SECRET=$(jq -r .org2.secret creds/apikeys.json)
ORG1_ENROLL_SECRET=$(jq -r '.certificateAuthorities["org1-ca"].registrar[0].enrollSecret' creds/org1ConnectionProfile.json)
ORG2_ENROLL_SECRET=$(jq -r '.certificateAuthorities["org2-ca"].registrar[0].enrollSecret' creds/org2ConnectionProfile.json)
ORG1_CA_URL=$(jq -r '.certificateAuthorities["org1-ca"].url' creds/org1ConnectionProfile.json | cut -d '/' -f 3)
ORG2_CA_URL=$(jq -r '.certificateAuthorities["org2-ca"].url' creds/org2ConnectionProfile.json | cut -d '/' -f 3)


############################################################
# STEP 1 - generate user certs and upload to remote fabric #
############################################################
# save the cert
jq -r '.certificateAuthorities["org1-ca"].tlsCACerts.pem' creds/org1ConnectionProfile.json > ${HOME}/cacert.pem
log "Enrolling admin user for org1."
export FABRIC_CA_CLIENT_HOME=${HOME}/creds/org1admin
fabric-ca-client enroll --tls.certfiles ${HOME}/cacert.pem -u https://admin:${ORG1_ENROLL_SECRET}@${ORG1_CA_URL}
# rename the keyfile
mv ${HOME}/creds/org1admin/msp/keystore/* ${HOME}/creds/org1admin/msp/keystore/priv.pem
# upload the cert
BODY1=$(cat <<EOF1
{
	"msp_id": "org1",
	"adminCertName": "PeerAdminCert1",
	"adminCertificate": "$(get_pem 1)",
	"peer_names": [
		"org1-peer1"
	],
	"SKIP_CACHE": true
}
EOF1
)
log "Uploading admin certificate for org 1."
curl -s -X POST \
	--header 'Content-Type: application/json' \
	--header 'Accept: application/json' \
	--basic --user ${ORG1_API_KEY}:${ORG1_API_SECRET} \
	--data "${BODY1}" \
    ${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/certificates

# STEP 1.2 - ORG2
log "Enrolling admin user for org2."
export FABRIC_CA_CLIENT_HOME=${HOME}/creds/org2admin
fabric-ca-client enroll --tls.certfiles ${HOME}/cacert.pem -u https://admin:${ORG2_ENROLL_SECRET}@${ORG2_CA_URL}
# rename the keyfile
mv ${HOME}/creds/org2admin/msp/keystore/* ${HOME}/creds/org2admin/msp/keystore/priv.pem
# upload the cert
BODY2=$(cat <<EOF2
{
 "msp_id": "org2",
 "adminCertName": "PeerAdminCert2",
 "adminCertificate": "$(get_pem 2)",
 "peer_names": [
   "org2-peer1"
 ],
 "SKIP_CACHE": true
}
EOF2
)
log "Uploading admin certificate for org 2."
curl -s -X POST \
	--header 'Content-Type: application/json' \
	--header 'Accept: application/json' \
	--basic --user ${ORG2_API_KEY}:${ORG2_API_SECRET} \
	--data "${BODY2}" \
    ${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/certificates


##########################
# STEP 2 - restart peers #
##########################
# STEP 2.1 - ORG1
PEER="org1-peer1"
log "Stoping ${PEER}"
curl -s -X POST \
	--header 'Content-Type: application/json' \
	--header 'Accept: application/json' \
	--basic --user ${ORG1_API_KEY}:${ORG1_API_SECRET} \
	--data-binary '{}' \
	${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/${PEER}/stop

log "Waiting for ${PEER} to stop..."
RESULT=""
while [[ ${RESULT} != "exited" ]]; do
	RESULT=$(curl -s -X GET \
		--header 'Content-Type: application/json' \
		--header 'Accept: application/json' \
		--basic --user ${ORG1_API_KEY}:${ORG1_API_SECRET} \
		${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/status | jq -r '.["'${PEER}'"].status')
done
log "${RESULT}"

log "Starting ${PEER}"
curl -s -X POST \
	--header 'Content-Type: application/json' \
	--header 'Accept: application/json' \
	--basic --user ${ORG1_API_KEY}:${ORG1_API_SECRET} \
	--data-binary '{}' \
	${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/${PEER}/start

log "Waiting for ${PEER} to start..."
RESULT=""
while [[ ${RESULT} != "running" ]]; do
	RESULT=$(curl -s -X GET \
		--header 'Content-Type: application/json' \
		--header 'Accept: application/json' \
		--basic --user ${ORG1_API_KEY}:${ORG1_API_SECRET} \
		${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/status | jq -r '.["'${PEER}'"].status')
done
log "${RESULT}"

# STEP 2.2 - ORG2
PEER="org2-peer1"
log "Stoping ${PEER}"
curl -s -X POST \
	--header 'Content-Type: application/json' \
	--header 'Accept: application/json' \
	--basic --user ${ORG2_API_KEY}:${ORG2_API_SECRET} \
	--data-binary '{}' \
	${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/${PEER}/stop

log "Waiting for ${PEER} to stop..."
RESULT=""
while [[ $RESULT != "exited" ]]; do
	RESULT=$(curl -s -X GET \
		--header 'Content-Type: application/json' \
		--header 'Accept: application/json' \
		--basic --user ${ORG2_API_KEY}:${ORG2_API_SECRET} \
		${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/status | jq -r '.["'${PEER}'"].status')
done
log "${RESULT}"

log "Starting ${PEER}"
curl -s -X POST \
	--header 'Content-Type: application/json' \
	--header 'Accept: application/json' \
	--basic --user ${ORG2_API_KEY}:${ORG2_API_SECRET} \
	--data-binary '{}' \
	${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/${PEER}/start

log "Waiting for ${PEER} to start..."
RESULT=""
while [[ $RESULT != "running" ]]; do
	RESULT=$(curl -s -X GET \
		--header 'Content-Type: application/json' \
		--header 'Accept: application/json' \
		--basic --user ${ORG2_API_KEY}:${ORG2_API_SECRET} \
		${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/nodes/status | jq -r '.["'${PEER}'"].status')
done
log "${RESULT}"


#########################
# STEP 3 - SYNC CHANNEL #
#########################
log "Syncing the channel."
curl -s -X POST \
	--header 'Content-Type: application/json' \
  	--header 'Accept: application/json' \
  	--basic --user ${ORG1_API_KEY}:${ORG1_API_SECRET} \
  	--data-binary '{}' \
  	${API_ENDPOINT}/api/v1/networks/${NETWORK_ID}/channels/defaultchannel/sync


#####################################################
# STEP 4 - generate the SCFile and place userInputs #
#####################################################
log "Generating PTE SCFile and userInputs."
python ${HOME}/scripts/2o2pp-1ch-create_SCFile.py
cp -r ${HOME}/userInputs ${HOME}/fabric-sdk-node/test/PTE/userInputs


##################################
# STEP 5 - PTE install chaincode #
##################################
log "---------------------------------"
log "---------------------------------"
log "Installing PTE Chaincode."
log "---------------------------------"
log "---------------------------------"
source ${HOME}/.nvm/nvm.sh
cd ${HOME}/fabric-sdk-node/test/PTE
./pte_driver.sh userInputs/runCases.install.marbles.txt | tee -a run.log
sleep 5


######################################
# STEP 6 - PTE instantiate chaincode #
######################################
log "---------------------------------"
log "---------------------------------"
log "Instantiating PTE Chaincode."
log "---------------------------------"
log "---------------------------------"
./pte_driver.sh userInputs/runCases.instantiate.marbles.txt | tee -a run.log
sleep 30


####################
# STEP 7 - PTE run #
####################
log "---------------------------------"
log "---------------------------------"
log "Running PTE Now!"
log "---------------------------------"
log "---------------------------------"
./pte_driver.sh userInputs/runCases.run.marbles.txt | tee -a run.log

cp run.log ${HOME}/results
mkdir -p ${HOME}/results/creds
cp -r ${HOME}/creds/* ${HOME}/results/creds