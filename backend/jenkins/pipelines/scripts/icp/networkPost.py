#!/usr/bin/env python
# coding=utf-8

#  step1: load networkspec file according to contentRepo, uid, requestid
#           $contentRepo/uid/requestid/networkspec.yml
#            /opt/hfrd/contentrepo/uid/requestid
#           mount host content repo to the docker container and use environment variables to pass uid and request id

#  step2: create icp network based on networkspec.yml

import os
import yaml
import subprocess


def getDictKey(dictVar, keys):
    interDict = dictVar
    for key in keys:
        interDict = interDict.get(key)
        if interDict == None:
            return ''
    return interDict

uid = os.getenv('USER_ID')
requestid = os.getenv('REQ_ID')
networkspec_file = '/opt/hfrd/contentRepo/' + uid + '/' + requestid + '/networkspec.yml'
networkspec = {}

with open(networkspec_file, 'r') as stream:
    networkspec = yaml.load(stream)

# extract and export all required environment variables from networkspec
rConfig = open("/opt/src/scripts/icp/config.tpl").read()
# icp related config
rConfig = rConfig.replace('cluster_ip', getDictKey(networkspec, ['icp', 'cluster_ip']))

rConfig = rConfig.replace('proxy_ip', getDictKey(networkspec, ['icp', 'proxy_ip']))

rConfig = rConfig.replace('namespace', getDictKey(networkspec, ['icp', 'namespace']))

rConfig = rConfig.replace('storage_class', getDictKey(networkspec, ['icp', 'storage_class']))

rConfig = rConfig.replace('user', getDictKey(networkspec, ['icp', 'user']))

rConfig = rConfig.replace('password', getDictKey(networkspec, ['icp', 'password']))

# helm related config
rConfig = rConfig.replace('helm_branch', getDictKey(networkspec, ['helm_branch']))

# network topology related config
rConfig = rConfig.replace('name', getDictKey(networkspec, ['network', 'name']))

rConfig = rConfig.replace('fabric_version', getDictKey(networkspec, ['network', 'fabric_version']))

rConfig = rConfig.replace('arch', getDictKey(networkspec, ['network', 'arch']))

rConfig = rConfig.replace('num_orderers', str(getDictKey(networkspec, ['network', 'orderer', 'num_orderers'])))

rConfig = rConfig.replace('num_orgs', str(getDictKey(networkspec, ['network', 'peer', 'num_orgs'])))

rConfig = rConfig.replace('peers_per_org', str(getDictKey(networkspec, ['network', 'peer', 'peers_per_org'])))

# image repo related config
rConfig = rConfig.replace('ca_image_repo', getDictKey(networkspec, ['images_repo', 'ca', 'image_repo']))

rConfig = rConfig.replace('ca_tag', getDictKey(networkspec, ['images_repo', 'ca', 'tag']))

rConfig = rConfig.replace('ca_init_image_repo', getDictKey(networkspec, ['images_repo', 'ca', 'init_image_repo']))

rConfig = rConfig.replace('ca_init_tag', getDictKey(networkspec, ['images_repo', 'ca', 'init_image_tag']))

rConfig = rConfig.replace('orderer_image_repo', getDictKey(networkspec, ['images_repo', 'orderer', 'image_repo']))

rConfig = rConfig.replace('orderer_tag', getDictKey(networkspec, ['images_repo', 'orderer', 'tag']))

rConfig = rConfig.replace('orderer_init_image_repo', getDictKey(networkspec, ['images_repo', 'orderer', 'init_image_repo']))

rConfig = rConfig.replace('orderer_init_tag', getDictKey(networkspec, ['images_repo', 'orderer', 'init_image_tag']))

rConfig = rConfig.replace('peer_image_repo', getDictKey(networkspec, ['images_repo', 'peer', 'image_repo']))

rConfig = rConfig.replace('peer_tag', getDictKey(networkspec, ['images_repo', 'peer', 'tag']))

rConfig = rConfig.replace('peer_dind_image_repo', getDictKey(networkspec, ['images_repo', 'peer', 'dind_image_repo']))

rConfig = rConfig.replace('peer_dind_tag', getDictKey(networkspec, ['images_repo', 'peer', 'dind_image_tag']))

rConfig = rConfig.replace('peer_init_image_repo', getDictKey(networkspec, ['images_repo', 'peer', 'init_image_repo']))

rConfig = rConfig.replace('peer_init_tag', getDictKey(networkspec, ['images_repo', 'peer', 'init_image_tag']))

# ca,orderer,peer pod resource related configurations
rConfig = rConfig.replace('ca_cpu', str(getDictKey(networkspec, ['network', 'ca', 'cpu'])))

rConfig = rConfig.replace('ca_c_limit', str(getDictKey(networkspec, ['network', 'ca', 'cpu_limit'])))

rConfig = rConfig.replace('ca_memory', str(getDictKey(networkspec, ['network', 'ca', 'memory'])))

rConfig = rConfig.replace('ca_m_limit', str(getDictKey(networkspec, ['network', 'ca', 'memory_limit'])))

rConfig = rConfig.replace('orderer_cpu', str(getDictKey(networkspec, ['network', 'orderer', 'cpu'])))

rConfig = rConfig.replace('orderer_c_limit', str(getDictKey(networkspec, ['network', 'orderer', 'cpu_limit'])))

rConfig = rConfig.replace('orderer_memory', str(getDictKey(networkspec, ['network', 'orderer', 'memory'])))

rConfig = rConfig.replace('orderer_m_limit', str(getDictKey(networkspec, ['network', 'orderer', 'memory_limit'])))


rConfig = rConfig.replace('peer_cpu', str(getDictKey(networkspec, ['network', 'peer', 'cpu'])))

rConfig = rConfig.replace('peer_c_limit', str(getDictKey(networkspec, ['network', 'peer', 'cpu_limit'])))

rConfig = rConfig.replace('peer_memory', str(getDictKey(networkspec, ['network', 'peer', 'memory'])))

rConfig = rConfig.replace('peer_m_limit', str(getDictKey(networkspec, ['network', 'peer', 'memory_limit'])))

# Orderer Settings
rConfig = rConfig.replace('max_batch_timeout', str(getDictKey(networkspec, ['orderer_settings', 'max_batch_timeout'])))
rConfig = rConfig.replace('max_message_count', str(getDictKey(networkspec, ['orderer_settings', 'batch_size', 'max_message_count'])))
rConfig = rConfig.replace('absolute_max_bytes', str(getDictKey(networkspec, ['orderer_settings', 'batch_size', 'absolute_max_bytes'])))
rConfig = rConfig.replace('preferred_max_bytes', str(getDictKey(networkspec, ['orderer_settings', 'batch_size', 'preferred_max_bytes'])))

wConfig = open("/opt/src/scripts/icp/config.cf", 'w')
wConfig.write(rConfig)
wConfig.close()

# Execute bash scripts and monitor the
subprocess.call(['/opt/src/scripts/icp/networkPost.sh'])

