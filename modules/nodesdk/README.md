## Node js SDK test modules
## 

## Pre-requisites:

1. Network is created and connection profile is saved and path to connection profile is known.
2. Required node packages are installed. npm install in the root directory.


## Test Modules:

## enrollAdmin.js 
- Purpose: Before you can use test modules to operate your peers, you need to enroll an admin user and sync it with your network (and, if you are using remote peers, you need to enroll your peer users). This module is used to generate admin certs.
- Usage: node enrollAdmin.js <org name> <path to connection profile>
- Output:
1. Signing certificates are created in the hfc-key-store folder.
2. Open the file named admin and copy the certificate inside the quotation marks after the certificate field.
3. Log in to your network on IBM Blockchain Platform, go to Network Monitor > Members > Certificates, and click Add Certificate. Give the certificate any name and paste the certificate copied in Step (ii). Click Restart to restart your peers.
 

## createChannel

\<To be developed\>

## joinChannel.js
- Purpose: Join org peers to a channel
- Config file: chan-join.yaml
- Usage: node joinChannel.js

#########################################################################################
\#Sample config in chan-join.yaml\#

Channels:\
   Join:\
     peerUrl: n509afa3a3f6d42dba96da98de9f43dde-org1-peer1.dev.blockchain.ibm.com \
     peerPort: 31002 \
     channelName: org2-channel-1 \
     orgName: org1 \
     ordererNum: 0 \
     isRemote: false \
     peerName: org1-peer1 \
     caCertFile: ./certs/cacert.pem \
     cprofDir: ./config \
     srvCertFile: ./certs/srvcert.pem \
#########################################################################################

\# Sample output \#
node scripts/joinChannel.js \
Looking in ./config for connection profile JSON files... \
Store path:/Users/nileshdeotale/go/src/hfrd/modules/nodesdk/hfc-key-store \
Successfully loaded admin from persistence \
Assigning transaction_id:  0f769534671c5cfae59c096d1e97cc908fd0e5903f098a572f4df3aee8e4414f \
Fetched genesis block. Sending channel join request... \
Peer at grpcs://n509afa3a3f6d42dba96da98de9f43dde-org1-peer1.dev.blockchain.ibm.com:31002 has successfully joined channel  org2-channel-1 

