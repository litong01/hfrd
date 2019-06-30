HFRD PTE Test Package
 ===============

* [Introduction](#introduction)
* [Support envs](#Support-envs)
* [How to run the pte test package](#How-to-run-pte-test-package)
    * [Prerequisites](#prerequisites)
    * [Configure before run the test package](#Configure-before-run-the-test-package)
        * [hfrd_test.cfg: main configuration file](#hfrd_test.cfg)
        * [conf/channels.json: configure the channels](#conf/channels.json)
        * [conf/hosts: configure the hosts](#conf/hosts)
    * [Start to run the test package](#Start-to-run-the-test-package)

## Introduction
HFRD PTE test package uses PTE(Performance Traffic Engine) to drive regression tests on blockchian networks.

PTE is a powerful blockchain traffic engine.You can find more details in:
- PTE official repo:
  https://github.com/hyperledger/fabric-test/tree/master/tools/PTE

## Support blockchain environments
Currently pte-onPrem-test-package can support:
- cm environment: POK Clusters
- starter plan : bxstaging and bxproduction


## How to run pte test package
### Prerequisites
PTE test package could work on both MacOS and Linux(x86 and s390x).But for on-prem environment (like in POK Clusters),the test package must be located in machine that can access the POK intranet.

To run the golang api server, clone this project under
&lt;somedir&gt;/src,

        git clone git@github.ibm.com:IBMCode/hfrd.git

then include &lt;somedir&gt; in your GOPATH
environment, then change directory to hfrd/tests/pte-OnPrem-test-package

        cd $GOPATH/src/hfrd/tests/pte-OnPrem-test-package

### Prerequisites
PTE test package could work on both MacOS and Linux(x86 and s390x).But for on-prem environment (like in POK Clusters),the test package must be located in machine that can access the POK intranet.

### Configure before run the test package
Before you start to run the test package, you need to make sure that you have changed the configuration files accordingly.
You need to change three config files now:
- hfrd_test.cfg
- conf/hosts (only for on-prem users)
- conf/channels_EP.json or conf/channels_SP.json

#### hfrd_test.cfg

    - apiuser: the user name of HFRD API server
    - apiserverhost: host of HFRD API server
    - apiversion: version of HFRD API server
    - apiserverbaseuri: base url of HFRD API server
    - env: test environment,currently support [ bxstaging, bxproduction, cm ]
  The rest is only used in cm env

    - loc: location of blockchian clusters,like POKM148_ZBC4
    - numOfOrgs: number of orgs you want to create
    - numOfPeers: number of peers per org you want to create
    - ledgerType: the ledger type you want configure.You can choose levelDB or couchdb accordingly

#### conf/hosts
The hosts file is used to build the mapping between private ip and domain names

    10.20.36.61    pokm148-lpzbc4a.3.secure.blockchain.ibm.com
    10.20.36.62    pokm148-lpzbc4b.3.secure.blockchain.ibm.com
    10.20.36.63    pokm148-lpzbc4c.3.secure.blockchain.ibm.com
    10.20.36.64    pokm148-lpzbc4d.3.secure.blockchain.ibm.com
    10.20.36.65    pokm148-lpzbc4e.3.secure.blockchain.ibm.com

#### conf/channels_${}.json
The channels.json file is used to configure the channels you want to use in test.
For example: This configuration is applying two channels 'channel1' and 'channel2'.And 'channel1' contains two members 'PeerOrg1' and 'PeerOrg2'.'channel2' contains one member 'PeerOrg2'.


{    

      "channels": [
        {
            "name":"channel1",
            "members": [
                    "PeerOrg1"
            ],
            "batchSize" : { "messageCount" : 100,
                            "absoluteMaxBytes" : 103809024,
                            "preferredMaxBytes" : 524288
                        },
            "batchTimeout": "10s",
            "channelRestrictionMaxCount" : "150"
        },
        {
            "name": "channel2",
            "members": [
                "PeerOrg1"
            ],
            "batchSize" : { "messageCount" : 100,
                            "absoluteMaxBytes" : 103809024,
                            "preferredMaxBytes" : 524288
                        },
            "batchTimeout": "10s",
            "channelRestrictionMaxCount" : "150"
        }
    ]
}

### Start to run the test package
From pte-OnPrem-test-package directory, issue

        ./hfrd_test.sh
The script will
- load configuration files: hfrd_test.cfg conf/hosts(if cm) conf/channels.json
- build the docker image : pte-fab:latest
- create a new container and run the pte-OnPrem-test-package/docker-entrypoint.sh
