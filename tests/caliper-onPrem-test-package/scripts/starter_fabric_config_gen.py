import json
import os

HOME = os.environ['HOME']
HOME = './'
# read in connection profiles for both orgs
BCPLAN = "starter"
SCFileObject = {}
SCFileObject['fabric'] = {}
SCFileObject['fabric']['network'] = {}
SCFileObject['fabric']['network']['orderer'] ={}
# starter plan use apikeys.json
# ibp plan use network.json
with open(os.path.join(HOME, 'apikeys.json'), 'r') as f:
    networkjson = json.load(f)
orgs = []
orderers =[]
channels = []
if isinstance(networkjson,dict):     
        for key in networkjson:
          orgs.append(key)
print (orgs)  

# starter plan Connection profile format {orgname}ConnectionProfile.json
# ibp plan Connection Profile format is ConnectionProfile{orgname}.json
with open(os.path.join(HOME, orgs[0]+'ConnectionProfile.json'), 'r') as f:
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

SCFileObject['fabric']['network']['orderer']['tls_cacerts'] = "network/fabric/"+BCPLAN+"/creds/starter_orderer_tls_cacerts.pem" 

fo = open("starter_orderer_tls_cacerts.pem","w")
fo.write(ConnectionPdef['orderers'][orderers[0]]['tlsCACerts']['pem'])
fo.close()
# starter plan connection profile filename format '{orgname}ConnectionProfile.json'   
for  key in orgs: 
    with open(os.path.join(HOME, key+'ConnectionProfile.json'), 'r') as fr:
        tempConnection = json.load(fr)
    tempPeers = tempConnection['organizations'][key]['peers'] 
    print (tempPeers)  
    SCFileObject['fabric']['network'][key] = {}
    SCFileObject['fabric']['network'][key]['name'] = tempConnection['organizations'][key]['mspid']
    SCFileObject['fabric']['network'][key]['mspid'] = tempConnection['organizations'][key]['mspid']
    SCFileObject['fabric']['network'][key]['ca'] = {}
    tempcaname = tempConnection['organizations'][key]['certificateAuthorities'][0]
    SCFileObject['fabric']['network'][key]['ca']['url'] = tempConnection['certificateAuthorities'][tempcaname]['url']
    SCFileObject['fabric']['network'][key]['ca']['name'] = tempcaname
    SCFileObject['fabric']['network'][key]['user'] ={}
    SCFileObject['fabric']['network'][key]['user']['key'] ="network/fabric/"+BCPLAN+"/creds/"+key+"admin/msp/keystore/priv.pem"
    SCFileObject['fabric']['network'][key]['user']['cert'] ="network/fabric/"+BCPLAN+"/creds/"+key+"admin/msp/signcerts/cert.pem"
    for tempPeer in tempPeers :
        SCFileObject['fabric']['network'][key]['peer_'+tempPeer] = {}
        SCFileObject['fabric']['network'][key]['peer_'+tempPeer]['requests'] = tempConnection['peers'][tempPeer]['url']
        SCFileObject['fabric']['network'][key]['peer_'+tempPeer]['events'] = tempConnection['peers'][tempPeer]['eventUrl']
        tempHostname = tempConnection['peers'][tempPeer]['url']
        tempHostname = tempHostname.strip()
        startp = tempHostname.find("://")
        endp = tempHostname.rfind(":")
        tempHostname = tempHostname[startp+3 : endp]
      
        SCFileObject['fabric']['network'][key]['peer_'+tempPeer]['server-hostname'] = tempHostname
        SCFileObject['fabric']['network'][key]['peer_'+tempPeer]['tls_cacerts'] = "network/fabric/"+BCPLAN+"/creds/"+tempPeer+"_tlscacerts.pem"
        
        tempTlsCaPEMFile = open(tempPeer+"_tlscacerts.pem","w")
        tempTlsCaPEMFile.write(tempConnection['peers'][tempPeer]['tlsCACerts']['pem'])
        tempTlsCaPEMFile.close()

    SCFileObject['fabric']['context'] = {}
    SCFileObject['fabric']['context']['open']  = 'defaultchannel'
    SCFileObject['fabric']['context']['query'] = 'defaultchannel'
    
    SCFileObject['fabric'] ['channel'] = []
    defaultChan = {}
    defaultChan['name']  = 'defaultchannel'
    defaultChan['organizations'] = ['org1']
    defaultChan['deployed'] = True
    defaultChan['config'] = ""

    SCFileObject['fabric'] ['channel'].append(defaultChan)
    SCFileObject['fabric']['cryptodir'] = "network/fabric/simplenetwork/crypto-config"
   
with open(os.path.join(HOME, 'starter_fabric.json'), 'w') as f:
    json.dump(SCFileObject, f, sort_keys=False,indent=4)

