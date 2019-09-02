#!/usr/bin/env python
# coding=utf-8

import configparser, time, yaml, sys
import utils, node

action = sys.argv[1]
config = configparser.ConfigParser()
config.read('templates/apis_template.ini')

networkspec_file = config.get('Initiate', 'networkspec_file')
with open(networkspec_file, 'r') as stream:
    networkspec = yaml.load(stream , Loader=yaml.FullLoader)

ibp4icp = networkspec['ibp4icp']
resources = networkspec['resources']
network = networkspec['network']
raftsettings = networkspec['raftsettings']
orderersettings = networkspec['orderersettings']
peersettings = networkspec['peersettings']

url = ibp4icp['url'] + config.get('Initiate', 'Api_Key_URL')
api_key, api_secret = utils.createApiKeySecret(url, ibp4icp['user'], ibp4icp['password'])
time.sleep(2)

config['Initiate']['Console_Url'] = ibp4icp['url']
config['Initiate']['Manager_User'] = ibp4icp['user']
config['Initiate']['Manager_Password'] = ibp4icp['password']
config['Initiate']['Api_Key'] = api_key
config['Initiate']['Api_Secret'] = api_secret
config['Initiate']['Work_Dir'] = networkspec['work_dir']
# Get orgs and nodes
peers = network['peers']
orderers = network['orderers']

utils.getComponentByDisplayName(config, 'ordererorg', 'ordererorg-orderer-1', api_key, api_secret)
