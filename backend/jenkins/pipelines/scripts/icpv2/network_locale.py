#!/usr/bin/env python
# coding=utf-8

import configparser, time, yaml, sys
import utils, node
import subprocess

action = sys.argv[1]
config = configparser.ConfigParser()
config.read('templates/apis_template.ini')

networkspec_file = config.get('Initiate', 'networkspec_file')
with open(networkspec_file, 'r') as stream:
    networkspec = yaml.load(stream , Loader=yaml.FullLoader)

icp = networkspec['icp']
ibp4icp = icp['ibp4icp']
resources = networkspec['resources']
network = networkspec['network']
raftsettings = networkspec['raftsettings']
orderersettings = networkspec['orderersettings']
peersettings = networkspec['peersettings']

url = ibp4icp['url'] + config.get('Initiate', 'Api_Key_URL')
api_key, api_secret = utils.createApiKeySecret(url, ibp4icp['user'], ibp4icp['password'])
time.sleep(2)

config['Initiate']['ICP_Url'] = icp['url']
config['Initiate']['ICP_User'] = icp['user']
config['Initiate']['ICP_Password'] = icp['password']
config['Initiate']['ICP_Namespace'] = icp['namespace']
config['Initiate']['ICP_Storageclass'] = icp['storageclass']

config['Initiate']['Console_Url'] = ibp4icp['url']
config['Initiate']['Manager_User'] = ibp4icp['user']
config['Initiate']['Manager_Password'] = ibp4icp['password']
config['Initiate']['Api_Key'] = api_key
config['Initiate']['Api_Secret'] = api_secret
config['Initiate']['Work_Dir'] = networkspec['work_dir']

with open('apis.ini', 'w') as configfile:
    config.write(configfile, space_around_delimiters=False)

# Get orgs and nodes
peers = network['peers']
orderers = network['orderers']
ordererorg_names = []
peerorg_names = []


def create_organization(config, networkspec,org_name, node_type):
    print 'create ca,msp for organization: ' + org_name
    node.create_ca(org_name, config)
    time.sleep(5)
    node.query_ca(org_name, config)
    print 'create msp for organization: ' + org_name
    node.create_msp(org_name, node_type ,config,networkspec)


def create_node(config, networkspec, node_object, node_type):
    split_arr = node_object.split('.')
    org_name = split_arr[1]
    node_name = org_name + '-' + split_arr[0]
    if org_name not in org_list:
        org_list[org_name] = []
        org_list[org_name].append(node_name)
        create_organization(config, networkspec,org_name,node_type)
    if node_type == 'peer':
        node.create_peer(config,networkspec, org_name, node_name)
    elif node_type == 'orderer':
        # if type is orderer, node_name would be the number of orderer nodes in raft conensus
        node.create_orderer(config, networkspec, org_name, int(split_arr[0]))
    print 'successfully created node ' + node_object


if action == 'create':
    org_list = {}
    for peer_object in peers:
        peerorg_names.append(peer_object.split('.')[1])
        create_node(config, networkspec, peer_object, 'peer')
    for orderer_object in orderers:
        ordererorg_names.append(orderer_object.split('.')[1])
        create_node(config, networkspec, orderer_object, 'orderer')
    # Update system channel
    ordererorg_names = list(set(ordererorg_names))
    peerorg_names = list(set(peerorg_names))
    peerorg_names_string = ','.join(peerorg_names)
    for ordererorg_name in ordererorg_names:
        if subprocess.call([networkspec['work_dir'] + '/update_system_channel.sh', ordererorg_name, peerorg_names_string ] ) == 1:
            print 'error found when update system channel '
            sys.exit(1)
elif action == 'delete_all':
    delete_all_url = config['Initiate']['Console_Url'] + config['Components']['delete_all_components']
    utils.sendDeleteRequest(delete_all_url, api_key, api_secret)
    while (len(utils.getAllComponents(config, api_key, api_secret)) > 0):
        print 'waiting for all components deleted'
        time.sleep(3)
    print 'current components:  '
    print utils.getAllComponents(config, api_key, api_secret)
    print 'all components are deleted '
elif action == 'delete':
    print 'delete the components specified in network spec file'
    org_list = []
    nodes = peers + orderers
    for node_object in nodes:
        split_arr = peer_object.split('.')
        org_name = split_arr[1]
        org_list.append(org_name)
    components = utils.getAllComponents(config, api_key, api_secret)
    for component in components:
        if component['display_name'] == org_name + 'ca':
            if not os.path.exists(work_dir + '/crypto-config/' + org_name):
                os.makedirs(work_dir + '/crypto-config/' + org_name)
            with open(work_dir + '/crypto-config/' + org_name + '/' + org_name + 'ca.json', 'w') as outfile:
                json.dump(component, outfile)
            break

