
#!/usr/bin/env python
# coding=utf-8
import json
import requests
import sys
import os
requests.packages.urllib3.disable_warnings()

def createApiKeySecret(url,user,password):
    payload = { 'roles': ['writer', 'manager'], 'description': 'newkey' }
    api_credentials = sendPostRequest(url, payload,user,password)
    return api_credentials['api_key'],api_credentials['api_secret']


def getComponentByDisplayName(config, org_name ,display_name ,basic_auth_user, basic_auth_password):
    work_dir = config['Initiate']['work_dir']
    all_components = requests.get(config.get('Initiate', 'Console_Url') + config.get('Components', 'All_Components'),
            auth=(config.get('Initiate', 'Api_Key'), config.get('Initiate', 'Api_Secret')), verify=False).json()
    for component in all_components:
        if component['display_name'] == display_name:
            if not os.path.exists(work_dir + '/crypto-config/' + org_name):
                os.makedirs(work_dir + '/crypto-config/' + org_name)
            with open(work_dir + '/crypto-config/' + org_name + '/' + display_name + '.json', 'w') as outfile:
                json.dump(component, outfile)
            break


def getAllComponents(config,basic_auth_user, basic_auth_password):
    url = config.get('Initiate', 'Console_Url') + config.get('Components', 'All_Components')
    response = requests.get(url, auth=(basic_auth_user, basic_auth_password), verify=False)
    if response.status_code == 200:
            return response.json()
    else:
        print 'get all components failed due to '
        print  response.json()
        sys.exit(1)

def checkComponentStatus(config,displayname):
    print 'checkComponentStatus for ' + displayname


def sendPostRequest(url, payload, basic_auth_user, basic_auth_password):
    headers = {'Content-Type': 'application/json', 'charset': 'utf-8'}
    post_response =  requests.post(url,
            headers=headers, data=json.dumps(payload), auth=(basic_auth_user, basic_auth_password), verify=False)
    if post_response.status_code == 200:
        return post_response.json()
    else:
        print 'post request ' + url +'failed due to '
        print  post_response.json()
        sys.exit(1)

def sendDeleteRequest(url, basic_auth_user, basic_auth_password):
    response = requests.delete(url, auth=(basic_auth_user, basic_auth_password), verify=False)
    if response.status_code == 200:
        return response.json()
    else:
        print 'delete request ' + url +'failed due to '
        print  response.json()
        sys.exit(1)

def writeToFile(filePath,content):
    file_content = open(filePath, 'w')
    file_content.write(content)
    file_content.close()

def loadJsonContent(jsonFile):
    with open(jsonFile, 'r') as f:
        temtplateDict = json.load(f)
    f.close()
    return temtplateDict


def constructConfigObject(ca_object_file, template_file, node_admin, ca_admin, node_enroll_id, node_tls_enroll_id):
    config_object = loadJsonContent(template_file)
    ca_object = loadJsonContent(ca_object_file)
    ca_host = ca_object['api_url'].split(':')[1]
    ca_port = ca_object['api_url'].split(':')[2]

    config_object['enrollment']['component']['cahost'] = ca_host[2:len(ca_host)]
    config_object['enrollment']['component']['caport'] = ca_port
    config_object['enrollment']['component']['caname'] = 'ca'
    config_object['enrollment']['component']['catls']['cacert'] = ca_admin
    config_object['enrollment']['component']['enrollid'] = node_enroll_id
    config_object['enrollment']['component']['enrollsecret'] = 'pass4chain'
    config_object['enrollment']['component']['admincerts'] = [node_admin]

    config_object['enrollment']['tls']['cahost'] = ca_host[2:len(ca_host)]
    config_object['enrollment']['tls']['caport'] = ca_port
    config_object['enrollment']['tls']['caname'] = 'tlsca'
    config_object['enrollment']['tls']['catls']['cacert'] = ca_admin
    config_object['enrollment']['tls']['enrollid'] = node_tls_enroll_id
    config_object['enrollment']['tls']['enrollsecret'] = 'pass4chain'
    # node_object['enrollment']['tls']['csr']['hosts'] = []
    return config_object
