import json
import os

HOME = os.environ['HOME']

# read in connection profiles for both orgs
with open(os.path.join(HOME, 'creds/org1ConnectionProfile.json'), 'r') as f:
    org1CP = json.load(f)
with open(os.path.join(HOME, 'creds/org2ConnectionProfile.json'), 'r') as f:
    org2CP = json.load(f)

# read in PTE template
with open(os.path.join(HOME, 'scripts/SCFile-template.json'), 'r') as f:
    SCtemp = json.load(f)

# read in admin cert and priv key for both orgs
with open(os.path.join(HOME, 'creds/org1admin/msp/signcerts/cert.pem'), 'r') as f:
    org1cert = ''.join(f.readlines())
with open(os.path.join(HOME, 'creds/org1admin/msp/keystore/priv.pem'), 'r') as f:
    org1priv = ''.join(f.readlines())
with open(os.path.join(HOME, 'creds/org2admin/msp/signcerts/cert.pem'), 'r') as f:
    org2cert = ''.join(f.readlines())
with open(os.path.join(HOME, 'creds/org2admin/msp/keystore/priv.pem'), 'r') as f:
    org2priv = ''.join(f.readlines())

# move values into the template
SCtemp['test-network']['orderer']['orderer0']['url'] = org1CP['orderers']['orderer']['url']
SCtemp['test-network']['tls_cert'] = org1CP['orderers']['orderer']['tlsCACerts']['pem']

SCtemp['test-network']['org1']['secret'] = org1CP['certificateAuthorities']['org1-ca']['registrar'][0]['enrollSecret']
SCtemp['test-network']['org1']['ca']['url'] = org1CP['certificateAuthorities']['org1-ca']['url']
SCtemp['test-network']['org1']['org1-peer1']['requests'] = org1CP['peers']['org1-peer1']['url']
SCtemp['test-network']['org1']['org1-peer1']['events'] = org1CP['peers']['org1-peer1']['eventUrl']

SCtemp['test-network']['org2']['secret'] = org2CP['certificateAuthorities']['org2-ca']['registrar'][0]['enrollSecret']
SCtemp['test-network']['org2']['ca']['url'] = org2CP['certificateAuthorities']['org2-ca']['url']
SCtemp['test-network']['org2']['org2-peer1']['requests'] = org2CP['peers']['org2-peer1']['url']
SCtemp['test-network']['org2']['org2-peer1']['events'] = org2CP['peers']['org2-peer1']['eventUrl']

SCtemp['test-network']['org1']['admin_cert'] = org1cert
SCtemp['test-network']['org1']['priv'] = org1priv
SCtemp['test-network']['org2']['admin_cert'] = org2cert
SCtemp['test-network']['org2']['priv'] = org2priv

with open(os.path.join(HOME, 'fabric-sdk-node/test/PTE/SCFiles/config-chan1-TLS.json'), 'w') as f:
    json.dump(SCtemp, f, indent=4)