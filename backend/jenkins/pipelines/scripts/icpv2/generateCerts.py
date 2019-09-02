#!/usr/bin/env python
# coding=utf-8
import os
import sys
import json
import yaml
import subprocess

def loadConfigContent(configFile):
    config_content = open(configFile)
    config_dict = {}
    for lines in config_content:
        items = lines.split('=', 1)
        config_dict[items[0]] = items[1]
    return config_dict

def loadJsonContent(jsonFile):
    with open(jsonFile, 'r') as f:
        temtplateDict = json.load(f)
    f.close()
    return temtplateDict


def generatePeerSection(templateContent, peerName, orgName, proxyIp):
    strCmd = "export PATH=$PATH:./bin; kubectl get svc " + peerName + "-service | grep NodePort |  awk -F '[[:space:]:/]+' '{print $6}'"
    templateContent['url'] = 'grpcs://' + proxyIp + ':' + os.popen(strCmd).read().strip()
    strCmd = "export PATH=$PATH:./bin; kubectl get svc " + peerName + "-service | grep NodePort |  awk -F '[[:space:]:/]+' '{print $8}'"
    templateContent['eventUrl'] = 'grpcs://' + proxyIp + ':' + os.popen(strCmd).read().strip()
    templateContent['grpcOptions']['ssl-target-name-override'] = proxyIp
    templateContent['tlsCACerts']['path'] = templateContent['tlsCACerts']['path'].replace('orgname',orgName)
    return templateContent


def generateOrdererSection(templateContent, ordererName, ordererorgName, proxyIp):
    strCmd = "export PATH=$PATH:./bin ; kubectl get svc " + ordererName + "-service | grep NodePort |  awk -F '[[:space:]:/]+' '{print $6}'"
    templateContent['url'] = 'grpcs://' + proxyIp + ':' + os.popen(strCmd).read().strip()
    templateContent['grpcOptions']['ssl-target-name-override'] = proxyIp
    templateContent['tlsCACerts']['path'] = templateContent['tlsCACerts']['path'].replace('ordererorg', ordererorgName)
    return templateContent

def generateOrgSection(templateContent,orgName,peers,orgType):
    templateContent['mspid'] = orgName
    templateContent['cryptoPath'] = templateContent['cryptoPath'].replace('orgname', orgName)
    if orgType == 'peerorg':
        for peer in peers:
            if peer.split('.')[1] == orgName:
                templateContent['peers'].append(orgName + peer.split('.')[0])
    return templateContent

def generateConnectionProfiles(networkspec):
    peerorg_names = []
    ordererorg_names = []
    peers = networkspec['network']['peers']
    for orderer_object in networkspec['network']['orderers']:
        ordererorg_names.append(orderer_object.split('.')[1])
    ordererorg_names = list(set(ordererorg_names))
    for peer_object in peers:
        peerorg_names.append(peer_object.split('.')[1])
    peerorg_names = list(set(peerorg_names))

    # generate collection profile for each peer organization
    for org in peerorg_names:
        connection_template = loadJsonContent('./templates/connection_template.json')
        # Load client
        connection_template['client']['organization'] = org
        # Load organizations including peer orgs and orderer org
        for org_name in peerorg_names:
            org_template = loadJsonContent('./templates/org_template.json')
            connection_template['organizations'][org_name] = generateOrgSection(org_template, org_name, peers, 'peerorg')
        org_template = loadJsonContent('./templates/org_template.json')
        for ordererorg_name in ordererorg_names:
            connection_template['organizations'][ordererorg_name] = generateOrgSection(org_template, ordererorg_name,'','ordererorg')

        # Load peers
        print peers
        for peer in peers:
            org_name = peer.split('.')[1]
            peer_name = peer.split('.')[0]
            peer_template = loadJsonContent('./templates/peer_template.json')
            peer_name = org_name + peer_name
            connection_template['peers'][peer_name] = generatePeerSection(peer_template, peer_name, org_name, networkspec['icp']['url'].split(':')[0])
        # Load orderers
        orderer_template = loadJsonContent('./templates/orderer_template.json')
        for ordererorg in networkspec['network']['orderers']:
            orderer_num = ordererorg.split('.')[0]
            ordererorg_name = ordererorg.split('.')[1]
            for orderer_index in range(int(orderer_num)):
                orderer_index += 1
                orderer_name = ordererorg_name + 'orderer' + str(orderer_index)
                connection_template['orderers'][orderer_name] = generateOrdererSection(orderer_template, orderer_name, ordererorg_name, networkspec['icp']['url'].split(':')[0])
        # write out connection file
        with open(networkspec['work_dir'] + '/crypto-config/' + org + '/connection.json', 'w') as f:
            print('\nWriting connection file for ' + str(org) + ' - ' + f.name)
            json.dump(connection_template, f, indent=4)
        f.close()
        with open(networkspec['work_dir'] + '/crypto-config/' + org + '/connection.yml', 'w') as f:
            print('\nWriting connection file for ' + str(org) + ' - ' + f.name)
            yaml.safe_dump(connection_template, f, allow_unicode=True)
        f.close()


# certsPath = /opt/src/scripts/icp/keyfiles
def generateCertificatesPackage(networkspec):
    certsPath = networkspec['work_dir'] + '/crypto-config/'
    # restructure msp dir
    mspCommand = 'cd '+ certsPath + ' && mkdir -p orgname/users/Admin@orgname/msp && cp -rf orgname/admin/* orgname/users/Admin@orgname/msp/'
    tlsCommand = 'cd ' + certsPath + ' && mkdir -p orgname/tlsca && cp -rf orgname/msp/tlscacerts/*.pem orgname/tlsca/ca.pem'
    peerorg_names = []
    ordererorg_names = []

    for orderer_object in networkspec['network']['orderers']:
        ordererorg_names.append(orderer_object.split('.')[1])
    ordererorg_names = list(set(ordererorg_names))
    for peer_object in networkspec['network']['peers']:
        peerorg_names.append(peer_object.split('.')[1])
    peerorg_names = list(set(peerorg_names))
    for org in peerorg_names:
        os.system(mspCommand.replace('orgname', org))
        os.system(tlsCommand.replace('orgname', org))
        os.rename(certsPath + "/" + org + "/users/Admin@" + org + "/msp/signcerts/cert.pem", certsPath + "/" + org + "/users/Admin@" + org + "/msp/signcerts/Admin@" + org + "-cert.pem")
    # ordererorg
    for ordererorg_name in ordererorg_names:
        os.system(mspCommand.replace('orgname', ordererorg_name))
        os.system(tlsCommand.replace('orgname', ordererorg_name))
        os.rename(certsPath + "/" + ordererorg_name + "/users/Admin@" + ordererorg_name + "/msp/signcerts/cert.pem", certsPath + "/" + ordererorg_name + "/users/Admin@" + ordererorg_name + "/msp/signcerts/Admin@" + ordererorg_name + "-cert.pem")
