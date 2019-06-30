#!/usr/bin/env python
# coding=utf-8
import httplib
import requests
import json
from requests.auth import HTTPBasicAuth
import os
import copy
import time
import yaml
import shutil
import getopt
import sys

inputNetworkFile = ""
inputNetworkFlag = False
outputDir = '/Users/sunhongwei/data/tempnet/'
outputDirFlag = False
allConnProfileMap = {}
allOrganizations = {}


def processOrdererSection(orderers, newSCFileObject):
    print ("\nprocessing orderers")
    for key in orderers:
        ordererTLSCAPath = outputDir + 'orderers/'+key+'/tls/'
        if not os.path.exists(ordererTLSCAPath):
            os.makedirs(ordererTLSCAPath)
        fo = open(ordererTLSCAPath + '/ca.pem', "w")
        fo.write(newSCFileObject['orderers'][key]['tlsCACerts']['pem'])
        fo.close()
        newSCFileObject['orderers'][key]['tlsCACerts']['path'] = ordererTLSCAPath+'ca.pem'
        newSCFileObject['orderers'][key]['tlsCACerts'].pop('pem')


def processPeersSection(peers, newSCFileObject):
    print ("\nprocessing peers")
    tempOrgName = ""
    for key in peers:
        for tempOrg in orgs:
            tempOrgName = "otherOrgs"
            if key in newSCFileObject['organizations'][tempOrg]['peers']:
                tempOrgName = tempOrg
                break
        peerTLSCAPath = outputDir + tempOrgName+'/Peers/'+key+'/tls/'
        if not os.path.exists(peerTLSCAPath):
            os.makedirs(peerTLSCAPath)
        fo = open(peerTLSCAPath+'/ca.pem', "w")
        fo.write(newSCFileObject['peers'][key]['tlsCACerts']['pem'])
        fo.close()
        newSCFileObject['peers'][key]['tlsCACerts']['path'] = peerTLSCAPath+'ca.pem'
        newSCFileObject['peers'][key]['tlsCACerts'].pop('pem')


def processOrgsSection(orgs, newSCFileObject):
    print ("\nprocessing organizations")
    for key in orgs:
        orgSingedCertPath = key+'/msp/singedCerts/'
        if not os.path.exists(outputDir + orgSingedCertPath):
            os.makedirs(outputDir + orgSingedCertPath)
        fo = open(outputDir + orgSingedCertPath+'/cert.pem', "w")
        fo.write(newSCFileObject['organizations']
                 [key]['signedCert']['pem'])
        fo.close()
        newSCFileObject['organizations'][key]['signedCert']['path'] = orgSingedCertPath+'cert.pem'
        newSCFileObject['organizations'][key]['cryptoPath'] = key + \
            '/users/{username}@'+key+'/msp'
        newSCFileObject['organizations'][key]['signedCert'].pop('pem')
        newSCFileObject['organizations'][key.lower()] = copy.deepcopy(
            newSCFileObject['organizations'][key])
        newSCFileObject['organizations'].pop(key)


def processCAsSections(cas, newSCFileObject):
    print ("\nprocessing certificateAuthorities")
    for key in cas:
        caTLSCAPath = outputDir + 'certificateAuthorities/'+key+'/tls/'
        if not os.path.exists(caTLSCAPath):
            os.makedirs(caTLSCAPath)
        fo = open(caTLSCAPath+'/ca.pem', "w")
        fo.write(
            newSCFileObject['certificateAuthorities'][key]['tlsCACerts']['pem'])
        fo.close()
        newSCFileObject['certificateAuthorities'][key]['tlsCACerts']['path'] = caTLSCAPath+'ca.pem'
        newSCFileObject['certificateAuthorities'][key]['tlsCACerts'].pop(
            'pem')


def removeRegistarOfCAsSections(cas, newSCFileObject):
    for key in cas:
        newSCFileObject['certificateAuthorities'][key].pop('registrar')


def enrollUserForOrg(cas, newSCFileObject, keyofOrg):
    caenrollId = newSCFileObject['certificateAuthorities'][cas[0]
                                                           ]['registrar'][0]['enrollId']
    orginalCaUrl = newSCFileObject['certificateAuthorities'][cas[0]]['url']
    startp = orginalCaUrl .find("://")
    caenrollUrl = orginalCaUrl[startp+3:]
    caenrollSecret = newSCFileObject['certificateAuthorities'][cas[0]
                                                               ]['registrar'][0]['enrollSecret']
    tlsCaCertfiles = outputDir + \
        'certificateAuthorities/'+cas[0]+'/tls/ca.pem'
    tempMspPath = outputDir + keyofOrg+'/users/'+caenrollId+'@'+keyofOrg+'/msp'
    mspPath = tempMspPath
    cmdRet = 0
    cmdRet = os.system('fabric-ca-client enroll  --tls.certfiles '+tlsCaCertfiles
                       + '  -u  https://'+caenrollId+":"+caenrollSecret+'@'+caenrollUrl + ' --mspdir '+mspPath)
    if (cmdRet != 0):
        print ("\n enroll user failed,return code=" + str(cmdRet))
    else:
        print ("\n enroll admin user of "+keyofOrg + " successed")
        os.rename(mspPath+'/signcerts/cert.pem', mspPath +
                  '/signcerts/admin@'+keyofOrg+'-cert.pem')


try:
    opts, args = getopt.getopt(sys.argv[1:], "hn:o:", [
                               "help", "networkfile=", "outputdir="])
