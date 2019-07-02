# HFRD Operations

<!-- toc -->

- [Test plan top level parameters](#test-plan-top-level-parameters)
- [Common operation parameters](#common-operation-parameters)
  * [CHANNEL_CREATE](#channel_create)
  * [CHANNEL_JOIN](#channel_join)
  * [CHANNEL_QUERY](#channel_query)
  * [CHANNEL_UPDATE](#channel_update)
  * [CHAINCODE_INSTALL](#chaincode_install)
  * [CHAINCODE_INSTANTIATE](#chaincode_instantiate)
  * [CHAINCODE_INVOKE](#chaincode_invoke)
  * [CHAINCODE_INVOKE (private data chaincode)](#chaincode_invoke-private-data-chaincode)
    + [Syntax of each parameter pattern](#syntax-of-each-parameter-pattern)
  * [EXECUTE_COMMAND](#execute_command)

<!-- tocstop -->

**HFRD** use yaml file to design the test plan. The test plan supports eight different operations to drive blockchain networks,including:

	* CHANNEL_CREATE : Used to create channels 
	* CHANNEL_JOIN : Used to join specified peers into channels
	* CHANNEL_QUERY: Used to query channel configuration,like BatchTimeout/BatchSize/AnchorPeers.(under verification)
	* CHANNEL_UPDATE : Used to update channel configurations,like BatchTimeout/BatchSize/AnchorPeers.(under verification)
	* CHAINCODE_INSTALL : Used to install chaincode to specified peers
	* CHAINCODE_INSTANTIATE : Used to instantiate chaincode to specified peers
	* CHAINCODE_INVOKE : Used to send traffic to fabric networks
	* CHAINCODE_INVOKE (private data chaincode) : Show how to configure private data parameters in 	`CHAINCODE_INVOKE`
	* EXECUTE_COMMAND: Used to execute shell command (a bash script or other binary files,like `ls` `go` `gosdk`)
**HFRD** users can define their own test plans by using these operations,then upload the test plan yaml file to **HFRD** .**HFRD** will help submit test plans and generate k8s job according to different operations inner test plan. You can find an example test plan file [here](testplan.yml)

Next will show how to **define different operation** in testplan.yml.

### Test plan top level parameters
There are few parameters that you can use to control how your test produce logs files, collect log files, metrics and if your test should continue if one test fails.

1. `continueAfterFail` **:** Yes or no to indicates if the test should continue or fail after one test failed
2. `logLevel` **:** the level of the logging for the test. Possible values are `ERROR`,`INFO`,`DEBUG`
3. `saveLog` **:** This defines if the log files should be saved at the end of the test. When set to `false`, the log files produced will not be saved. When set to `true`, the log will be saved.
4. `collectFabricMetrics` **:** This defines if the test should collect fabric network metrics, this capability is only available when you have a fabric network using fabric 1.4.0 or newer release. The value should be either true or false.

### Common operation parameters
You may find that each operation in the testplan has some common parameters,including:  
1. `name` **:** The name of this operation.    
2. `operation` **:** The operation name of this section.  
3. `iterationCount` **:**  This parameter indicates the loop count of this operatoin.For example: If `iterationCount:10` and `operation:CHANNEL_CREATE`, then hfrd will help you create **10** channels with same `channelNamePrefix` .
For `CHAINCODE_INVOKE`,this will be different.  
4. `iterationInterval` **:** The interval between each iteration.  
5. `retryCount` **:** The max retry times if operation fails.  
6. `loadSpread` **:** HFRD use kubernetes job to run the operations. `loadSpread` **:** defines how many kubernetes pods will run this operation **in parallel**.    
7. `logLevel` **:** setting log level of test modules. Currently supporting `DEBUG`, `INFO` and `ERROR`, all case insensitive.     
8. `ignoreErrors` **:** If `ignoreErrors: true`, then gosdk will continue to run next iteration even errors are detected.    
9. `concurrencyLimit` **:** setting test module sending requests concurrency limit. If user
sets concurrencyLimit to 10, test modules will start **AT MOST** 10 goroutines to send requests.
Users can also set `concurrencyLimit` to 1 to send requests sequentially. `CHAINCODE_INSTANTIATE` and
`EXECUTE_COMMAND` operations will always set `concurrencyLimit` to 1 even if user sets a different value.
10. `delayTime` **:** This parameter indicates the delay time before the first
iteration starts. This parameter provides a user tool so that parallel operations can be started at a different time. This is necessary for dependent operations, for example, if a chaincode invoke has to start after another chaincode finishes, then the chaincode invoke should be delayed for a period time before its execution. This parameter takes time period in seconds as its value. For example, `delayTime: 60s`, this means that the operation should wait 60 seconds before its first execution.

#### CHANNEL_CREATE
```yaml
  - name: "create-channel"
    operation: "CHANNEL_CREATE"
    iterationCount: 20
    iterationInterval: 0s
    loadSpread: 1
    ignoreErrors: false
    parameters:
      applicationCapability: V1_4
      connectionProfile: orga
      channelNamePrefix: "mychannel"
      prefixOffset: 0 # optional
      channelConsortium: "FabricConsortium"
      channelOrgs: "orga,orgb"
      ordererName: orderer1st-orgc
```

You can also use `channelNameList` instead of `channelNamePrefix` and `prefixOffset`.
`channelNameList` and `channelNamePrefix` are mutual exclusive

```yaml
  - name: "create-channel"
    operation: "CHANNEL_CREATE"
    iterationCount: 20
    iterationInterval: 0s
    loadSpread: 1
    ignoreErrors: false
    parameters:
      applicationCapability: V1_4
      connectionProfile: orga
      channelNameList: mychannel0,mychannel2,mychannel5,mychannel11
      channelConsortium: "FabricConsortium"
      channelOrgs: "orga,orgb"
      ordererName: orderer1st-orgc
```


`CHANNEL_CREATE ` will help to create fabric channels according to the parameters,including:  
1. `applicationCapability `: Used to set application channel capability.This is required if you want to use `Private data collections` in fabric version > 1.1.       
2. `connectionProfile`: Specify org name or full path of connection profile here. 	
For example, if you set `connectionProfile: org1`,test modules will automatically generate the full path of connection profile `/fabric/keyfiles/org1/connection.yml` that belongs to this organization for you. If you set `connectionProfile: /fabric/keyfiles/org1/connection.yml`, the modules will directly use this path to look for connection profile.   
3. `channelNamePrefix`: The prefix of channel name.For example,if you set the `channelNamePrefix:"mychannel"` and `iterationCount:10`, then HFRD will help create 10 channels with name from `mychannel0` to `mychannel9`.   
4. `prefixOffset`: The offset of `channelNamePrefix`. This is optional and default number is 0. When `prefixOffset: 3` and `iterationCount: 10`,then HFRD will create 10 channels with name from `mychannel3 (mychannel + prefixOffset)` to `mychannel12 (mychannel + prefixOffset + iterationCount - 1 )`.       
5. `channelConsortium`: The channel consortium to be used in channel creation.In IBP networks, you can use `SampleConsotium`.In cello networks, you can use `FabricConsotium`.   
6. `channelOrgs `: The orgs that allowed to join the channels created in this operation.  
7. `ordererName `: The orderer to be used to create channels.


#### CHANNEL_JOIN
```yaml
  - name: "join-orga-peers-to-channel"
    operation: "CHANNEL_JOIN"
    iterationCount: 10
    iterationInterval: 1s
    loadSpread: 1
    parameters:
      connectionProfile: orga
      channelNamePrefix: "mychannel"
      prefixOffset: 0 # optional
      peers: "peer1st-orga"
      ordererName: orderer1st-orgc
```
You can also use `channelNameList` instead of `channelNamePrefix` and `prefixOffset`.
`channelNameList` and `channelNamePrefix` are mutual exclusive

```yaml
  - name: "join-orga-peers-to-channel"
    operation: "CHANNEL_JOIN"
    iterationCount: 1
    iterationInterval: 1s
    loadSpread: 1
    parameters:
      connectionProfile: orga
      channelNameList: mychannel0,mychannel2,mychannel5,mychannel11
      peers: "peer1st-orga"
      ordererName: orderer1st-orgc
```

`iterationCount` should be set to 1 to avoid joining peers in a specific channel for multiple times
`CHANNEL_JOIN ` will help to join peers into channels according to the parameters,including:  
1. `connectionProfile`: Specify org name or full path of connection profile here. 	
For example, if you set `connectionProfile: org1`,test modules will automatically generate the full path of connection profile `/fabric/keyfiles/org1/connection.yml` that belongs to this organization for you. If you set `connectionProfile: /fabric/keyfiles/org1/connection.yml`, the modules will directly use this path to look for connection profile.  
2. `channelNamePrefix`: The prefix of channel name.For example,if you set the `channelNamePrefix:"mychannel"` and `iterationCount:10`, then HFRD will help join the specified peers into channels from`mychannel0` to `mychannel9`.   
3. `prefixOffset`: The offset of `channelNamePrefix`. This is optional and default number is 0. When `prefixOffset: 3` and `iterationCount: 10`,then HFRD will join peers into 10 channels with name from `mychannel3 (mychannel + prefixOffset)` to `mychannel12 (mychannel + prefixOffset + iterationCount - 1 )`.       
4. `peers `: The peers will be joined to channels.Be careful,only the peers that belong to the **connectionProfile client.organization**  can be joined by using this operation.If you want to join another org's peers into channel,you need to add another operation with that org's connection profile.    
5. `ordererName `: The orderer to be used to join channels.
6. `channelNameList` : Peers will join all the channels in `channelNameList`

#### CHANNEL_QUERY
```yaml
  - name: "query-channel-config"
    operation: "CHANNEL_QUERY"
    iterationCount: 10
    iterationInterval: 1s
    loadSpread: 1
    parameters:
      connectionProfile: orga
      channelName: "mychannel"
      peers: "peer1st-orga"
```

`CHANNEL_QUERY ` will help to query channel configurations like BatchSize/BatchTimeout/Anchor peers according to the parameters,including:  
1. `connectionProfile`: Specify org name or full path of connection profile here. 	 
2. `channelName`: The channel name
3. `peers`: The peers that will be used to query channels.`peers` must have been joined into channels.

#### CHANNEL_UPDATE
```yaml
  - name: "updae-channel-config"
    operation: "CHANNEL_UPDATE"
    iterationCount: 10
    iterationInterval: 1s
    loadSpread: 1
    parameters:
      connectionProfile: orga
      channelNamePrefix: "mychannel"
      channelNameList: mychannel0,hfrdchannel,samplechannel
      peers: "peer1st-orga"
      ordererOrgName: ordererorg
      ordererName: orderer.example.com
      batchTimeout: 1s
      maxMessageCount: 100
      absoluteMaxBytes: 103802353
      anchorPeers: peer0.org1.example.com:7051
      ordererAddressesAction: replace/add/remove
      ordererAddresses: orderer0.example.com,orderer1.example.com
```

You can also use `channelNameList` instead of `channelNamePrefix` and `prefixOffset`. channelNameList and channelNamePrefix are mutual exclusive.
Same like `CHANNEL_JOIN`. 


`CHANNEL_UPDATE ` will help to update channel configurations like BatchSize/BatchTimeout/Anchor peers according to the parameters,including:  
1. `connectionProfile`: Specify org name or full path of connection profile here. 	 
2. `channelNamePrefix`: The prefix of channel name.For example,if you set the `channelNamePrefix:"mychannel"` and `iterationCount:10`, then HFRD will help query channel configurations from`mychannel0` to `mychannel9`.      
3.  `peers`: The peers that will be used to update channels.`peers` must have been joined into these channels.     
4. `ordererOrgName `: The orderer org name.   
5. `ordererName`: The orderer name that you want to use when update channels.  
6. `batchTimeout ` `maxMessageCount ` `absoluteMaxBytes ` `anchorPeers ` are the channel configurations currently supported in hfrd. Remember that the anchorPeers should be the list of peers' url which contains the IP and port.   
7. `ordererAddressesAction `: Used to specify the operation type when update orderer addresses.Currently support `replace`,`remove`,`add`.    
 `ordererAddressesAction: replace` will replace the orderer addresses with  orderer addresses specified in`ordererAddresses` .   
 `ordererAddressesAction: remove` will remove the orderer addresses that are specified in `ordererAddresses` from the channel configuration.             
`ordererAddressesAction: add` will add the orderer addresses that specified in `ordererAddresses` into the channel configuration.   
8. `ordererAddresses ` : The orderer addresses you want to use for update.


#### CHAINCODE_INSTALL
```yaml
  - name: "install-chaincode"
    operation: "CHAINCODE_INSTALL"
    iterationCount: 10
    iterationInterval: 1s
    loadSpread: 1
    parameters:
      connectionProfile: orga
      chaincodeNamePrefix: "mysamplecc-"
      prefixOffset: 0 # optional
      chaincodeVersion: "v1"
      path: "chaincode/samplecc" # cc src code path relative to $GOPATH
      peers: "peer1st-orga"
```
`CHAINCODE_INSTALL ` will help to install chaincodes to peers according to the parameters,including:  
1. `connectionProfile`: Specify org name or full path of connection profile here. 	
For example, if you set `connectionProfile: org1`,test modules will automatically generate the full path of connection profile `/fabric/keyfiles/org1/connection.yml` that belongs to this organization for you. If you set `connectionProfile: /fabric/keyfiles/org1/connection.yml`, the modules will directly use this path to look for connection profile.    
2. `chaincodeNamePrefix `: The prefix of chaincode name.For example,if you set the `chaincodeNamePrefix:"mysamplecc-"` and `iterationCount:10`, then HFRD will help install chaincodes from `mysamplecc-0` to `mysamplecc-9` to the specified peers.    
3. `prefixOffset`: The offset of `chaincodeNamePrefix `. This is optional and default number is 0. When `prefixOffset: 3` and `iterationCount: 10`,then HFRD will install chiancodes from `mysamplecc-3 (mysamplecc- + prefixOffset)` to `mysamplecc-12 (mysamplecc- + prefixOffset + iterationCount - 1 )`.   
4. `chaincodeVersion `: The chaincode version     
5. `path` : The path of the chaincode.This shoule be same dir structure as the uploaded chaincode.tgz file.For example,if your tgz file's dir structure is `chaincode/samplecc/samplecc.go`, you should set `path` to `chaincode/samplecc`.   
6. `peers `: The peers that chaincode will be installed on.Be careful,only the peers that belong to the **connectionProfile client.organization**  can be installed by using this operation.If you want to install chaincode on another org's peers,you need to add another operation with that org's connection profile. 


#### CHAINCODE_INSTANTIATE
```yaml
  - name: "instantiate-chaincode"
    operation: "CHAINCODE_INSTANTIATE"
    iterationCount: 1
    iterationInterval: 1s
    loadSpread: 1
    parameters:
      connectionProfile: orga
      chaincodeName: "mysamplecc"
      chaincodeVersion: "v1"
      path: "chaincode/samplecc"
      channelNamePrefix: "mychannel-"
      peers: "peer1st-orga"
      policyStr: "OR ('Org1MSP.member','Org2MSP.member')"
      collectionsConfigPath: "chaincode/marbles_private/collections_config.json"
```
`CHAINCODE_INSTANTIATE ` will help to instantiate chaincodes to peers according to the parameters,including:  
1. `connectionProfile`: Specify org name or full path of connection profile here. 	
For example, if you set `connectionProfile: org1`,test modules will automatically generate the full path of connection profile `/fabric/keyfiles/org1/connection.yml` that belongs to this organization for you. If you set `connectionProfile: /fabric/keyfiles/org1/connection.yml`, the modules will directly use this path to look for connection profile.  
2. `chaincodeName`: The chaincode name
3. `prefixOffset`: The offset of `chaincodeNamePrefix `. This is optional and default number is 0. When `prefixOffset: 3` and `iterationCount: 10`,then HFRD will instantiate chiancodes from `mysamplecc-3 (mysamplecc- + prefixOffset)` to `mysamplecc-12 (mysamplecc- + prefixOffset + iterationCount - 1 )`.     
4. `chaincodeVersion `: The chaincode version 
5. `path` : The path of the chaincode.This shoule be same dir structure as the uploaded chaincode.tgz file.For example,if your tgz file's dir structure is `chaincode/samplecc/samplecc.go`, you should set `path` to `chaincode/samplecc`.  
6. `channelNamePrefix` : Indicate which channel the chaincode will be instantiated in.For example,if you set the `channelNamePrefix:"mychannel-"` and `iterationCount:10`, then HFRD will help instantiate chaincode in channels `mychannel-0` to `mychannel-9`.
7. `peers `: The peers that chaincode will be instantiated on.Be careful if you are using fabric network created by `cello`.
8. `policyStr`: Optional parameter, set this parameter when specific endorsement policy will be used instead of default one. Valid endorsement policy string example: "OR ('Org1MSP.member','Org2MSP.member')", "AND ('Org1MSP.member','Org2MSP.member')" etc.   
9. `collectionsConfigPath `: Optional parameter, set this parameter when the chaincode is used to operate private data.You can put the collections-config.json in chaincode.tgz, then use relative path for this parameter setting `collectionsConfigPath : chaincode/{chaincode_name}/collections_config.json`. Or you can specify the absolute path of collections_config.json `collectionsConfigPath : /fabric/src/chaincode/marbles_private/collections_config.json`. 
#### CHAINCODE_INVOKE
```yaml
 - name: "execute-cc"
   operation: "CHAINCODE_INVOKE"
   iterationCount: 100
   iterationInterval: 2r
   loadSpread: 5
   parameters:
      connectionProfile: orga
      channelName: "mychannel0"
      chaincodeName: "mysamplecc-0"
      queryOnly: false
      peers: "peer1st-orga"
      serviceDiscovery: true
      fabricVersion: "1.2"
      chaincodeParams:
        - type: "literal"
          value: "invoke"
        - type: "literal"
          value: "put"
        - type: "stringPattern"
          value: "account[0-9]"
        - type: "stringPattern"
          value: "account[a-z]"
        - type: "intRange"
          min: "0"
          max: "100"
        - type: "payloadRange"
          min: "1024"
          max: "2048"
        - type: "sequentialString"
          value: "marbles*"
      logLevel: INFO
      concurrencyLimit: 100
```
`CHAINCODE_INVOKE ` will help to invoke chaincodes according to the parameters,including:  
1. `connectionProfile`: Specify org name or full path of connection profile here. 	
For example, if you set `connectionProfile: org1`,test modules will automatically generate the full path of connection profile `/fabric/keyfiles/org1/connection.yml` that belongs to this organization for you. If you set `connectionProfile: /fabric/keyfiles/org1/connection.yml`, the modules will directly use this path to look for connection profile.     
2. `channelName` : The channel name that transactions will be invoked on.
3. `chaincodeName`: The chaincode that will be used.   
4. `queryOnly`: If `queryOnly:true`,the operation will only query the ledger.If `queryOnly:false`,the operation will invoke chaincode and make ledger commits.
3. `peers `: The peers that will be used for endorsement or service discovery
4. `chaincodeParams `: The chaincode parameters to be used when invoke chaincode.Totally we support 4 parameter patterns, including `literal/stringPattern/intRange/payloadRange`.  
5. `serviceDiscovery`: default to `false`. When set to true,
'cc invoke' module will do service discovery with the first peer in peers
list to discover a peer group to send proposal to. service discovery feature
is supported from v1.2
6. `fabricVersion`: default to `1.1`. Should be set according to the
Application Capability specified in application channel. Currently hfrd
test modules support 1.1, 1.2, 1.3 and 1.4

#### CHAINCODE_INVOKE (private data chaincode)
```yaml
 - name: "execute-cc"
   operation: "CHAINCODE_INVOKE"
   iterationCount: 100
   iterationInterval: 2r
   loadSpread: 5
   parameters:
      connectionProfile: orga
      channelName: "mychannel0"
      chaincodeName: "marble_private-0"
      queryOnly: false
      peers: "peer1st-orga,peer1st-orgb"  # service discovery will be leveraged when peers param is empty
      chaincodeParams:
        - type: "literal"
          value: "initMarble"
      transientMapParams:
        - key: "marble"
          value:
            - type: "sequentialString"
              key: "name"
              value: "marble_pvt"
            - type: "literal"
              key: "color"
              value: "blue"
            - type: "literal"
              key: "size"
              value: "30"
            - type: "literal"
              key: "price"
              value: "100"
            - type: "literal"
              key: "owner"
              value: "nilesh1"
        - key: "marble_owner"
          value:
            - type: "sequentialString"
              key: "name"
              value: "test"
            - type: "literal"
              key: "owner"
              value: "nilesh"
      logLevel: INFO
      concurrencyLimit: 100
```
There are big differences when invoke private data chaincodes.Take `marbles private chaincode` [link](https://github.ibm.com/IBMCode/hfrd/tree/master/chaincode/marbles_private) as an example:
1) When instantiate `marbles_private` chaindode,you should provide the `collectionsConfigPath`(see more details [here](https://github.ibm.com/IBMCode/hfrd/blob/master/docs/hfrd-operations.md#chaincode_instantiate))
2) When invoke `marbles_private` chaincode,you should provide `transientMapParams`. The above `transientMapParams` contains two `key-value` pairs. Take the first one for an example:

```
        - key: "marble".
          value:
            - type: "sequentialString"
              key: "name"
              value: "marble"
            - type: "literal"
              key: "color"
              value: "blue"

```
Obviously the key is `marble`.`value`follows the rules we generate `chaincodeParams`(Support `literal` `intRange` `stringPattern` `sequentialString`  `payloadRange`),but you need provide another `key` to make sure transientMap is valid map type.The result would be
`{"marble":{"name":"marble$iterationIndex","color":"blue"}}`


**Note**
For `invoke` operation ,we support two different running modes.	 
1. Running with fixed total transaction count `n`. `n ` must be a fixed integer.For example: if `n` is `10000`,then the operation will send `10000` transactions totally.   
Just set	
```
	iterationCount: n
```	 
in test plan to use this mode.      
2. Running with fixed total time `t`.For example:
if `t` is `1h5m2s`,then the operation will send transactions for 1 hour 4 minutes 2 seconds. 
Just set    
```
	iterationCount: t
```	 
in test plan to use this mode.

##### Syntax of each parameter pattern

	- type:"literal"
	  value:"invoke"
param type is "literal".value is "invoke". This kind of param will return the value "invoke" as the chaincode parameter

	- type:"stringPattern"
	  value:"account[0-9]"
In type `stringPattern`, the `value` should be a legal `regular expression` string.   
For this example, param `type` is "stringPattern" and the `value` is "account[0-9]",so HFRD will generate a string according to the `regular exparession` **account[0-9]**  as the chaincode parameter.

	- type:"sequentialString"
	  value:"*marbles*"
Param type is `sequentialString`.Value is `marbles*`. `sequentialString ` will return a string by replace `*` in value `marbles*` with `loopIndex`.`loopIndex` is the current iteration index.For example,if `iterationCount` is `100` and current loopIndex (**Note** loopIndex will range from `0` to `iterationCount - 1` ) is `10`,then hfrd will return string `10marbles10` for current iteration.

	- type: "intRange"
	  min: "0"
	  max: "100"
For type `intRange`, hfrd will return a random integer which will be within `0-100`

	- type: "payloadRange"
	  min: "1024"
	  max: "2048"
For type `payloadRange`, hfrd will generate a random bytes array whose size will be within `1024-2048`

#### EXECUTE_COMMAND
```
  - name: "execute-command"
    operation: "EXECUTE_COMMAND"
    loadSpread: 5
    container: email4tong/myown
    parameters:
      commandName: "/fabric/keyfiles/runPTE.sh"
      commandParams:
      # Totally we have 4 different parameter patterns:
      # 'literal' will return the value as string
      # 'stringPattern' will generate a string based on the value(regular expression)
      # 'intRange' will generate a integer which ranges from min to max
      # 'payloadRange' will generate a bytes array which size ranges from min to max
        - type: "literal"
          value: "version"
```
`EXECUTE_COMMAND ` will help to execute shell command(a script file or other binay file). In this operation, you can specify your own docker image as the runner box by changing parameter `container` (default is `hfrd/gosdk:latest`):    
1. `commandName`: The command name that would be used to execute.For script file ,this should be the script path. For other binary command like `ls` `go` or even `gosdk`, this should be the binary path.
2. `commandParams` : The command parameters to be used when execute the above command .Totally we support 4 parameter patterns, including `literal/stringPattern/intRange/payloadRange`.   

For example:	 
If you want to run a script file like `runPTE.sh` on docker image `email4tong/myown:latest` by using this operation,you can follow the below steps:		

**Step 1:** Build image `email4tong/myown:latest` based on image `hfrd/gosdk:latest`().Directly **jump to step 5** if you have put your scripts into the custom image.

**Step 2:** Get your `certs.tgz`(For ibp environment, this would be `ibmcert.tar.gz`) and untar the package.

```
tar zxvf certs.tgz
```		

After you untar the package,you will get a folder named `keyfiles`	

**Step 3:**  Put your scripts into the `keyfiles`

```
	cp ${Path of runPTE.sh} $(pwd)/keyfiles/
```
**Step 4:** Regerate the `certs.tgz`

```
	rm certs.tgz		
	tar zcvf certs.tgz keyfiles
```
**Step 5:** Upload the `certs.tgz` in HFRD UI `HFRD_IP:9090/v1/{userid}/ui/moduletest`	![](https://github.ibm.com/IBMCode/hfrd/blob/master/docs/images/managetests.png)			

**Step 6:** Update the test plan

```
  - name: "execute-command"
    operation: "EXECUTE_COMMAND"
    loadSpread: 5
    container: email4tong/myown
    parameters:
      commandName: "/fabric/keyfiles/runPTE.sh"
      commandParams:
      	 - type: "literal"
          value: "--help"
```

`commandName ` will be the full path of script `runPTE.sh`.As `certs.tgz` will be untared to directory `/farbic/keyfiles`, the path of `runPTE.sh` will be `/fabric/keyfiles/runPTE.sh`.	

**Step 7:** Upload other required materials. Then you are ready to run test!
