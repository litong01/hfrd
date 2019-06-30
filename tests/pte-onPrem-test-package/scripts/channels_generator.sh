#!/bin/bash
#
#   Channel Generator
#   Description: Used to generate channels.
#   Dependencies: network.json, channel.json
#   Note: network.json contains each org's service credentials in blockchain network
#   Two parameters :
#		1)path of network.json
#		2)path of channel.json
source hfrd_test.cfg

. scripts/utils.sh

MAX_RETRY=5
PROG="[hfrd-channels-generator]"
export MAXMESSAGECOUNTPATH=".config.channel_group.groups.Orderer.values.BatchSize.value.max_message_count"
export ABSOLUTEMAXBYTESPATH=".config.channel_group.groups.Orderer.values.BatchSize.value.absolute_max_bytes"
export PREFERREDMAXBYTESPATH=".config.channel_group.groups.Orderer.values.BatchSize.value.preferred_max_bytes"
export CHANNELRESTRICTIONSMAXCOUNTPATH=".config.channel_group.groups.Orderer.values.ChannelRestrictions.value.max_count"
export MAXBATCHTIMEOUTPATH=".config.channel_group.groups.Orderer.values.BatchTimeout.value.timeout"



log() {
	printf "${PROG}  ${1}\n" | tee -a $HOME/results/logs/channels_generator.log
}

# Help function
Print_Help() {
	log ""
	log "Usage:"
	log "./channels.sh [OPTIONS]"
	log ""
	log "Options:"
	log "-n | --networkPath           :   The path of network.json which contains all of the service credentials in blockchain network,including msp_id,networkId,API key/secret,"
	log "-c | --channelPath           :   The path of channel configuration file which is used to create/join orgs into different channels.This will provide greate flexibility"
}

# Parse the input arguments
Parse_Arguments() {
	while [ $# -gt 0 ]; do
		case $1 in
			--network | -n)
				shift
				networkPath=$1
				;;
			--channels | -c)
				shift
				channelPath=$1
				;;
			--help | -h)
				Print_Help
				;;
		esac
		shift
	done

	# If no components are passed, print error
	if [ -z "$networkPath" ] || [ -z "$channelPath" ]; then
		log "ERROR: No enough parameters supplied."
		Print_Help
		exit 1
	fi

}

buildChannelObject(){
	channelObject="{\"channel_req\":{\"updatePolicy\":{\"n_out_of\":0,\"implicit_policy\":\"ANY\"},\"members\":{"
	member="\"{msp_id}\": {\"roles\":[\"admin\",\"editor\",\"viewer\"]}"
	for msp_id in ${channelMembers[@]}
	do
		channelMember=$(echo $member | sed -e 's@{msp_id}@'"$msp_id"'@g')','
		channelObject=${channelObject}${channelMember}
	done
	channelObject=${channelObject%,*}'}}}'
}

createChannel(){
	log "Create channel: $channelName "
	buildChannelObject
    statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
                    -H "Accept:application/json" \
                    -H "Content-type:application/json" \
					-u $ORG_API_KEY:$ORG_API_SECRET \
					--data "$channelObject" \
                    -X POST \
                    $NET_CREATE_CHANNEL
    )
    if [ "$statuscode" -eq 200 ]; then
        return 0
    else
		return 1
    fi
}

