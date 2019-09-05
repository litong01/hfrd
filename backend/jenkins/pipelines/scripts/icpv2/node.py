#!/usr/bin/env python
# coding=utf-8

import json
import requests
import utils
import sys
import os
import subprocess
requests.packages.urllib3.disable_warnings()


def create_ca(org_name, config):
    create_ca_url = config.get('Initiate', 'Console_Url') + config.get('Components', 'CA')
    payload = { 'display_name': org_name + 'ca', 'enroll_id': 'admin', 'enroll_secret': 'pass4chain' }
    utils.sendPostRequest(create_ca_url, payload, config.get('Initiate', 'Api_Key'), config.get('Initiate', 'Api_Secret'))
    print 'successfully created ca for organization ' + org_name


def query_ca(org_name, config):
    work_dir = config.get('Initiate', 'Work_Dir')
    all_components = requests.get(config.get('Initiate', 'Console_Url') + config.get('Components', 'All_Components'),
            auth=(config.get('Initiate', 'Api_Key'), config.get('Initiate', 'Api_Secret')), verify=False).json()
    for component in all_components:
        if component['display_name'] == org_name + 'ca':
            if not os.path.exists(work_dir + '/crypto-config/' + org_name):
                os.makedirs(work_dir + '/crypto-config/' + org_name)
            with open(work_dir + '/crypto-config/' + org_name + '/' + org_name + 'ca.json', 'w') as outfile:
                json.dump(component, outfile)
            break

def create_msp(org_name, node_type ,config,networkspec):
    work_dir = config.get('Initiate', 'Work_Dir')
    binary_url = networkspec['repo']['bin']
    if node_type == 'orderer':
        if subprocess.call([work_dir + '/enroll_orderer.sh', org_name, work_dir, binary_url]) == 1:
            print 'error found when create msp for org ' + org_name
            sys.exit(1)
    else:
        if subprocess.call([work_dir + '/enroll_peer.sh', org_name, work_dir, binary_url]) == 1:
            print 'error found when create msp for org ' + org_name
            sys.exit(1)
    # Create a msp definition
    admin = open(work_dir + '/crypto-config/' + org_name + '/peer_signed_cert', 'r')
    root_certs = open(work_dir + '/crypto-config/' + org_name + '/ca_admin_cert', 'r')
    tls_root_certs = open(work_dir + '/crypto-config/' + org_name + '/tls_ca_cert', 'r')
    msp_data = {'msp_id':org_name, 'display_name': org_name, 'root_certs': [root_certs.read()], 'intermediate_certs': [],
            'admins':[admin.read()], 'tls_root_certs':[tls_root_certs.read()], 'revocation_list':[], 'organizational_unit_identifiers':['client','peer'],
            'fabric_node_ous':['client','peer']}
    print msp_data
    with open(work_dir + '/crypto-config/' + org_name + '/' + org_name + 'msp.json', 'w') as outfile:
                json.dump(msp_data, outfile)
    utils.sendPostRequest(config.get('Initiate', 'Console_Url') + config.get('Components', 'MSP'),
                msp_data, config.get('Initiate', 'Api_Key'), config.get('Initiate', 'Api_Secret'))
    print 'successfully created msp for organization ' + org_name


def create_peer(config,networkspec, org_name, peer_name):
    work_dir = config.get('Initiate', 'Work_Dir')
    create_peer_url = config.get('Initiate', 'Console_Url') + config.get('Components', 'PEER')
    api_key = config.get('Initiate', 'Api_Key')
    api_secret = config.get('Initiate', 'Api_Secret')
    peer_admin = open(work_dir + '/crypto-config/' + org_name + '/peer_signed_cert', 'r')
    ca_tls_admin = open(work_dir + '/crypto-config/' + org_name + '/ca_tls_cert', 'r')
    peer_config = utils.constructConfigObject(work_dir + '/crypto-config/' + org_name + '/' + org_name + 'ca.json',
                                work_dir + '/templates/config_template.json' ,
                                peer_admin.read(), ca_tls_admin.read(), 'peeradmin', 'admin')
    peer_payload = utils.loadJsonContent(work_dir + '/templates/peer_config_template.json')
    peer_payload['msp_id'] = org_name
    peer_payload['type'] = 'fabric-peer'
    peer_payload['display_name'] = peer_name
    peer_payload['config'] = peer_config
    peer_payload['resources']['peer']['requests']['cpu'] = networkspec['resources']['peer']['cpu_req']
    peer_payload['resources']['peer']['requests']['memory'] = networkspec['resources']['peer']['mem_req']
    utils.sendPostRequest(create_peer_url, peer_payload,api_key,api_secret)
    print 'successfully created peer ' + peer_name + ' for organization ' + org_name
    # Get peer component
    utils.getComponentByDisplayName(config, org_name, peer_name, api_key, api_secret)


def create_orderer(config, networkspec, service_name, num_of_orderers):
    work_dir = config.get('Initiate', 'Work_Dir')
    create_orderer_url = config.get('Initiate', 'Console_Url') + config.get('Components', 'ORDERER')
    api_key = config.get('Initiate', 'Api_Key')
    api_secret = config.get('Initiate', 'Api_Secret')

    orderer_payload = utils.loadJsonContent(work_dir + '/templates/orderer_config_template.json')
    orderer_payload['msp_id'] = service_name
    orderer_payload['cluster_name'] = service_name
    orderer_payload['display_name'] = service_name + '-orderer'
    orderer_payload['resources']['orderer']['requests']['cpu'] = networkspec['resources']['orderer']['cpu_req']
    orderer_payload['resources']['orderer']['requests']['memory'] = networkspec['resources']['orderer']['mem_req']

    while(num_of_orderers > 0):
        orderer_admin = open(work_dir + '/crypto-config/' + service_name + '/peer_signed_cert', 'r')
        ca_tls_admin = open(work_dir + '/crypto-config/' + service_name + '/ca_tls_cert', 'r')
        orderer_config = utils.constructConfigObject(work_dir + '/crypto-config/' + service_name + '/' + service_name + 'ca.json',
                                work_dir + '/templates/config_template.json' ,
                                orderer_admin.read(), ca_tls_admin.read(), 'admin', 'admin')
        orderer_payload['config'].append(orderer_config)
        num_of_orderers -=  1
    utils.sendPostRequest(create_orderer_url, orderer_payload,
            config.get('Initiate', 'Api_Key'),
            config.get('Initiate', 'Api_Secret'))

    # Get orderer component
    utils.getComponentByDisplayName(config, service_name, service_name + '-orderer-1', api_key, api_secret)
