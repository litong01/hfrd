# Test Modules - gosdk(Work In Progress)

* [Quick start test modules development](#quick-start-test-modules-development)
    * [Prerequisites](#prerequisites)
    * [Quick start](#quick-start)
        * [Provisioning a fabric network for testing](#provisioning-a-fabric-network-for-testing)
        * [Start with the sample code](#start-with-the-sample-code)
    * [Build gosdk test modules](#build-gosdk-test-modules)
    * [Channel](#channel)
        * [Create](#create)
        * [Join](#join)
        * [Update](#update)
    * [Chaincode](#chaincode)
        * [Install](#install)
        * [Instantiate](#instantiate)
        * [Invoke](#invoke)

## Quick start test modules development
### Prerequisites
- docker
- docker-compose
- golang
- [dep](https://github.com/golang/dep)

### Quick start
#### Provisioning a fabric network for testing
```shell
cd fixtures
./network_setup.sh up
```
#### Start with the sample code
This part is only to demo how to use fabric-sdk-go
```shell
dep ensure -vendor-only
cd sample
go build
./sample
```

### Build gosdk test modules
From modules/gosdk directory
```shell
dep ensure -vendor-only
go build
```
This will generate an executable binary file named as `gosdk` in modules/gosdk directory

### Example
After you have build the gosdk ,you may run the gosdk_example.sh. This script will help you create/join/update/query channels and install/instantiate/invoke chaincode
Just run :
```shell
    ./gosdk_example.sh
```
Notes:
       1. If you are running hfrd test modules in sample network,you must set host env `TEST_TYPE` to  `local`.

```shell
     export TESTR_TYPE=local
```

2. You can set `CRYPTO_ROOT_PATH` to your certs root directory.Default setting is `/fabric/keyfiles`

```shell
    export CRYPTO_ROOT_PATH=/fabric/keyfiles
 ```
3. Certs  directory structure must be exactly the same as official requirements.

```shell
                $CRYPTO_ROOT_PATH/orgname/users/{username}@orgname/msp
                $CRYPTO_ROOT_PATH/orgname/tlsca/tlsca.orgname-cert.pem
```


### Channel
#### Create
With the fabric network created in [Privisioning a fabric network for testing](#provisioning-a-fabric-network-for-testing),
we can create channels with yaml configuration file
From modules/gosdk directory
```shell
./gosdk channel create -c ./fixtures/ConnectionProfile_org1.yaml --channelNamePrefix mychannel --channelConsortium SampleConsortium \
--channelOrgs org1,org2 --ordererName orderer.example.com --iterationCount 10 --iterationInterval 1s
```
#### Join
With the channel created above,we can specify which member to join which channel
From modules/gosdk directory
```shell
./gosdk channel join -c ./fixtures/ConnectionProfile_org1.yaml --channelNamePrefix mychannel \
--peers peer0.org1.example.com,peer1.org1.example.com --ordererName orderer.example.com --iterationCount 10 --iterationInterval 1s

./gosdk channel join -c ./fixtures/ConnectionProfile_org2.yaml --channelNamePrefix mychannel \
--peers peer0.org2.example.com,peer1.org2.example.com --ordererName orderer.example.com --iterationCount 10 --iterationInterval 1s

```

#### Update
With the channel created above,we can update the channel configurations member like batchTimeout/maxMessageCount/absoluteMaxBytes/anchorPeers
From modules/gosdk directory
```shell
./gosdk channel update -c ./fixtures/ConnectionProfile_org1.yaml --channelNamePrefix mychannel \
        --ordererOrgName ordererorg --ordererName orderer.example.com --peers peer0.org1.example.com\
        --batchTimeout 20s --maxMessageCount 200 --absoluteMaxBytes 103802353 --anchorPeers peer0.org1.example.com:7051 \
        --iterationCount 2 --iterationInterval 2s

./gosdk channel update -c ./fixtures/ConnectionProfile_org2.yaml --channelNamePrefix mychannel \
        --ordererOrgName ordererorg --ordererName orderer.example.com --peers peer0.org2.example.com\
        --anchorPeers peer0.org2.example.com:7051 \
        --iterationCount 2 --iterationInterval 2s

```

#### Query
With the channel created/joined/updated above,we can query the channel configurations member like batchTimeout/maxMessageCount/absoluteMaxBytes/anchorPeers
From modules/gosdk directory
```shell
./gosdk channel query -c ./fixtures/ConnectionProfile_org2.yaml --channelNamePrefix mychannel --peers peer0.org2.example.com --iterationCount 2 --iterationInterval 2s

./gosdk channel query -c ./fixtures/ConnectionProfile_org1.yaml --channelNamePrefix mychannel --peers peer0.org1.example.com --iterationCount 2 --iterationInterval 2s

```

### Chaincode
#### Install
With the fabric network created in [Privisioning a fabric network for testing](#provisioning-a-fabric-network-for-testing).
From modules/gosdk directory
```shell
./gosdk chaincode install -c ./fixtures/ConnectionProfile_org1.yaml --chaincodeNamePrefix samplecc- --chaincodeVersion v1 \
--path hfrd/chaincode/samplecc --peers peer0.org1.example.com,peer1.org1.example.com \
--iterationCount 2 --iterationInterval 1s
```
Use `./gosdk chaincode install -h` for help
#### Instantiate
After 1) creating channels 2). peers joining channels 3). installing chaincodes on peers,
we can instantiate chaincode on a specified channel
```shell
./gosdk chaincode instantiate -c ./fixtures/ConnectionProfile_org1.yaml --chaincodeNamePrefix samplecc- --chaincodeVersion v1 \
--channelName mychannel0 --path hfrd/chaincode/samplecc \
--peers peer0.org1.example.com --iterationCount 1
```
Use `./gosdk chaincode instantiate -h` for help
#### Invoke
```shell
./gosdk chaincode invoke  -c ./fixtures/ConnectionProfile_org1.yaml --chaincodeName samplecc-0 --channelName mychannel0 \
--chaincodeParams invoke,put,a,b --peers peer0.org1.example.com \
--iterationCount 100
```
Use `./gosdk chaincode invoke -h` for help. If `--peers` parameter is not specified,
then `gosdk` will randomly(round robin) pick a peer of the org specified in connection
profile to send proposal to.
