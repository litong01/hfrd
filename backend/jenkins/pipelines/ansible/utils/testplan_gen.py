
#!/usr/bin/python

import yaml
import os
import ast
import sys
from collections import OrderedDict

curr_dir = os.getcwd()
work_dir = sys.argv[1]
network_type = sys.argv[2]

testplan_dict = {}
testplan_dict["name"] = "System performance test"
testplan_dict["description"] = "This test is to create as much chaincode computation load as possible"
testplan_dict["runid"] = "RUNID_HERE"
if network_type == "ibp":
    testplan_dict["networkid"] = sys.argv[3]
testplan_dict["collectFabricMetrics"] = False
testplan_dict["storageclass"] = "default"
testplan_dict["saveLog"] = False
testplan_dict["continueAfterFail"] = True
testplan_dict["tests"] = []
testplan_dict["peernodeAlias"] =[]

if os.path.exists(work_dir) != True:
    print 'certs keyfiles directory do not exist'
    exit(1)
# Load template file
with open(curr_dir + "/templates/testplan_template.yml", 'r') as stream:
    template = yaml.load(stream)
    channel_create = template["CHANNEL_CREATE"]
    # channel_join = template["CHANNEL_JOIN"]
    chaincode_install = template["CHAINCODE_INSTALL"]
    chaincode_instantiate = template["CHAINCODE_INSTANTIATE"]
    chaincode_invoke = template["CHAINCODE_INVOKE"]
    execute_command = template["EXECUTE_COMMAND"]

connectionProfile = {}
org_list = []
org_list_lowercase = []
orderer_list = []
peer_list = []
org_peers_dict = {}
org_anchor_dict ={}
allAnchor_list =[]

# Load connection profile
for orgName in os.listdir(work_dir + '/keyfiles'):
    if os.path.isfile(work_dir + '/keyfiles/' + orgName + '/connection.yml'):
        with open(work_dir + '/keyfiles/' + orgName + '/connection.yml', 'r') as stream:
                connectionProfile = yaml.load(stream)
                if connectionProfile["orderers"] is None:
                    continue
                orderer_list = orderer_list + connectionProfile["orderers"].keys()
                if (connectionProfile["organizations"][orgName.lower()]["peers"] != None):
                    org_list.append(orgName)
                    org_list_lowercase.append(orgName.lower())
                    org_peers_dict[orgName] = connectionProfile["organizations"][orgName.lower(
                    )]["peers"]
                    peer_list = peer_list + \
                        connectionProfile["organizations"][orgName.lower(
                        )]["peers"]

                    org_anchor_dict[orgName] = sorted(
                        connectionProfile["organizations"][orgName.lower(
                        )]["peers"])[0]     
# When there is only peer or orderer, we skip tests.
if len(orderer_list) == 0 or len(peer_list) == 0:
        outputfile =open(work_dir + '/testplan_example.yml','w')
        outputfile.write("")
        outputfile.close()
        exit(0)

orderer_list = list(OrderedDict.fromkeys(orderer_list))
peer_list = list(OrderedDict.fromkeys(peer_list))

for orgName in org_list :
        tempOrgAnchorObj={}
        tempOrgAnchorObj[orgName+"Anchor"] = org_anchor_dict[orgName]
        testplan_dict["peernodeAlias"].append(tempOrgAnchorObj)
        tempOrgPeersObj={}
        tempOrgPeersObj[orgName+"Peers"] = ','.join(org_peers_dict[orgName])
        testplan_dict["peernodeAlias"].append(tempOrgPeersObj)
        allAnchor_list.append(org_anchor_dict[orgName])
testplan_dict["peernodeAlias"].append({"allAnchors":','.join(allAnchor_list)})
testplan_dict["peernodeAlias"].append({"allPeers":','.join(peer_list)})

print 'org list: '
print org_list_lowercase
print 'orderer_list: '
print  orderer_list
print 'peer_list: '
print  peer_list
print 'allAnchor_list'
print allAnchor_list

