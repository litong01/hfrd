#!/bin/bash

set -e

FABRIC_VERSION="$1"
DEFAULT_FABRIC_VERSION=1.2
FABRIC_ENV_FILE=./fixtures/.env
# If you are using sample network to drive test modules,you must set TEST_TYPE to local
export TEST_TYPE=local
export GOPATH=/opt/gopath
export WORKDIR=${GOPATH}/src/hfrd/modules/gosdk
export GOSDK_IMAGE=hfrd/gosdk:amd64-latest

# Required input parameter: fabricVersion string
function setFabricVersion() {
    fabricVersion=$1
    if [[ ${fabricVersion} == "" ]]; then
        fabricVersion=${DEFAULT_FABRIC_VERSION}
    fi
    # Supported fabric versions: ["1.2", "1.3", "1.4"]
    case ${fabricVersion} in
    "1.1")
        echo "Set fabric version to 1.1"
        echo "FABRIC_CA_FIXTURE_TAG=1.1.0
FABRIC_COUCHDB_FIXTURE_TAG=0.4.10
FABRIC_PEER_FIXTURE_TAG=1.1.0
FABRIC_ORDERER_FIXTURE_TAG=1.1.0" >${FABRIC_ENV_FILE}
        ;;
    "1.2")
        echo "Set fabric version to 1.2"
        echo "FABRIC_CA_FIXTURE_TAG=1.2.0
FABRIC_COUCHDB_FIXTURE_TAG=0.4.10
FABRIC_PEER_FIXTURE_TAG=1.2.0
FABRIC_ORDERER_FIXTURE_TAG=1.2.0" >${FABRIC_ENV_FILE}
        ;;
    "1.3")
        echo "Set fabric version to 1.3"
        echo "FABRIC_CA_FIXTURE_TAG=1.3.0
FABRIC_COUCHDB_FIXTURE_TAG=0.4.10
FABRIC_PEER_FIXTURE_TAG=1.3.0
FABRIC_ORDERER_FIXTURE_TAG=1.3.0">${FABRIC_ENV_FILE}
        ;;
    "1.4")
        echo "Set fabric version to 1.4"
        echo "FABRIC_CA_FIXTURE_TAG=1.4.0
FABRIC_COUCHDB_FIXTURE_TAG=0.4.14
FABRIC_PEER_FIXTURE_TAG=1.4.0
FABRIC_ORDERER_FIXTURE_TAG=1.4.0">${FABRIC_ENV_FILE}
        ;;
    *)
        echo "Unsupported fabric version: ${fabricVersion}"
        exit 1
    esac
}

setFabricVersion ${FABRIC_VERSION}

if [[ ${DOCKER_RUN} != false ]]
then
    echo "Run test in docker mode"
    export SAMPLE_CC_PATH=chaincode/samplecc
    export MARBLES_CC_PATH=chaincode/marbles
    export MARBLES_PVT_CC_PATH=chaincode/marbles_private
    export COLLECTIONS_CONFIG_PATH=${GOPATH}/src/chaincode/marbles_private/collections_config.json
    cd ../../
    make gosdk-docker
    cd modules/gosdk
    export DOCKER_BASE_CMD="docker run --rm --network host -v $(eval pwd):${WORKDIR} -v $(eval pwd)/../../chaincode:${GOPATH}/src/chaincode -e \"TEST_TYPE=${TEST_TYPE}\" -e METRICS_TARGET_URL=${METRICS_TARGET_URL} ${GOSDK_IMAGE}"
else
    echo "Run test in binary mode"
    export SAMPLE_CC_PATH=hfrd/chaincode/samplecc
    export MARBLES_CC_PATH=hfrd/chaincode/marbles
    export MARBLES_PVT_CC_PATH=hfrd/chaincode/marbles_private
    export COLLECTIONS_CONFIG_PATH=${GOPATH}/src/hfrd/chaincode/marbles_private/collections_config.json
    export PATH=$PATH:`pwd`
    dep ensure --vendor-only -v
    go build
    export DOCKER_BASE_CMD=""
fi

cd fixtures
./network_setup.sh restart
cd ..

applicationCapability=V1_1
echo "#############################Create channels with channelNamePrefix########################"
if [[ $fabricVersion != "1.1" ]]
then
    applicationCapability=V1_2
