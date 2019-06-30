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


def generatePeerSection(templateContent, peerName, orgName):
    strCmd = "kubectl get svc " + peerName + " | grep NodePort |  awk -F '[[:space:]:/]+' '{print $6}'"
    templateContent['url'] = 'grpcs://' + proxy_ip + ':' + os.popen(strCmd).read().strip()
    strCmd = "kubectl get svc " + peerName + " | grep NodePort |  awk -F '[[:space:]:/]+' '{print $8}'"
    templateContent['eventUrl'] = 'grpcs://' + proxy_ip + ':' + os.popen(strCmd).read().strip()
    templateContent['grpcOptions']['ssl-target-name-override'] = peerName
    templateContent['tlsCACerts']['path'] = templateContent['tlsCACerts']['path'].replace('orgname',orgName)
    return templateContent

def generateOrdererSection(templateContent,ordererName):
    strCmd = "kubectl get svc " + network_name + "-orderer-orderer | grep NodePort |  awk -F '[[:space:]:/]+' '{print $6}'"
    templateContent['url'] = 'grpcs://' + proxy_ip + ':' + os.popen(strCmd).read().strip()
    templateContent['grpcOptions']['ssl-target-name-override'] = ordererName
    templateContent['tlsCACerts']['path'] = templateContent['tlsCACerts']['path'].replace('ordererorg', ordererorg_name)
    return templateContent

def generateOrgSection(templateContent,orgName,orgType):
    templateContent['mspid'] = orgName
    templateContent['cryptoPath'] = templateContent['cryptoPath'].replace('orgname', orgName)
    if (orgType != 'ordererorg'):
        for peer_index in range(int(peers_per_org)):
            templateContent['peers'].append(network_name + '-' + orgName + '-peer' + str(peer_index))
    return templateContent

def generateConnectionProfiles(certs_path):
    # generate collection profile for each peer organization
    for index in range(int(num_of_orgs)):
        connection_template = loadJsonContent('./template/connection_template.json')
        connection_template['name'] = network_name
        org = org_name_prefix + str(index)
        # Load client
        connection_template['client']['organization'] = org

        # Load organizations including peer orgs and orderer org
        for org_index in range(int(num_of_orgs)):
            org_name = org_name_prefix + str(org_index)
            org_template = loadJsonContent('./template/org_template.json')
            connection_template['organizations'][org_name] = generateOrgSection(org_template, org_name,'peer')
        org_template = loadJsonContent('./template/org_template.json')
        connection_template['organizations'][ordererorg_name] = generateOrgSection(org_template, ordererorg_name,'ordererorg')

        # Load peers
        for org_index in range(int(num_of_orgs)):
            for peer_index in range(int(peers_per_org)):
                org_name = org_name_prefix + str(org_index)
                for peer_index in range(int(peers_per_org)):
                    peer_template = loadJsonContent('./template/peer_template.json')
                    peer_name = network_name + '-' + org_name + '-peer' + str(peer_index)
                    connection_template['peers'][peer_name] = generatePeerSection(peer_template, peer_name, org_name)

        # Load orderers
        orderer_template = loadJsonContent('./template/orderer_template.json')
        connection_template['orderers'][network_name + '-orderer-orderer'] = generateOrdererSection(orderer_template, network_name + '-orderer-orderer')
        # write out connection file
        with open(certs_path + '/' + org + '/connection.json', 'w') as f:
            print('\nWriting connection file for ' + str(org) + ' - ' + f.name)
            json.dump(connection_template, f, indent=4)
        f.close()
        with open(certs_path + '/' + org + '/connection.yml', 'w') as f:
            print('\nWriting connection file for ' + str(org) + ' - ' + f.name)
            yaml.safe_dump(connection_template, f, allow_unicode=True)
        f.close()


# certsPath = /opt/src/scripts/icp/keyfiles
def generateCertificatesPackage(certsPath):
    # restructure msp dir
    mspCommand = 'cd '+ certsPath + ' && mkdir -p orgname/users/Admin@orgname/msp && cp -rf orgname/admin/* orgname/users/Admin@orgname/msp/'
    tlsCommand = 'cd ' + certsPath + ' && mkdir -p orgname/tlsca && cp -rf orgname/admin/tlscacerts/*-tlsca.pem orgname/tlsca/ca.pem'
    for org_index in range(int(num_of_orgs)):
        org = org_name_prefix + str(org_index)
        os.system(mspCommand.replace('orgname', org))
        os.system(tlsCommand.replace('orgname', org))
        os.rename(certsPath + "/" + org + "/users/Admin@" + org + "/msp/signcerts/cert.pem", certsPath + "/" + org + "/users/Admin@" + org + "/msp/signcerts/Admin@" + org + "-cert.pem")
    # ordererorg
    os.system(mspCommand.replace('orgname', ordererorg_name))
    os.system(tlsCommand.replace('orgname', ordererorg_name))
    os.rename(certsPath + "/" + ordererorg_name + "/users/Admin@" + ordererorg_name + "/msp/signcerts/cert.pem", certsPath + "/" + ordererorg_name + "/users/Admin@" + ordererorg_name + "/msp/signcerts/Admin@" + ordererorg_name + "-cert.pem")


config_file_path = sys.argv[1]
certs_path = sys.argv[2]

config = loadConfigContent(config_file_path)
network_name = config['NAME'].strip()
num_of_orgs = config['NUM_ORGS']
peers_per_org = config['PEERS_PER_ORG']
namespace = config['NAMESPACE'].strip()

org_name_prefix = network_name + 'org'
ordererorg_name = network_name + 'ordererorg'
# # Get Proxy_IP
proxy_ip = config['PROXY_IP'].strip()
if (proxy_ip == ''):
    strCmd = "kubectl get nodes --namespace " + namespace + " -l \"proxy = true\" -o jsonpath=\"{.items[0].status.addresses[0].address}\""
    proxy_ip = os.popen(strCmd).read()

generateCertificatesPackage(certs_path)
generateConnectionProfiles(certs_path)
