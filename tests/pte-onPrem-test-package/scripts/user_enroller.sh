#!/bin/bash
#
#   User Enroller
#   Description: Used to enroll and upload user certs.
#   Dependencies: network.json,channel.json ,Connection profiles
#   Note: network.json contains each org's service credentials in blockchain network
#   Connection Profiles containes the description of one organization
#   Two parameters :
#		1)path of network.json
#		2)path of channel.json
#		3)path of Connection Profile
. utils.sh

MAX_RETRY=5

PROG="[hfrd-enrollUsers]"
DefaultCredsDir=$HOME/results/creds

log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/enroll.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./upload_certs.sh [OPTIONS]"
	log ""
	log "Options:"
	log "-n | --networkPath           :   The path of network.json which contains all of the service credentials in blockchain network,including msp_id,networkId,API key/secret,"
	log "-c | --channelPath           :   The path of channel configuration file.Here we use channel.json to know which channel should be synced"
    log "-p | --profilePath           :   The path of Connection Profile for a specific organization"
}

# Parse the input arguments
Parse_Arguments() {
	while [ $# -gt 0 ]; do
		case $1 in
			--network | -n)
				shift
				networkPath=$1
				;;
			--channel | -c)
				shift
				channelPath=$1
				;;
			--profile | -p)
				shift
				profilePath=$1
				;;
			--help | -h)
				Print_Help
				;;
		esac
		shift
	done

	# If no components are passed, print error
	if [[ -z "$networkPath" || -z "$channelPath" || -z "$profilePath" ]]; then
		log "ERROR: No enough parameters supplied."
		Print_Help
		exit 1
	fi
}

Parse_Arguments $@

# Sanity check
if [[ ! -f $networkPath ]]; then
	log "Missing hfrd configuration file, cannot continue."
	exit 1
elif [[ ! -f $channelPath ]]; then
	log "Missing channel configuration file, cannot continue."
	exit 1
fi

cleanOldCerts(){
	rm -rf $HOME/results/creds/${msp_id}*
}
enrollAdminUser(){
	fabric-ca-client enroll --tls.certfiles $HOME/results/cacert.pem -u https://admin:${ORG_ENROLL_SECRET}@${ORG_CA_URL} --mspdir $DefaultCredsDir/${msp_id}admin/msp
    mv $DefaultCredsDir/${msp_id}admin/msp/keystore/* $DefaultCredsDir/${msp_id}admin/msp/keystore/priv.pem
}

uploadUserCert(){
	log "Upload userCert for $msp_id"
	statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
					-H 'Content-Type: application/json' \
					-H 'Accept: application/json' \
					-u ${ORG_API_KEY}:${ORG_API_SECRET} \
					--data "${UPLOAD_BODY}" \
					-X POST \
    				${API_ENDPOINT}/api/v1/networks/${networkId}/certificates
			)
	if [ "$statuscode" -eq 200 ]; then
		return 0
    else
		return 1
    fi
}

startPeer(){
	statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
		-X POST \
		-H 'Content-Type: application/json' \
		-H 'Accept: application/json' \
		-u ${ORG_API_KEY}:${ORG_API_SECRET} \
		${API_ENDPOINT}/api/v1/networks/${networkId}/nodes/${PeerID}/start
	)
	if [ "$statuscode" -eq 200 ]; then
		return 0
    else
		return 1
    fi
}

stopPeer(){
	statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
		-X POST \
		-H 'Content-Type: application/json' \
		-H 'Accept: application/json' \
		-u ${ORG_API_KEY}:${ORG_API_SECRET} \
		${API_ENDPOINT}/api/v1/networks/${networkId}/nodes/${PeerID}/stop
	)
	if [ "$statuscode" -eq 200 ]; then
		return 0
    else
		return 1
    fi
}

syncChannel(){
    log "Sync channel: ${channelName} for $msp_id"
	statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
		-X POST \
		-H 'Content-Type: application/json' \
		-H 'Accept: application/json' \
  		-u ${ORG_API_KEY}:${ORG_API_SECRET} \
  		${API_ENDPOINT}/api/v1/networks/${networkId}/channels/${channelName}/sync
	)
	if [ "$statuscode" -eq 200 ]; then
		return 0
    else
		return 1
    fi
}

# Get msp_id from connection profile
msp_id=$(jq -r '.client.organization' $profilePath)
# Common variables
API_ENDPOINT=$(jq -r .$msp_id.url $networkPath)
networkId=$(jq -r .$msp_id.network_id $networkPath)
ORG_API_KEY=$(jq -r .$msp_id.key $networkPath)
ORG_API_SECRET=$(jq -r .$msp_id.secret $networkPath)
# Clean old certs before enroll and upload certs
cleanOldCerts

# Step 1 : Get cacert.pem,ORG_ENROLL_SECRET,ORG_CA_URL
ORG_CA1=$(jq -r --arg msp_id "$msp_id" '.organizations[$msp_id].certificateAuthorities[0]' $profilePath)
jq -r --arg CA "$ORG_CA1" '.certificateAuthorities[$CA].tlsCACerts.pem' $profilePath > $HOME/results/cacert.pem
ORG_ENROLL_SECRET=$(jq -r --arg CA "$ORG_CA1" '.certificateAuthorities[$CA].registrar[0].enrollSecret' $profilePath)
ORG_CA_URL=$(jq -r --arg CA "$ORG_CA1" '.certificateAuthorities[$CA].url' $profilePath | cut -d '/' -f 3)

# Step 2 : Enroll admins user
enrollAdminUser

# Step 3 : Upload cert to each peer
peers=$(jq -r --arg msp_id "$msp_id" '.organizations[$msp_id].peers' $profilePath)
UPLOAD_BODY="{\"msp_id\":\"$msp_id\",\"adminCertName\":\"PeerAdminCert1\",\"adminCertificate\":\"$(get_pem $msp_id)\",\"peer_names\":$peers,\"SKIP_CACHE\": true}"
runWithRetry 'uploadUserCert'

# Step 4 : Restart peers
jq -r --arg msp_id "$msp_id" '.organizations[$msp_id].peers[]' $profilePath | while read ORG_PEER; do
	PeerID=$ORG_PEER
	log "Stopping peer $PeerID"
	runWithRetry 'stopPeer'

	log "Waiting for ${PeerID} to stop..."
	RESULT=""
	loop_times=0
	while [[ ${RESULT} != "exited" ]]; do
		((loop_times++))
		RESULT=$(curl -s -X GET \
			-H 'Content-Type: application/json' \
			-H 'Accept: application/json' \
			-u ${ORG_API_KEY}:${ORG_API_SECRET} \
			${API_ENDPOINT}/api/v1/networks/${networkId}/nodes/status | jq -r '.["'${PeerID}'"].status')
		sleep 2s
		if [ $loop_times -eq 10 ];then
			break 1
		fi
	done
	log "${RESULT}"

	log "Starting ${PeerID}"
	runWithRetry 'startPeer'
done

#####################
# Sync the channels #
#####################
jq -c '.channels[]' $channelPath | while read channel; do
    channelName=$(echo $channel | jq -r '.name')
    declare -a channelMembers="($(echo $channel | jq -r '.members[]'))"
    for member in ${channelMembers[@]}
    do
        if [ "$member" == "$msp_id" ]; then
		    runWithRetry 'syncChannel'
        fi
    done
done