fi
${DOCKER_BASE_CMD} \
gosdk channel create -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --applicationCapability ${applicationCapability} --channelNamePrefix mychannel --channelConsortium SampleConsortium \
--channelOrgs org1,org2 --ordererName orderer.example.com --iterationCount 5 --iterationInterval 0.1s --retryCount 5 --logLevel DEBUG 

echo "#############################Create channels with channelNameList########################"
${DOCKER_BASE_CMD} \
gosdk channel create -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --applicationCapability ${applicationCapability} --channelNameList mychannel5,mychannel6,mychannel7,mychannel8,mychannel9 --channelConsortium SampleConsortium \
--channelOrgs org1,org2 --ordererName orderer.example.com --iterationCount 5 --iterationInterval 0.1s --retryCount 5 --logLevel DEBUG 

echo "#############################Join org1 into channels########################"
${DOCKER_BASE_CMD} \
gosdk channel join -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --channelNamePrefix mychannel \
--peers peer0.org1.example.com,peer1.org1.example.com --ordererName orderer.example.com --iterationCount 10 --iterationInterval 0.1s --retryCount 5 --logLevel DEBUG 

echo "#############################Join org2 into channels with channelNameList########################"
${DOCKER_BASE_CMD} \
gosdk channel join -c ${WORKDIR}/fixtures/ConnectionProfile_org2.yaml --channelNameList \
mychannel0,mychannel1,mychannel2,mychannel3,mychannel4,mychannel5,mychannel6,mychannel7,mychannel8,mychannel9 \
--peers peer0.org2.example.com,peer1.org2.example.com --ordererName orderer.example.com \
--iterationCount 10 --iterationInterval 0.1s --retryCount 5 --logLevel DEBUG 

echo "############################# Replace orderer addresses ########################"
${DOCKER_BASE_CMD} \
gosdk channel update -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --channelNamePrefix mychannel --prefixOffset 0 \
        --ordererOrgName ordererorg --ordererName orderer.example.com --peers peer0.org1.example.com \
        --ordererAddresses orderer.example.com:7050 \
        --batchTimeout 1s --maxMessageCount 200 --preferredMaxBytes 103802353 --anchorPeers peer0.org1.example.com:7051 \
        --iterationCount 1 --iterationInterval 2s --retryCount 5 --logLevel DEBUG 

echo "#############################Query org1 channels########################"
${DOCKER_BASE_CMD} \
gosdk channel query -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --channelName mychannel0 --logLevel INFO \
        --peers peer0.org1.example.com --iterationCount 2 --iterationInterval 0.1s --retryCount 5

echo "#############################Install chaincode samplecc on org1 peers########################"
${DOCKER_BASE_CMD} \
gosdk chaincode install -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeNamePrefix samplecc- --chaincodeVersion v1 \
--path ${SAMPLE_CC_PATH} --peers peer0.org1.example.com,peer1.org1.example.com \
--iterationCount 2 --iterationInterval 0.1s --retryCount 5

echo "#############################Install chaincode samplecc on org2 peers########################"
${DOCKER_BASE_CMD} \
gosdk chaincode install -c ${WORKDIR}/fixtures/ConnectionProfile_org2.yaml --chaincodeNamePrefix samplecc- --chaincodeVersion v1 \
--path ${SAMPLE_CC_PATH} --peers peer0.org2.example.com,peer1.org2.example.com \
--iterationCount 2 --iterationInterval 0.1s --retryCount 5

# Loop cc instantiate to catch random errors
for i in {0..9}
do
    echo "#############################Instantiate chaincode samplecc-0 with policy on mychannel${i}########################"
    # Instantiate cc on all peers
    ${DOCKER_BASE_CMD} \
    gosdk chaincode instantiate -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName samplecc-0 --chaincodeVersion v1 \
    --channelNamePrefix mychannel --prefixOffset ${i} --path ${SAMPLE_CC_PATH} --policyStr "AND ('Org1MSP.member','Org2MSP.member')" \
    --peers peer0.org1.example.com,peer1.org1.example.com,peer0.org2.example.com,peer1.org2.example.com \
    --iterationCount 1 --iterationInterval 0.2s --retryCount 5
done

