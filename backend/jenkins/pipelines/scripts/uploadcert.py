
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

inputNetworkDir = ""
tempConnectionProfile = {}
MAXRETRYNUM = 10


def uploadCert4Org(key, httpAuth, connectprofile, heliosBaseUrl):
    certFilePath = inputNetworkDir+'/'+key+'/users/Admin@' + \
        key+'/msp/signcerts/'+'Admin@'+key+'-cert.pem'
    certfile = open(certFilePath, 'r')
    certString = certfile.read()
    certBody = {}
    certBody["msp_id"] = tempConnectionProfile['organizations'][key.lower()
                                                                ]['mspid']
    certBody["adminCertName"] = "hfrdCert_"+key
    certBody["adminCertificate"] = certString
    certBody["peer_names"] = tempConnectionProfile['organizations'][key.lower()
                                                                    ]['peers']
    certBody["SKIP_CACHE"] = True

    heliosCertsUrl = heliosBaseUrl+'/certificates'
    tempHttpResp = requests.post(heliosCertsUrl, data=json.dumps(
        certBody), auth=authx, headers=heliosPostHeaders,verify=False)
    if tempHttpResp.status_code == 200:
        print ("\nupload hfrdCert for ["+key+"] successed")
        #print('\nResponse content:'+tempHttpResp.content)
    else:
        print(
            "\nupload hfrdCert for ["+key+"] failed, returned code=" + str(tempHttpResp.status_code))
        print('\nResponse content:'+tempHttpResp.content)

    heliosCertsGetUrl = heliosBaseUrl+'/certificates/fetch'
    getCertsBody = {}
    getCertsBody["peer_names"] = tempConnectionProfile['organizations'][key.lower()
                                                                        ]['peers']
    tempHttpResp = requests.post(heliosCertsGetUrl, data=json.dumps(
        getCertsBody), auth=authx, headers=heliosPostHeaders,verify=False)
    print('\nResponse content:'+tempHttpResp.content)


def restartPeers4Org(key, httpAuth, connectprofile, heliosBaseUrl):
    print('\nRestart peers for org:' + str(key))
    peerRunning = True
    for peerItem in tempConnectionProfile['organizations'][key.lower()]['peers']:
        startPeerUrl = heliosBaseUrl+'/nodes/'+peerItem+'/start'
        stopPeerUrl = heliosBaseUrl+'/nodes/'+peerItem+'/stop'
        statusUrl = heliosBaseUrl+'/nodes/status'
        tempHttpResp = requests.get(
            statusUrl, auth=authx, headers=heliosGetHeaders,verify=False)
        jsonContent = json.loads(tempHttpResp.content)
        currentStatus = jsonContent[peerItem]['status']
        if currentStatus == "running":
            peerRunning = True
            tempHttpResp = requests.post(
                stopPeerUrl, auth=authx, headers=heliosPostHeaders,verify=False)
            if tempHttpResp.status_code == 200:
                print(peerItem + " is stopping...")
                maxRetry = MAXRETRYNUM
                while (peerRunning and maxRetry > 0):
                    tempHttpResp = requests.get(
                        statusUrl, auth=authx, headers=heliosGetHeaders,verify=False)
                    maxRetry = maxRetry - 1
                    jsonContent = json.loads(tempHttpResp.content)
                    currentStatus = jsonContent[peerItem]['status']
                    print(peerItem + ' status='+currentStatus)
                    if currentStatus == "running":
                        peerRunning = True
                    if currentStatus == "exited":
                        peerRunning = False
                        time.sleep(1)
        else:
            peerRunning = False
        if (peerRunning == False) and (currentStatus == "exited"):
            tempHttpResp = requests.post(
                startPeerUrl, auth=authx, headers=heliosPostHeaders,verify=False)
            if tempHttpResp.status_code == 200:
                print(peerItem + " is starting...")
                maxRetry = MAXRETRYNUM
                while ((not currentStatus == "running") and maxRetry > 0):
                    tempHttpResp = requests.get(
                        statusUrl, auth=authx, headers=heliosGetHeaders,verify=False)
                    maxRetry = maxRetry - 1
                    jsonContent = json.loads(tempHttpResp.content)
                    currentStatus = jsonContent[peerItem]['status']
                    print(peerItem + ' status='+currentStatus)
                    if currentStatus != "running":
                        time.sleep(1)
            else:
                print ("starting "+peerItem +
                       " failed, response :"+tempHttpResp.content)