getChannelConfig(){
    log "Channel configuration: $channelName "
    getUrl=${Host}/api/v1/networks/${networkId}/channels/${channelName}/config
    command="$ORG_API_KEY:$ORG_API_SECRET $getUrl"
    curl -X POST -u $command > ./channel_update/${channelName}.json

    log ""
    log "---------------------------------------------"
    log "Original ${channelName} configuration"
    log "---------------------------------------------"
    max_message_count=$(jq -r "$MAXMESSAGECOUNTPATH" ./channel_update/${channelName}.json)
    log "Max message count: "$max_message_count
    jq "$MAXMESSAGECOUNTPATH = $messageCount" ./channel_update/${channelName}.json > ./channel_update/temp_${channelName}.json

    absolute_max_bytes=$(jq -r "$ABSOLUTEMAXBYTESPATH" ./channel_update/${channelName}.json)
    log "Absolute max bytes: "$absolute_max_bytes
    jq "$ABSOLUTEMAXBYTESPATH = $absoluteMaxBytes" ./channel_update/temp_${channelName}.json > ./channel_update/temp_${channelName}_1.json

    preferred_max_bytes=$(jq -r "$PREFERREDMAXBYTESPATH" ./channel_update/${channelName}.json)
    log "Preferred max byte: "$preferred_max_bytes
    jq "$PREFERREDMAXBYTESPATH = $preferredMaxBytes" ./channel_update/temp_${channelName}_1.json > ./channel_update/temp_${channelName}.json

    channel_restrictions_max_count=$(jq -r "$CHANNELRESTRICTIONSMAXCOUNTPATH" ./channel_update/${channelName}.json)
    log "Channel restriction max count: "$channel_restrictions_max_count
    jq "$CHANNELRESTRICTIONSMAXCOUNTPATH = \"$channelRestrictionMaxCount\" " ./channel_update/temp_${channelName}.json > ./channel_update/temp_${channelName}_1.json

    batch_timeout=$(jq -r "$MAXBATCHTIMEOUTPATH" ./channel_update/${channelName}.json)
    log "Batch Timeout: "$batch_timeout
    jq "$MAXBATCHTIMEOUTPATH = \"$batchTimeout\" " ./channel_update/temp_${channelName}_1.json > ./channel_update/modified_${channelName}.json

    # cleaning up
    rm -f ./channel_update/temp_${channelName}*.json

    # Checking if update to channel configuration is needed or not
    if [ "$max_message_count" == "$messageCount" ] && [ "$absolute_max_bytes" == "$absoluteMaxBytes" ] && \
       [ "$preferred_max_bytes" == "$preferredMaxBytes" ] && [ "$channel_restrictions_max_count" == "$channelRestrictionMaxCount" ] && \
       [ "$batch_timeout" == "$batchTimeout" ]; then
        log "Update to ${channelName} config not necessary. Your specification matches current ${channelName} configuration"
        return 1
    else
        log "Updating ${channelName} configuration"
        return 0
    fi
}

updateChannelConfig(){
    # Updating channel configuration with updated json
    original_config_json=$(cat ./channel_update/${channelName}.json)
    updated_config_json=$(cat ./channel_update/modified_${channelName}.json)
    #msp_id=PeerOrg1
    msp_id=${channelMembers[0]}  #getting msp_id from the channel members
    #echo "MSPID: ${msp_id}"

    curl -H "Content-Type:application/json" -X POST \
    -u admin:pass4chain \
    --data "{\"original_json\":$original_config_json,\"updated_json\":$updated_config_json,\"msp_id\":\"$msp_id\"}" \
    ${Host}/api/network/${networkId}/channel/${channelName}/update

    log ""
    log "Download the latest config file for ${channelName}"

    # verifying if changes worked
    getUrl=${Host}/api/v1/networks/${networkId}/channels/${channelName}/config
    command="$ORG_API_KEY:$ORG_API_SECRET $getUrl"
    curl -X POST -u $command > ./channel_update/updated_${channelName}.json

    max_message_count=$(jq -r $MAXMESSAGECOUNTPATH ./channel_update/updated_${channelName}.json)
    absolute_max_bytes=$(jq -r $ABSOLUTEMAXBYTESPATH ./channel_update/updated_${channelName}.json)
    preferred_max_bytes=$(jq -r $PREFERREDMAXBYTESPATH ./channel_update/updated_${channelName}.json)
    channel_restrictions_max_count=$(jq -r $CHANNELRESTRICTIONSMAXCOUNTPATH ./channel_update/updated_${channelName}.json)
    batch_timeout=$(jq -r $MAXBATCHTIMEOUTPATH ./channel_update/updated_${channelName}.json)

    if [ "$max_message_count" != "$messageCount" ];then
        echo "Update batchsize message count failed in ${channelName}: Expected :$messageCount Actually:$max_message_count" >> ./channel_update/failed_list.out
    fi
    if [ "$absolute_max_bytes" != "$absoluteMaxBytes" ];then
        echo "Update absolute max byte count failed in ${channelName}: Expected :$absoluteMaxBytes Actually:$absolute_max_bytes" >> ./channel_update/failed_list.out
    fi
    if [ "$preferred_max_bytes" != "$preferredMaxBytes" ];then
        echo "Update preferred max bytes failed in ${channelName}: Expected :$preferredMaxBytes Actually:$preferred_max_bytes" >> ./channel_update/failed_list.out
    fi
    if [ "$channel_restrictions_max_count" != "$channelRestrictionMaxCount" ];then
        echo "Update channel restrictions max count failed in ${channelName}: Expected :$channelRestrictionMaxCount Actually:$channel_restrictions_max_count" >> ./channel_update/failed_list.out
    fi
    if [ "$batch_timeout" != "$batchTimeout" ];then
        echo "Update batchtimeout count failed in ${channelName}: Expected :$batchTimeout Actually:$batch_timeout" >> ./channel_update/failed_list.out
    fi
}

