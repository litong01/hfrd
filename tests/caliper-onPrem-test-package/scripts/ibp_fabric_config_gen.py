import json
import os

HOME = os.environ['HOME']
HOME = './'
# read in connection profiles for both orgs
BCPLAN = "ibpep"
SCFileObject = {}
SCFileObject['fabric'] = {}
SCFileObject['fabric']['network'] = {}
SCFileObject['fabric']['network']['orderer'] ={}

with open(os.path.join(HOME, 'creds/network.json'), 'r') as f:
    networkjson = json.load(f)
orgs = []
orderers =[]
channels = []
if isinstance(networkjson,dict):     
        for key in networkjson:
          orgs.append(key)
print (orgs)      

with open(os.path.join(HOME, 'creds/ConnectionProfile_'+orgs[0]+'.json'), 'r') as f:
    ConnectionPdef = json.load(f)

for key2 in ConnectionPdef['orderers']:
    orderers.append (key2) 
for key2 in ConnectionPdef['channels']:
    channels.append (key2)

SCFileObject['fabric']['network']['orderer']['url'] = ConnectionPdef['orderers'][orderers[0]]['url'] 
SCFileObject['fabric']['network']['orderer']['mspid'] = ""
SCFileObject['fabric']['network']['orderer']['domain'] = ""
SCFileObject['fabric']['network']['orderer']['msp'] =""

tempHostname = ConnectionPdef['orderers'][orderers[0]]['url'] 
tempHostname = tempHostname.strip()
startp = tempHostname.find("://")
endp = tempHostname.rfind(":")
length = len(tempHostname)
print (startp, endp, length)
tempHostname = tempHostname[startp+3 : endp]
print (tempHostname)

SCFileObject['fabric']['network']['orderer']['server-hostname'] = tempHostname

SCFileObject['fabric']['network']['orderer']['tls_cacerts'] = "network/fabric/"+BCPLAN+"/ep_orderer_tls_cacerts.pem" 

fo = open("ep_orderer_tls_cacerts.pem","w")
fo.write(ConnectionPdef['orderers'][orderers[0]]['tlsCACerts']['pem'])
fo.close()
    
for  key in orgs: 
    with open(os.path.join(HOME, 'creds/ConnectionProfile_'+key+'.json'), 'r') as fr:
        tempConnection = json.load(fr)
    tempPeers = tempConnection['organizations'][key]['peers'] 
    print (tempPeers)  
    SCFileObject['fabric']['network']['org_'+key] = {}
    SCFileObject['fabric']['network']['org_'+key]['name'] = tempConnection['organizations'][key]['mspid']
    SCFileObject['fabric']['network']['org_'+key]['mspid'] = tempConnection['organizations'][key]['mspid']
    SCFileObject['fabric']['network']['org_'+key]['ca'] = {}
    tempcaname = tempConnection['organizations'][key]['certificateAuthorities'][0]
    SCFileObject['fabric']['network']['org_'+key]['ca']['url'] = tempConnection['certificateAuthorities'][tempcaname]['url']
    SCFileObject['fabric']['network']['org_'+key]['ca']['name'] = tempcaname
    SCFileObject['fabric']['network']['org_'+key]['user'] ={}
    SCFileObject['fabric']['network']['org_'+key]['user']['key'] ="network/fabric/"+BCPLAN+"/creds/"+key+"admin/msp/keystore/priv.pem"
    SCFileObject['fabric']['network']['org_'+key]['user']['cert'] ="network/fabric/"+BCPLAN+"/creds/"+key+"admin/msp/signcerts/cert.pem"
    for tempPeer in tempPeers :
        SCFileObject['fabric']['network']['org_'+key]['peer_'+tempPeer] = {}
        SCFileObject['fabric']['network']['org_'+key]['peer_'+tempPeer]['requests'] = tempConnection['peers'][tempPeer]['url']
        SCFileObject['fabric']['network']['org_'+key]['peer_'+tempPeer]['events'] = tempConnection['peers'][tempPeer]['eventUrl']
        tempHostname = tempConnection['peers'][tempPeer]['url']
        tempHostname = tempHostname.strip()
        startp = tempHostname.find("://")
        endp = tempHostname.rfind(":")
        tempHostname = tempHostname[startp+3 : endp]
      
        SCFileObject['fabric']['network']['org_'+key]['peer_'+tempPeer]['server-hostname'] = tempHostname
        SCFileObject['fabric']['network']['org_'+key]['peer_'+tempPeer]['tls_cacerts'] = "network/fabric/"+BCPLAN+"/"+tempPeer+"_tlscacerts.pem"
        
        tempTlsCaPEMFile = open(tempPeer+"_tlscacerts.pem","w")
        tempTlsCaPEMFile.write(tempConnection['peers'][tempPeer]['tlsCACerts']['pem'])
        tempTlsCaPEMFile.close()
    
    SCFileObject['fabric']['context'] = {}
    if len(channels) == 0:
        channels.append("channel1") 
    SCFileObject['fabric']['context']['open']= 'channel1'
    SCFileObject['fabric']['context']['query'] = 'channel1'
    
    SCFileObject['fabric']['channel'] = []
    defultChannel ={}
    defultChannel['name']  = 'channel1'
    defultChannel['organizations'] =['org_PeerOrg1']
    defultChannel['deployed'] = True
    defultChannel['config'] = ""
    SCFileObject['fabric']['channel'].append(defultChannel) 

    SCFileObject['fabric']['chaincodes'] = [{"id": "simple", "path": "contract/fabric/simple/go", "language":"golang", "version": "v0", "channel": "channel1"}]
    SCFileObject['fabric']['cryptodir'] = "network/fabric/simplenetwork/crypto-config"
with open(os.path.join(HOME, 'ibpep_fabric.json'), 'w') as f:
    json.dump(SCFileObject, f, sort_keys=False,indent=4)