# CREATE_CHANNEL
channel_create["parameters"]["connectionProfile"] = org_list[0]
if network_type == 'cello':
    channel_create["parameters"]["channelConsortium"] = 'FabricConsortium'
else:
    channel_create["parameters"]["channelConsortium"] = 'SampleConsortium'
channel_create["parameters"]["channelOrgs"] = ','.join(org_list_lowercase)
channel_create["parameters"]["ordererName"] = orderer_list[0]
testplan_dict["tests"].append(channel_create)

# JOIN_CHANNEL and INSTALL_CHAINCODE
join_list = []
install_list = []
for org in org_list:
    channel_join = template["CHANNEL_JOIN"]
    channel_join["parameters"]["connectionProfile"] = org
    channel_join["parameters"]["peers"] = ','.join(org_peers_dict[org])
    channel_join["parameters"]["ordererName"] = orderer_list[0]
    join_list.append(str(channel_join))

    # CHAINCODE_INSTALL
    chaincode_install["parameters"]["connectionProfile"] = org
    chaincode_install["parameters"]["peers"] = ','.join(org_peers_dict[org])
    install_list.append(str(chaincode_install))
for join_org in join_list:
    join_item = ast.literal_eval(join_org)
    testplan_dict["tests"].append(join_item)

for install_org in install_list:
    install_item = ast.literal_eval(install_org)
    testplan_dict["tests"].append(install_item)

# CHAINCODE_INSTANTIATE
chaincode_instantiate["parameters"]["connectionProfile"] = org_list[0]
chaincode_instantiate["parameters"]["peers"] = peer_list[0]

# CHAINCODE_INVOKE
# Invoke with fixed transaction count : 100
chaincode_invoke["iterationCount"] = '100'
chaincode_invoke["parameters"]["connectionProfile"] = org_list[0]
chaincode_invoke["parameters"]["peers"] = ','.join(peer_list)
chaincoode_invoke_count = str(chaincode_invoke)

# Invoke with fixed running duration : 0 hour 10 minutes 0 second.
# And enable running tests parallel by setting waitUntilFinish to true
chaincode_invoke["iterationCount"] = '0h10m0s'
chaincode_invoke["waitUntilFinish"] = False
chaincoode_invoke_time = str(chaincode_invoke)

# Invoke with fixed running duration : 0 hour 10 minutes 0 second
chaincode_invoke["iterationCount"] = '0h10m0s'
chaincode_invoke["parameters"]["peers"] = peer_list[0]
chaincoode_invoke_parallel = str(chaincode_invoke)

testplan_dict["tests"].append(chaincode_instantiate)
testplan_dict["tests"].append(ast.literal_eval(chaincoode_invoke_count))
testplan_dict["tests"].append(ast.literal_eval(chaincoode_invoke_time))
testplan_dict["tests"].append(ast.literal_eval(chaincoode_invoke_parallel))

# Execute command with default images
testplan_dict["tests"].append(ast.literal_eval(str(execute_command)))
# Execute command with customized image
execute_command["name"] = "execute-command-with-customized-image"
execute_command["container"] = "user/ownimage"
testplan_dict["tests"].append(ast.literal_eval(str(execute_command)))

connYamlStr= yaml.dump(testplan_dict,default_flow_style=False)
tempstr= connYamlStr
for orgName in org_list :
    tempstr = tempstr.replace(orgName+"Anchor:",orgName+"Anchor: &"+orgName+"Anchor")
    tempstr = tempstr.replace(orgName+"Peers:",orgName+"Peers: &"+orgName+"Peers")
tempstr = tempstr.replace("allAnchors:","allAnchors: &allAnchors")
tempstr = tempstr.replace("allPeers:","allPeers: &allPeers")

tempstr = tempstr.replace("runid:","runid: &runid")
if network_type == "ibp":
    tempstr = tempstr.replace("networkid:","networkid: &networkid")
# Dump testplan file
outputfile =open(work_dir + '/testplan_example.yml','w')
outputfile.write(tempstr)
outputfile.close()