sleep 5
# Loop cc invoke module to catch random errors
for i in {1..10}
do
    echo "#############################Invoke chaincode samplecc-$i########################"
    ${DOCKER_BASE_CMD} \
    sh -c "gosdk chaincode invoke  -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName samplecc-0 --channelName mychannel0 \
    --chaincodeParams literal~~~invoke#literal~~~put#stringPattern~~~account[0-9]#stringPattern~~~[0-9]{5}#sequentialString~~~*marbles \
    --peers peer0.org1.example.com,peer1.org2.example.com \
    --iterationCount 1 --retryCount 5 --concurrencyLimit 1 --logLevel ERROR --fabricVersion ${fabricVersion} --serviceDiscovery false \
    || (cat config/mychannel0.yaml && exit 1)"

    echo "#############################Invoke chaincode samplecc--Verify service discovery feature-$i########################"
    ${DOCKER_BASE_CMD} \
    sh -c "gosdk chaincode invoke  -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName samplecc-0 --channelName mychannel0 \
    --chaincodeParams literal~~~invoke#literal~~~put#stringPattern~~~account[0-9]#stringPattern~~~[0-9]{5}#sequentialString~~~*marbles \
    --peers  peer0.org1.example.com --serviceDiscovery true \
    --iterationCount 1 --retryCount 5 --concurrencyLimit 1 --logLevel ERROR --fabricVersion ${fabricVersion} \
    || (cat config/mychannel0.yaml && exit 1)"
