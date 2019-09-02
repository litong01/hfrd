#!/usr/bin/env python
# coding=utf-8

import utils

def create_orderer(ordererorg_name, orderer_name, config):
    work_dir = config.get('Initiate', 'Work_Dir')
    create_orderer_url = config.get('Initiate', 'Console_Url') + config.get('Components', 'ORDERER')

    peer_admin = open(work_dir + '/crypto-config/' + ordererorg_name + '/peer_signed_cert', 'r')
    ca_tls_admin = open(work_dir + '/crypto-config/' + ordererorg_name + '/ca_tls_cert', 'r')
    peer_payload = utils.constructPeerObject(work_dir + '/crypto-config/' + ordererorg_name + '/' + ordererorg_name + 'ca.json', peer_admin.read(), ca_tls_admin.read(), 'peer', 'peertls')
    peer_payload['msp_id'] = ordererorg_name + 'msp'
    peer_payload['type'] = 'fabric-peer'
    peer_payload['display_name'] = ordererorg_name

    orderer_response = utils.sendPostRequest(create_orderer_url, peer_payload,
            config.get('Initiate', 'Api_Key'),
            config.get('Initiate', 'Api_Secret'))
    # TODO: check response status

