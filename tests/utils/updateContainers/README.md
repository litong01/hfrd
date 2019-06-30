# Update containers' vCPU and memory

## Prerequisites
- Node.js- tested with v8.11.3
- npm- tested with v5.6.0

This is only tested with ASH-PERFORMANCE cluster whose cluster manager is
authenticated with bluemix staging OIDC. We may need to adjust the OIDC
endpoint for production clusters.
## Quick start
From updateContainers directory
```shell
npm install
node app.js --help
```
You'll see the Usage help
```
Usage: node app.js --cmBaseUrl [url] --networkId [networkId] --apiKey [apiKey]
--nodeType [nodeType] --cpu [num] --memory [num]

Options:
  --help       Show help                                               [boolean]
  --version    Show version number                                     [boolean]
  --cmBaseUrl                                                         [required]
  --networkId                                                         [required]
  --apiKey                                                            [required]
  --nodeType                                                          [required]

Examples:
  node app.js --cmBaseUrl https://cm-ash-perf.4.secure.blockchain.ibm.com:444
  --networkId 31f53963e39342738d3804a2d0470db5 --apiKey
  OIDXvGPNea41CsXYEpn37GvoNHMib5unuQUtpWpps8hC --nodeType peer --cpu 1.2 --memory
  1024
```

Provide the required options as described above.

`--apiKey`: follow the below steps to get apiKey. Make sure your bluemix account
has the authority to access cluster manager to do the vCPU/memory change of containers
- Login https://console.stage1.bluemix.net/ with your IBM id
- Open menu Manage->Security->Platform API
- Create a new API key and copy the API key secret which will be used later

`--nodeType`: pick one from 'ca', 'orderer', 'peer', 'kafka', 'mysql', 'zookeeper'

`--memory`: unit: MB

`--cpu`: could be 1.1, 2, etc...

Other options should be self explanatory
## Examples
>> The --apiKey in the below examples are fake. You need to generate your own apiKey
1. change all 5 kafka containers cpu to be 1.1 and memory to be 2048MB
```shell
node app.js --cmBaseUrl https://cm-ash-perf.4.secure.blockchain.ibm.com:444
  --networkId 31f53963e39342738d3804a2d0470db5 --apiKey
  OIDXvGPNea41CsXYEpn37GvoNHMib5unuQUtpWpps8hC --nodeType kafka --cpu 1.1 --memory
  2048
```
Output:
```
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-kafka-11132a
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-kafka-11132b
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-kafka-11132c
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-kafka-11132e
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-kafka-11132d
containers list of type kafka after cpu/memory change
31f53963e39342738d3804a2d0470db5-kafka-11132a | 1.1 | 2048 MB
31f53963e39342738d3804a2d0470db5-kafka-11132b | 1.1 | 2048 MB
31f53963e39342738d3804a2d0470db5-kafka-11132e | 1.1 | 2048 MB
31f53963e39342738d3804a2d0470db5-kafka-11132c | 1.1 | 2048 MB
31f53963e39342738d3804a2d0470db5-kafka-11132d | 1.1 | 2048 MB
```
2. change all peers's vCPU to 1.5 and memory to 1024MB with a networkId
```shell
node app.js --cmBaseUrl https://cm-ash-perf.4.secure.blockchain.ibm.com:444
  --networkId 31f53963e39342738d3804a2d0470db5 --apiKey
  OIDXvGPNea41CsXYEpn37GvoNHMib5unuQUtpWpps8hC --nodeType peer --cpu 1.5 --memory
  1024
```
Output
```
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20376c
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20378a
containers list of type peer after cpu/memory change
31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20376c | 1.5 | 1024 MB
31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20378a | 1.5 | 1024 MB
```
If you issue the command twice, you may receive the below output which
means all the peers' vCPU and memory configuration is already in the expected
state and no need to make the change
```
31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20376c cpu: 1.5 memory: 1024 MB no need to change cpu and memory
31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20378a cpu: 1.5 memory: 1024 MB no need to change cpu and memory
containers list of type peer after cpu/memory change
31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20376c | 1.5 | 1024 MB
31f53963e39342738d3804a2d0470db5-fabric-peer-org1-20378a | 1.5 | 1024 MB
```

3. change all mysql containers' vCPU to be 2
```shell
node app.js --cmBaseUrl https://cm-ash-perf.4.secure.blockchain.ibm.com:444
  --networkId 31f53963e39342738d3804a2d0470db5 --apiKey
  OIDXvGPNea41CsXYEpn37GvoNHMib5unuQUtpWpps8hC --nodeType mysql --cpu 2
```
Output
```
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-mysql-b
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-mysql-e
successfully changed cpu/memory for 31f53963e39342738d3804a2d0470db5-mysql-d
containers list of type mysql after cpu/memory change
31f53963e39342738d3804a2d0470db5-mysql-e | 2 | 512 MB
31f53963e39342738d3804a2d0470db5-mysql-d | 2 | 512 MB
31f53963e39342738d3804a2d0470db5-mysql-b | 2 | 512 MB
```