except getopt.GetoptError as e:
    print 'zipgen.py -n <networkfile> -o <outputdir>'
    print(e)
    sys.exit(2)
for opt, arg in opts:
    if opt == '-h':
        print 'zipgen.py -n <networkfile> -o <outputdir>'
    elif opt in ("-n", "--networkfile"):
        inputNetworkFile = arg
        inputNetworkFlag = True
    elif opt in ("-o", "--outputdir"):
        outputDir = arg+'/'
        outputDirFlag = True
if not os.path.isfile(inputNetworkFile):
    print (inputNetworkFile + "file not existed")
    sys.exit(2)
if not (inputNetworkFlag and outputDirFlag):
    print ("parameters are not correct, please follow the command format like:")
    print 'zipgen.py -n <networkfile> -o <outputdir>'
    sys.exit(2)

networkJson = {}
orgsInNetwork = []
httpClient = None
heliosGetHeaders = {'Accept': 'application/json',
                    'Content-Type': 'application/json'}
heliosPostHeaders = {'Content-Type': 'application/json',
                     'Accept': 'application/json'}
with open(inputNetworkFile, 'r') as f:
    networkJson = json.load(f)
orgItem_network = "DefaultNetworkId"
for keyofOrg in networkJson:
    orgsInNetwork.append(keyofOrg)
    orgItem_key = networkJson[keyofOrg]['key']
    orgItem_network = networkJson[keyofOrg]['network_id']
    orgItem_secret = networkJson[keyofOrg]['secret']
    orgItem_url = networkJson[keyofOrg]['url']
    # get Initial connection profile
    heliosNetworkBaseUrl = orgItem_url+'/api/v1/networks/'+orgItem_network+'/'
    authx = HTTPBasicAuth(orgItem_key, orgItem_secret)
    try:
        connProfileUrl = heliosNetworkBaseUrl+'connection_profile'
        print(
            'getting orignal connection profile for organization['+keyofOrg + '] from: '+connProfileUrl)
        tempHttpResp = requests.get(
            connProfileUrl, auth=authx, headers=heliosGetHeaders)
        tempConnectionProfile = json.loads(tempHttpResp.text)
        orgs = []
        orderers = []
        channels = []
        orgsinCP = []
        peers = []
        cas = []
        # print(tempConnectionProfile)
        for tempKey in tempConnectionProfile['peers']:
            peers.append(tempKey)
        for tempKey in tempConnectionProfile['orderers']:
            orderers.append(tempKey)
        for tempKey in tempConnectionProfile['channels']:
            channels.append(tempKey)
        for tempKey in tempConnectionProfile['organizations']:
            orgs.append(tempKey)
        for tempKey in tempConnectionProfile['certificateAuthorities']:
            cas.append(tempKey)

        newSCFileObject = copy.deepcopy(tempConnectionProfile)
        # orderers section
        processOrdererSection(orderers, newSCFileObject)
        # peers section
        processPeersSection(peers, newSCFileObject)
        # organizations section
        processOrgsSection(orgs, newSCFileObject)
        # client section
        newSCFileObject['client']['organization'] = newSCFileObject['client']['organization'].lower()
        newSCFileObject['client']['cryptoconfig'] = {}
        newSCFileObject['client']['cryptoconfig']['path'] = outputDir

        # certificates section
        processCAsSections(cas, newSCFileObject)
        # enroll admin user
        enrollUserForOrg(cas, newSCFileObject, keyofOrg)
        # remove certificates registrar section
        removeRegistarOfCAsSections(cas, newSCFileObject)

        allConnProfileMap[keyofOrg] = {}
        allConnProfileMap[keyofOrg]['hfrdCP'] = copy.deepcopy(newSCFileObject)
        allConnProfileMap[keyofOrg]['jsonFilePath'] = os.path.join(
            outputDir+orgItem_key+'/', orgItem_key+'_ConnProf.json')
        allConnProfileMap[keyofOrg]['yamlFilePath'] = os.path.join(
            outputDir+orgItem_key+'/', orgItem_key+'_ConnProf.yaml')
        allOrganizations[keyofOrg.lower()] = copy.deepcopy(
            newSCFileObject['organizations'][newSCFileObject['client']['organization']])

    except Exception as e:
        print(e)
    finally:
        print("\nconnetion profile generated for orgnization:"+orgItem_key)

# output network.json and connection profiles for each org
with open(os.path.join(outputDir, 'network.json'), 'w') as f:
    json.dump(networkJson, f, sort_keys=False, indent=4)
    f.close()
for eachOrg in allConnProfileMap:
    allConnProfileMap[eachOrg]['hfrdCP']['organizations'] = allOrganizations
    with open(allConnProfileMap[eachOrg]['yamlFilePath'], 'w') as f:
        yaml.safe_dump(allConnProfileMap[eachOrg]['hfrdCP'],
                       default_flow_style=False, stream=f, indent=4)
        f.close()
    with open(allConnProfileMap[eachOrg]['jsonFilePath'], 'w') as f:
        json.dump(allConnProfileMap[eachOrg]
                  ['hfrdCP'], f, sort_keys=False, indent=4)
        f.close()
# put network.json into zip package
shutil.make_archive(outputDir+'CP_'+orgItem_network, 'zip', outputDir)
print ("\n network["+orgItem_network + "]connetion profiles and all certs save into file:\n" +
       'CP_'+orgItem_network+'.zip')
