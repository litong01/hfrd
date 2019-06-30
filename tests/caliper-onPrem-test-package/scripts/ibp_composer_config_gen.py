import json
import os

HOME = os.environ['HOME']
HOME = './'
# read in connection profiles for both orgs
BCPLAN = "ibpep"
SCFileObject = {}
SCFileObject['composer'] = {}
SCFileObject['composer']['network'] = {}
SCFileObject['composer']['network']['x-type'] = "hlfv1"
SCFileObject['composer']['network']['timeout'] = 3000
SCFileObject['composer']['network']['version'] = "1.0.0"
SCFileObject['composer']['network']['tls'] = True
SCFileObject['composer']['network']['orderers'] ={}
SCFileObject['composer']['network']['certificateAuthorities'] ={} 
SCFileObject['composer']['network']['organizations'] =[]
SCFileObject['composer']['network']['peers'] ={}
SCFileObject['composer']['network']['channels'] ={}

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

for  tempOrderer in orderers: 
    SCFileObject['composer']['network']['orderers'][tempOrderer] =  {}
    SCFileObject['composer']['network']['orderers'][tempOrderer]['url']= ConnectionPdef['orderers'][tempOrderer]['url'] 
    SCFileObject['composer']['network']['orderers'][tempOrderer]['mspid'] = ""
    SCFileObject['composer']['network']['orderers'][tempOrderer]['domain'] = ""
    SCFileObject['composer']['network']['orderers'][tempOrderer]['msp'] =""
    SCFileObject['composer']['network']['orderers'][tempOrderer]['mspid'] =""
 
    tempHostname = ConnectionPdef['orderers'][tempOrderer]['url'] 
    tempHostname = tempHostname.strip()
    startp = tempHostname.find("://")
    endp = tempHostname.rfind(":")
    length = len(tempHostname)
    tempHostname = tempHostname[startp+3 : endp]

    SCFileObject['composer']['network']['orderers'][tempOrderer]['hostname'] = tempHostname
    SCFileObject['composer']['network']['orderers'][tempOrderer]['hosturl'] =""
    SCFileObject['composer']['network']['orderers'][tempOrderer]['cert'] ="network/fabric/"+BCPLAN+"/"+tempOrderer+"_orderer_tlscacerts.pem" 
    
    fo = open(tempOrderer+"_orderer_tlscacerts.pem","w")
    fo.write(ConnectionPdef['orderers'][tempOrderer]['tlsCACerts']['pem'])
    fo.close()

for tempPeer in ConnectionPdef['peers']:
    SCFileObject['composer']['network']['peers'][tempPeer] ={}
    SCFileObject['composer']['network']['peers'][tempPeer]['url']=ConnectionPdef['peers'][tempPeer]['url']
    SCFileObject['composer']['network']['peers'][tempPeer]['eventUrl']=ConnectionPdef['peers'][tempPeer]['eventUrl']
    SCFileObject['composer']['network']['peers'][tempPeer]['hostname']=""
    SCFileObject['composer']['network']['peers'][tempPeer]['cert']="network/fabric/"+BCPLAN+"/"+tempPeer+"_tlscacerts.pem"
    SCFileObject['composer']['network']['peers'][tempPeer]['channels']=[]
    tempTlsCaPEMFile = open(tempPeer+"_tlscacerts.pem","w")
    tempTlsCaPEMFile.write(ConnectionPdef['peers'][tempPeer]['tlsCACerts']['pem'])
    tempTlsCaPEMFile.close()
    

for  key in orgs: 
    with open(os.path.join(HOME, 'creds/ConnectionProfile_'+key+'.json'), 'r') as fr:
        tempConnection = json.load(fr)
    tempOrg={} 

    tempOrg["name"] =tempConnection['organizations'][key]['mspid']
    tempOrg["mspid"] =tempConnection['organizations'][key]['mspid']
    tempOrg["mspconfig"] = ""
    tempOrg["adminCert"] = "network/fabric/ibpep/creds/"+key+"admin/msp/signcerts/cert.pem"
    tempOrg["adminKey"] = "network/fabric/ibpep/creds/"+key+"admin/msp/keystore/priv.pem"
    tempOrg["certificateAuthorities"] = []
    tempOrg["peers"] = tempConnection['organizations'][key]['peers']
    for tempkey in tempConnection['certificateAuthorities']:
        tempOrg["certificateAuthorities"].append (tempkey)
        SCFileObject['composer']['network']['certificateAuthorities'][tempkey]= {}
        SCFileObject['composer']['network']['certificateAuthorities'][tempkey]['url']= \
           tempConnection['certificateAuthorities'][tempkey]['url']
        SCFileObject['composer']['network']['certificateAuthorities'][tempkey]['name']=\
           tempConnection['certificateAuthorities'][tempkey]['caName']
    SCFileObject['composer']['network']['organizations'].append(tempOrg) 

    SCFileObject['composer']['network']['certificateAuthorities'] 

    for tempChannel in tempConnection['channels']: 
        SCFileObject['composer']['network']['channels'][tempChannel]={}
        SCFileObject['composer']['network']['channels'][tempChannel]['config'] =""
        SCFileObject['composer']['network']['channels'][tempChannel]['mspconfig']=""
        SCFileObject['composer']['network']['channels'][tempChannel]['orderers']=tempConnection['channels'][tempChannel]['orderers']
        SCFileObject['composer']['network']['channels'][tempChannel]['cafile']= \
        "network/fabric/"+BCPLAN+"/"+tempConnection['channels'][tempChannel]['orderers'][0]+"_orderer_tlscacerts.pem"
        SCFileObject['composer']['network']['channels'][tempChannel]['peers']=[]
        for tempPeer in tempConnection['channels'][tempChannel]['peers']:
             SCFileObject['composer']['network']['channels'][tempChannel]['peers'].append(tempPeer)


SCFileObject['composer']['chaincodes'] = [{"id": "basic-sample-network", "version": "0.1.0", "path": "src/contract/composer", "orgs": ["PeerOrg1"], "loglevel": "INFO"}]
SCFileObject['composer']['cryptodir'] = "network/fabric/simplenetwork/crypto-config"
with open(os.path.join(HOME, 'ibpep_composer_net.json'), 'w') as f:
    json.dump(SCFileObject, f, sort_keys=False,indent=4)