verifyUpdateSuccess(){
    # Verifies successful update of channel configuration
    failed_file=./channel_update/failed_list.out
    if [[ -f $failed_file ]]; then
        log "Failed to update channel configuration"
        log "Please take a look at ${failed_file}"
        return 1
    else
        log "Successfully updated ${channelName} configuration"
        return 0
    fi
}


joinChannel(){
    log "Join $msp_id into channel: $channelName "
    statuscode=$(curl -k -o /dev/null -s -w %{http_code} \
                -H "Accept:application/json" \
                -H "Content-type:application/json" \
                -u $ORG_API_KEY:$ORG_API_SECRET \
                -X POST \
                $NET_JOIN_CHANNEL
        )
    if [ "$statuscode" -eq 200 ]; then
        return 0
    else
        return 1
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

# Setting up channel_update directory
rm -rf ./channel_update
mkdir ./channel_update


#	Step 1: send request 'create channel' to helios
#	Step 2: send request 'join channel for each channel member' to helios
jq -c '.channels[]' $channelPath | while read channel; do
    channelName=$(echo $channel | jq -r '.name')
    declare -a channelMembers="($(echo $channel | jq -r '.members[]'))"

    API_ENDPOINT=$(jq -r .${channelMembers[0]}.url $networkPath)
    networkId=$(jq -r .${channelMembers[0]}.network_id $networkPath)
    Host=$(jq -r .${channelMembers[0]}.url $networkPath)

    # create channel
    ORG_API_KEY=$(jq -r .${channelMembers[0]}.key $networkPath)
    ORG_API_SECRET=$(jq -r .${channelMembers[0]}.secret $networkPath)
    NET_CREATE_CHANNEL=$API_ENDPOINT"/api/v1/networks/"$networkId"/channels/$channelName/create"
    runWithRetry 'createChannel'

    # below feature only supported in cm environment for now
    if [[ $env == 'cm' ]];then
        messageCount="$(echo $channel | jq -r '.batchSize.messageCount')"
        absoluteMaxBytes="$(echo $channel | jq -r '.batchSize.absoluteMaxBytes')"
        preferredMaxBytes="$(echo $channel | jq -r '.batchSize.preferredMaxBytes')"
        batchTimeout="$(echo $channel | jq -r '.batchTimeout')"
        channelRestrictionMaxCount="$(echo $channel | jq -r '.channelRestrictionMaxCount')"


        # getting channel configuration for each channel
        getChannelConfig

        update_needed=$? # stores the return code of verify_success function
        if [[ $update_needed -eq 0 ]] # if update is needed, call update channel config function else jump to join channel
        then
            # Updating channel configuration
            runWithRetry 'updateChannelConfig'

            # Verify no update failure success
            verifyUpdateSuccess
            success_status=$? # stores the return code of verify_success function
            if [[ $success_status -eq 1 ]]
            then
                break 1
            fi
        else
            log "Attempting to join channel"
        fi
    fi
    # join channel
    NET_JOIN_CHANNEL=$API_ENDPOINT"/api/v1/networks/"$networkId"/channels/$channelName/join"
    for msp_id in ${channelMembers[@]}
    do
        ORG_API_KEY=$(jq -r .${msp_id}.key $networkPath)
        ORG_API_SECRET=$(jq -r .${msp_id}.secret $networkPath)
        runWithRetry 'joinChannel'
    done
done