def makePeersRunning4Org(key, httpAuth, connectprofile, heliosBaseUrl):
    print('\nChecking peers status for org:' + str(key))
    peerRunning = False
    for peerItem in tempConnectionProfile['organizations'][key.lower()]['peers']:
        peerRunning = False
        startPeerUrl = heliosBaseUrl+'/nodes/'+peerItem+'/start'
        statusUrl = heliosBaseUrl+'/nodes/status'
        tempHttpResp = requests.get(
            statusUrl, auth=authx, headers=heliosGetHeaders,verify=False)
        jsonContent = json.loads(tempHttpResp.content)
        currentStatus = jsonContent[peerItem]['status']
        if currentStatus == "running":
            peerRunning = True
            print(" peer["+peerItem + "] is running" )
        else:
            maxRetry = MAXRETRYNUM
            while ((not peerRunning) and maxRetry > 0):
                tempHttpResp = requests.get(
                    statusUrl, auth=authx, headers=heliosGetHeaders,verify=False)
                maxRetry = maxRetry - 1
                jsonContent = json.loads(tempHttpResp.content)
                currentStatus = jsonContent[peerItem]['status']
                print(peerItem + ' status='+currentStatus)
                if currentStatus == "running":
                    peerRunning = True
                    print(" peer["+peerItem + "] is running" ) 
                    break
                elif currentStatus == "exited":
                    tempHttpResp = requests.post(
                        startPeerUrl, auth=authx, headers=heliosPostHeaders,verify=False)
                    peerRunning = False
                    time.sleep(1)
                    print(" waiting peer["+peerItem + "] start up" ) 
                elif currentStatus == "Unknown":
                    peerRunning = False
                    time.sleep(1)
                    print(" waiting peer["+peerItem + "] start up" ) 
                else:
                    print("peer["+peerItem + "] status is unexpected: "+ currentStatus )  
            if (peerRunning ==False) :
               print("peer["+peerItem + "] is not running, and time out..." )






# Final sysch all channels in the network
# tempChannelName = 'channel1'
# synchChannelUrl=heliosBaseUrl+networkId+'/channels/'+tempChannelName+'/sync'
# tempHttpResp= requests.post(synchChannelUrl,auth=authx,headers=getHeaders)
#print('\n synchChannel=\n'+tempHttpResp.text)
try:
    opts, args = getopt.getopt(sys.argv[1:], "hd:", ["help", "networkdir="])
except getopt.GetoptError as e:
    print 'uploadcert.py -d <networkdir> '
    print(e)
    sys.exit(2)
for opt, arg in opts:
    if opt == '-h':
        print 'uploadcert.py -d <networkdir>'
    elif opt in ("-d", "--networkdir"):
        inputNetworkDir = arg
if not os.path.exists(inputNetworkDir):
    print (inputNetworkDir + " directory not existed")
    sys.exit(2)

networkJson = {}
orgsInNetwork = []
httpClient = None
heliosGetHeaders = {'Accept': 'application/json',
                    'Content-Type': 'application/json'}
heliosPostHeaders = {'Content-Type': 'application/json',
                     'Accept': 'application/json'}
with open(inputNetworkDir + '/network.json', 'r') as f:
    networkJson = json.load(f)
orgItem_network = "DefaultNetworkId"
for keyofOrg in networkJson:
    orgsInNetwork.append(keyofOrg)
    orgItem_key = networkJson[keyofOrg]['key']
    orgItem_network = networkJson[keyofOrg]['network_id']
    orgItem_secret = networkJson[keyofOrg]['secret']
    orgItem_url = networkJson[keyofOrg]['url']

    authx = HTTPBasicAuth(orgItem_key, orgItem_secret)
    heliosBaseUrl = orgItem_url+'/api/v1/networks/'+orgItem_network
    # get  connection profile
    with open(inputNetworkDir+'/'+keyofOrg+'/connection.json', 'r') as f:
        tempConnectionProfile = json.load(f)
    try:
        # cert name format : {username}@{key}-cert.pem
        makePeersRunning4Org(keyofOrg, authx, tempConnectionProfile, heliosBaseUrl)
        uploadCert4Org(keyofOrg, authx, tempConnectionProfile, heliosBaseUrl)
        restartPeers4Org(keyofOrg, authx, tempConnectionProfile, heliosBaseUrl)

    except Exception as e:
        print(e)

print ("\nuploaded certificates  and restarted peers for network:"+orgItem_network)
