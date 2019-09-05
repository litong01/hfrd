#!/bin/bash

work_dir=$PWD
ORDERER_ORG_NAME=$1
ORG_NAME_STRING=$2

IFS='.' read -ra ORDERER_OBJECT <<< "$ORDERER_ORG_NAME"
NUMBER_OF_ORDERER_NODES=${ORDERER_OBJECT[0]}
ORDERER_ORG_NAME=${ORDERER_OBJECT[1]}
IFS=',' read -ra ORG_NAMES <<< "$ORG_NAME_STRING"

source $work_dir'/apis.ini' || true
ARTIFACTS=$work_dir/config_artifacts
#Orderer Configuration Parameters
MAX_MESSAGE_COUNT=${orderer_max_message_count:-"100"}
PREFERRED_MAX_BYTES=${orderer_preferred_max_bytes:-"33554432"}
ABSOLUTE_MAX_BYTES=${orderer_absolute_max_bytes:-"103809024"}
MAX_BATCH_TIMEOUT=${orderer_batch_timeout:-"2s"}

rm -rf ${ARTIFACTS} config_update.json *.block

## restructure msps
function restructure_msps(){
    set -x
    local BASE_DIR=$PWD/crypto-config/${ORDERER_ORG_NAME}
    cp ${BASE_DIR}/ca/tls-ca-cert.pem ${BASE_DIR}/
    mv ${BASE_DIR}/admin/msp/cacerts ${BASE_DIR}/admin/
    mv ${BASE_DIR}/admin/msp/keystore ${BASE_DIR}/admin/
    mv ${BASE_DIR}/admin/msp/signcerts ${BASE_DIR}/admin/
    mkdir -p ${BASE_DIR}/admin/admincerts
    mkdir -p ${BASE_DIR}/admin/tlscacerts
    cp ${BASE_DIR}/admin/signcerts/* ${BASE_DIR}/admin/admincerts/
    mv ${BASE_DIR}/ca/tls/msp/cacerts/* ${BASE_DIR}/admin/tlscacerts/
    rm -rf ${BASE_DIR}/admin/fabric-ca-client-config.yaml
    rm -rf ${BASE_DIR}/admin/msp
    cp ${BASE_DIR}/ca/secret.json ${BASE_DIR}/
    rm -rf ${BASE_DIR}/ca
    mkdir -p ${BASE_DIR}/msp
    cp -r ${BASE_DIR}/admin/admincerts ${BASE_DIR}/msp/
    cp -r ${BASE_DIR}/admin/cacerts ${BASE_DIR}/msp/
    cp -r ${BASE_DIR}/admin/tlscacerts ${BASE_DIR}/msp/

    for PEER_ORG_NAME in ${ORG_NAMES[*]}
    do
        set -x
        local BASE_DIR=$PWD/crypto-config/${PEER_ORG_NAME}
        cp ${BASE_DIR}/ca/tls-ca-cert.pem ${BASE_DIR}/
        mv ${BASE_DIR}/admin/msp/cacerts ${BASE_DIR}/admin/
        mv ${BASE_DIR}/admin/msp/keystore ${BASE_DIR}/admin/
        mv ${BASE_DIR}/admin/msp/signcerts ${BASE_DIR}/admin/
        mkdir -p ${BASE_DIR}/admin/admincerts
        mkdir -p ${BASE_DIR}/admin/tlscacerts
        cp ${BASE_DIR}/admin/signcerts/* ${BASE_DIR}/admin/admincerts/
        mv ${BASE_DIR}/ca/tls/msp/cacerts/* ${BASE_DIR}/admin/tlscacerts/
        rm -rf ${BASE_DIR}/admin/fabric-ca-client-config.yaml
        rm -rf ${BASE_DIR}/admin/msp
        cp ${BASE_DIR}/ca/secret.json ${BASE_DIR}/
        rm -rf ${BASE_DIR}/ca
        mkdir -p ${BASE_DIR}/msp
        cp -r ${BASE_DIR}/admin/admincerts ${BASE_DIR}/msp/
        cp -r ${BASE_DIR}/admin/cacerts ${BASE_DIR}/msp/
        cp -r ${BASE_DIR}/admin/tlscacerts ${BASE_DIR}/msp/

        # User Cert Copying
        mv ${BASE_DIR}/user/msp/cacerts ${BASE_DIR}/user/
        mv ${BASE_DIR}/user/msp/keystore ${BASE_DIR}/user/
        mv ${BASE_DIR}/user/msp/signcerts ${BASE_DIR}/user/
        mkdir -p ${BASE_DIR}/user/admincerts
        mkdir -p ${BASE_DIR}/user/tlscacerts
        cp ${BASE_DIR}/admin/signcerts/* ${BASE_DIR}/user/admincerts/
        cp ${BASE_DIR}/admin/tlscacerts/* ${BASE_DIR}/user/tlscacerts/
        rm -rf ${BASE_DIR}/user/fabric-ca-client-config.yaml
        rm -rf ${BASE_DIR}/user/msp
        set +x
    done
}

restructure_msps
#downloadFabricBinaries

export PATH=$PATH:$work_dir/bin/
export ORDERER_TLS_CA=$(ls $work_dir/crypto-config/${ORDERER_ORG_NAME}/msp/tlscacerts/*)
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=${ORDERER_ORG_NAME}
export CORE_PEER_TLS_ROOTCERT_FILE=$ORDERER_TLS_CA
export CORE_PEER_MSPCONFIGPATH=$work_dir/crypto-config/${ORDERER_ORG_NAME}/admin

ORDERER_URL=$(jq -r .api_url $work_dir/crypto-config/${ORDERER_ORG_NAME}/${ORDERER_ORG_NAME}-orderer-1.json)
ORDERER_URL=${ORDERER_URL:8}
echo -e "\n Orderer URL : ${ORDERER_URL}"

echo "Waiting for Orderering service up"

while [ ${NUMBER_OF_ORDERER_NODES} -gt 0 ]
do
    NAME=${ORDERER_ORG_NAME}'orderer'${NUMBER_OF_ORDERER_NODES} ./wait_for_pod.sh
    NUMBER_OF_ORDERER_NODES=$(($NUMBER_OF_ORDERER_NODES-1))
done

sleep 30s

mkdir -p ${ARTIFACTS}

cat << EOF > ${ARTIFACTS}/configtx.yaml
---
Organizations:
EOF

for PEER_ORG_NAME in ${ORG_NAMES[@]}
do
   for ((peer_num=0;peer_num<1;peer_num++))
   do
      if [ ${peer_num} -eq 0 ]
      then
         cat << EOF >> ${ARTIFACTS}/configtx.yaml
             - &${PEER_ORG_NAME}
              Name: ${PEER_ORG_NAME}
              ID: ${PEER_ORG_NAME}
              MSPDir: $work_dir/crypto-config/${PEER_ORG_NAME}/msp
EOF
      fi
   done
done
cat << EOF >> ${ARTIFACTS}/configtx.yaml
Capabilities:
    Application: &ApplicationCapabilities
        V1_2: true
Application: &ApplicationDefaults
    Organizations:
    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: "ANY Readers"
        Writers:
            Type: ImplicitMeta
            Rule: "ANY Writers"
        Admins:
            Type: ImplicitMeta
            Rule: "MAJORITY Admins"
    Capabilities:
        <<: *ApplicationCapabilities
Profiles:
    MultiOrgsChannel:
        Consortium: SampleConsortium
        Application:
            <<: *ApplicationDefaults
            Organizations:
EOF
for PEER_ORG_NAME in ${ORG_NAMES[@]}
do
   cat << EOF >> ${ARTIFACTS}/configtx.yaml
                - *${PEER_ORG_NAME}
EOF
done

## To use the above generated configtx.yaml
export FABRIC_CFG_PATH=${ARTIFACTS}
set -x
for PEER_ORG_NAME in ${ORG_NAMES[@]}
do
   configtxgen -printOrg ${PEER_ORG_NAME} > ${ARTIFACTS}/${PEER_ORG_NAME}.json
done
set +x

## Default System channel that the orderer bootstrapped
export SYS_CHANNEL_NAME=testchainid

## To use default config configtx.yaml downloaded along with binaries
export FABRIC_CFG_PATH=$PWD/config

UPDATE_ORDERER=false
# ## Update System channel to include the org1 and org2 as members of the SampleConsortium
for PEER_ORG_NAME in ${ORG_NAMES[@]}
do
    # Fetch System channel genesis block
    set -x
    peer channel fetch config ${ARTIFACTS}/genesis.pb -o ${ORDERER_URL} -c ${SYS_CHANNEL_NAME} --cafile ${ORDERER_TLS_CA} --tls

    # Decode the genesis block and extract the config section
    configtxlator proto_decode --input ${ARTIFACTS}/genesis.pb --type common.Block --output ${ARTIFACTS}/genesis_block.json
    jq .data.data[0].payload.data.config ${ARTIFACTS}/genesis_block.json > ${ARTIFACTS}/config.json

    # Include the corresponding org msps to the sample consortium
    jq -s '.[0] * {"channel_group":{"groups":{"Consortiums":{"groups":{"SampleConsortium":{"groups": {"'${PEER_ORG_NAME}'":.[1]} }}}}}}' ${ARTIFACTS}/config.json ${ARTIFACTS}/${PEER_ORG_NAME}.json > ${ARTIFACTS}/modify_channel.json

    # Update Orderer Defaults for channels
    if [ ${UPDATE_ORDERER} = 'false' ]
    then
        export MAXBATCHSIZEPATH=".channel_group.groups.Orderer.values.BatchSize.value.max_message_count"  ABSOLUTEMAXBYTESPATH=".channel_group.groups.Orderer.values.BatchSize.value.absolute_max_bytes" PREFERREDMAXBYTESPATH=".channel_group.groups.Orderer.values.BatchSize.value.preferred_max_bytes" MAXBATCHTIMEOUT=".channel_group.groups.Orderer.values.BatchTimeout.value.timeout" ORDERERADDRESS=".channel_group.values.OrdererAddresses.value.addresses[0]"
        jq "$MAXBATCHSIZEPATH = $MAX_MESSAGE_COUNT" ${ARTIFACTS}/modify_channel.json > ${ARTIFACTS}/config1.json && jq "$PREFERREDMAXBYTESPATH = $PREFERRED_MAX_BYTES" ${ARTIFACTS}/config1.json > ${ARTIFACTS}/config2.json && jq "$ABSOLUTEMAXBYTESPATH = $ABSOLUTE_MAX_BYTES" ${ARTIFACTS}/config2.json > ${ARTIFACTS}/config1.json && jq "$ORDERERADDRESS = \"$ORDERER_URL\"" ${ARTIFACTS}/config1.json > ${ARTIFACTS}/config2.json && jq "$MAXBATCHTIMEOUT = \"$MAX_BATCH_TIMEOUT\"" ${ARTIFACTS}/config2.json > ${ARTIFACTS}/modify_channel.json
        rm -rf ${ARTIFACTS}/config1.json ${ARTIFACTS}/config2.json
    fi
    UPDATE_ORDERER=true

    # Encode the original and modified json to protobufs
    configtxlator proto_encode --input ${ARTIFACTS}/config.json --type common.Config --output ${ARTIFACTS}/config.pb
    configtxlator proto_encode --input ${ARTIFACTS}/modify_channel.json --type common.Config --output ${ARTIFACTS}/modified_config.pb

    # Compute the delta between the config changes
    configtxlator compute_update --channel_id ${SYS_CHANNEL_NAME} --original ${ARTIFACTS}/config.pb --updated ${ARTIFACTS}/modified_config.pb --output ${ARTIFACTS}/config_update.pb

    # Decode the config difference (convert to json)
    configtxlator proto_decode --input ${ARTIFACTS}/config_update.pb --type common.ConfigUpdate --output=./config_update.json

    # Prepend header etc., (enveloper) to the delta
    echo '{"payload":{"header":{"channel_header":{"channel_id":"'${SYS_CHANNEL_NAME}'", "type":2}},"data":{"config_update":'$(cat config_update.json)'}}}' | jq . > ${ARTIFACTS}/update_in_envelope.json

    # Encode the config update
    configtxlator proto_encode --input ${ARTIFACTS}/update_in_envelope.json --type common.Envelope --output ${ARTIFACTS}/update_in_envelope.pb
    echo -e "\nUpdating the system channel ${SYS_CHANNEL_NAME} to include the member org: ${PEER_ORG_NAME}\n"

    # Broadcast to the orderer to update the system channel to include the org member
    peer channel update -f ${ARTIFACTS}/update_in_envelope.pb -c ${SYS_CHANNEL_NAME} -o $ORDERER_URL --cafile $ORDERER_TLS_CA --tls
    set +x
    sleep 3
done

echo -e "\n\n S Y S T E M   C H A N N E L   U P D A T E D  T O   I N C L U D E    T H E   M E M B E R   O R G s"