done

 echo "#############################Install chaincode marbles on org1 peers########################"
 ${DOCKER_BASE_CMD} \
 gosdk chaincode install -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeNamePrefix marbles- --chaincodeVersion v1 \
 --path ${MARBLES_CC_PATH} --peers peer0.org1.example.com,peer1.org1.example.com \
 --iterationCount 2 --iterationInterval 0.1s --retryCount 5

 echo "#############################Install chaincode marbles on org2 peers########################"
 ${DOCKER_BASE_CMD} \
 gosdk chaincode install -c ${WORKDIR}/fixtures/ConnectionProfile_org2.yaml --chaincodeNamePrefix marbles- --chaincodeVersion v1 \
 --path ${MARBLES_CC_PATH} --peers peer0.org2.example.com,peer1.org2.example.com \
 --iterationCount 2 --iterationInterval 0.1s --retryCount 5

 echo "#############################Instantiate chaincode marbles with policy########################"
 # Instantiate cc on all peers
 ${DOCKER_BASE_CMD} \
 gosdk chaincode instantiate -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName marbles-0 --chaincodeVersion v1 \
 --channelNamePrefix mychannel --path ${MARBLES_CC_PATH} --policyStr "AND ('Org1MSP.member','Org2MSP.member')" \
 --peers peer0.org1.example.com,peer1.org1.example.com,peer0.org2.example.com,peer1.org2.example.com --iterationCount 2 --iterationInterval 0.2s --retryCount 5

 sleep 5 # wait for gossip sync service discovery
 for i in {1..10}
 do
     echo "#############################Invoke chaincode marbles-iteration ${i}########################"
     ${DOCKER_BASE_CMD} \
     sh -c "gosdk chaincode invoke  -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName marbles-0 --channelName mychannel0 \
     --chaincodeParams literal~~~initMarble#sequentialString~~~marbles-${i}-*#literal~~~blue#literal~~~100#literal~~~tom \
     --peers peer0.org1.example.com,peer1.org2.example.com \
     --iterationCount 1 --retryCount 5 --concurrencyLimit 5 --logLevel ERROR --fabricVersion ${fabricVersion} \
     || (cat config/mychannel0.yaml && exit 1)"
 done

 echo "#############################Invoke chaincode marbles with MVCC_READ_CONFLICT########################"
 sleep 1 # sleep batchTimeout to ensure no tx waiting for block cut
 ${DOCKER_BASE_CMD} \
 gosdk chaincode invoke  -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName marbles-0 --channelName mychannel0 \
 --chaincodeParams literal~~~initMarble#literal~~~marblesName#literal~~~blue#literal~~~100#literal~~~tom \
 --peers peer0.org1.example.com,peer1.org2.example.com \
 --iterationCount 3 --iterationInterval 0s --retryCount 5 --concurrencyLimit 5 --logLevel DEBUG --fabricVersion ${fabricVersion}

 if [[ ${fabricVersion} != "1.1" ]]; then
     echo "#############################Install chaincode marbles-pvt on org1 peers########################"
     ${DOCKER_BASE_CMD} \
 gosdk chaincode install -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeNamePrefix marbles-pvt- --chaincodeVersion v1 \
 --path ${MARBLES_PVT_CC_PATH} --peers peer0.org1.example.com,peer1.org1.example.com \
 --iterationCount 2 --iterationInterval 0.1s --retryCount 5

     echo "#############################Install chaincode marbles-pvt on org2 peers########################"
     ${DOCKER_BASE_CMD} \
 gosdk chaincode install -c ${WORKDIR}/fixtures/ConnectionProfile_org2.yaml --chaincodeNamePrefix marbles-pvt- --chaincodeVersion v1 \
 --path ${MARBLES_PVT_CC_PATH} --peers peer0.org2.example.com,peer1.org2.example.com \
 --iterationCount 2 --iterationInterval 0.1s --retryCount 5

     echo "#############################Instantiate chaincode marbles-pvt with policy########################"
     # Instantiate cc on all peers
     ${DOCKER_BASE_CMD} \
 gosdk chaincode instantiate -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName marbles-pvt-0 --chaincodeVersion v1 \
 --channelNamePrefix mychannel --path ${MARBLES_CC_PATH} --collectionsConfigPath  ${COLLECTIONS_CONFIG_PATH} --policyStr "OR ('Org1MSP.member','Org2MSP.member')" \
 --peers peer0.org1.example.com,peer1.org1.example.com,peer0.org2.example.com,peer1.org2.example.com --iterationCount 2 --iterationInterval 0.2s --retryCount 5

     sleep 5 # wait for gossip sync service discovery
     for i in {1..10}
     do
         echo "#############################Invoke chaincode marbles pvt : initMarble ${i}########################"
         export MARBLE=$(echo -n "{\"name\":\"marble1\",\"color\":\"blue\",\"size\":35,\"owner\":\"tom\",\"price\":99}" | base64)
         ${DOCKER_BASE_CMD} \
         sh -c "gosdk chaincode invoke  -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName marbles-pvt-0 --channelName mychannel0 \
             --peers peer0.org1.example.com \
             --chaincodeParams literal~~~initMarble \
             --dynamicTransientMapKs marble,test --dynamicTransientMapVs sequentialString~~~name~~~marble-${i}-*#literal~~~color~~~blue#literal~~~size~~~35#literal~~~owner~~~tom#literal~~~price~~~99,sequentialString~~~name~~~test#literal~~~purpose~~~pvt \
             --iterationCount 1 --iterationInterval 1s --fabricVersion ${fabricVersion} \
             || (cat config/mychannel0.yaml && exit 1)"
           #  --transientMap "{\"marble\":\"$MARBLE\"}" \
     done

     echo "#############################Invoke chaincode marbles pvt: readMarble ########################"
     export MARBLE_DELETE=$(echo -n "{\"name\":\"marble-1-0\"}" | base64)
     ${DOCKER_BASE_CMD} \
     gosdk chaincode invoke  -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --chaincodeName marbles-pvt-0 --channelName mychannel0 \
         --peers peer0.org1.example.com \
         --chaincodeParams literal~~~readMarble#sequentialString~~~marble-1-* \
         --logLevel DEBUG --iterationCount 1 --iterationInterval 0.1s --queryOnly true --fabricVersion ${fabricVersion}
 fi

echo "#############################Add new org to existing channel########################"
${DOCKER_BASE_CMD} \
gosdk channel addorg -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --channelNamePrefix mychannel \
--ordererOrgName ordererorg --ordererName orderer.example.com --peers peer0.org1.example.com --orgConfigPath ${WORKDIR}/org3.json \
--iterationCount 1 --retryCount 5

echo "#############################Verify add new org with query channel config again########################"
${DOCKER_BASE_CMD} \
gosdk channel query -c ${WORKDIR}/fixtures/ConnectionProfile_org1.yaml --channelName mychannel0 \
--peers peer0.org1.example.com --iterationCount 2 --iterationInterval 2s --retryCount 5 --logLevel INFO

cd fixtures
./network_setup.sh down
