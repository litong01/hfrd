#!/bin/bash -xe

org_name=$1
work_dir=$2
binary_url=$3
ca_name=${org_name}'ca'

source $work_dir'/apis.ini' || true


TLS_CERT=$(jq -r .tls_cert $work_dir/crypto-config/${org_name}/${ca_name}.json)
ENROLL_ID=$(jq -r .enroll_id $work_dir/crypto-config/${org_name}/${ca_name}.json)
ENROLL_PASS=$(jq -r .enroll_secret $work_dir/crypto-config/${org_name}/${ca_name}.json)
CA_URL=$(jq -r .api_url $work_dir/crypto-config/${org_name}/${ca_name}.json)
CA_URL=${CA_URL:8}
CA_NAME=$(jq -r .ca_name $work_dir/crypto-config/${org_name}/${ca_name}.json)
TLS_CA_NAME=$(jq -r .tlsca_name $work_dir/crypto-config/${org_name}/${ca_name}.json)

if [ ! -d $work_dir'/bin/' ]; then
    curl -f -s -C - ${binary_url} -o fabric.tar.gz
    tar zxf fabric.tar.gz
fi

if [ ! -f $work_dir'/bin/cloudctl' ]; then

  curl -kLo cloudctl-linux-amd64-v3.2.0-634 https://$icp_url/api/cli/cloudctl-linux-amd64
  mv cloudctl* $work_dir'/bin/cloudctl'
  chmod +x $work_dir'/bin/cloudctl'

  curl -kLo kubectl-linux-amd64-v1.13.5 https://$icp_url/api/cli/kubectl-linux-amd64
  mv kubectl* $work_dir'/bin/kubectl'
  chmod +x $work_dir'/bin/kubectl'
fi

export PATH=$PATH:$work_dir/bin

cloudctl login -a https://$icp_url --skip-ssl-validation -u $icp_user -p $icp_password -n $icp_namespace


BASE_FOLDER=$work_dir'/crypto-config'

PEER_ORG_NAME=${org_name}
export ${PEER_ORG_NAME}_FOLDER="${BASE_FOLDER}/${PEER_ORG_NAME}"
export ${PEER_ORG_NAME}_CA_FOLDER=${BASE_FOLDER}/${PEER_ORG_NAME}/ca
export ${PEER_ORG_NAME}_ECA_FOLDER=${BASE_FOLDER}/${PEER_ORG_NAME}/ca/enrollment/
export ${PEER_ORG_NAME}_TLSCA_FOLDER=${BASE_FOLDER}/${PEER_ORG_NAME}/ca/tls/
export ${PEER_ORG_NAME}_ADMIN_FOLDER=${BASE_FOLDER}/${PEER_ORG_NAME}/admin
export ${PEER_ORG_NAME}_USER_FOLDER=${BASE_FOLDER}/${PEER_ORG_NAME}/user
var=${PEER_ORG_NAME}_FOLDER
mkdir -p ${!var}
var=${PEER_ORG_NAME}_CA_FOLDER
mkdir -p ${!var}
var=${PEER_ORG_NAME}_ECA_FOLDER
mkdir -p ${!var}
var=${PEER_ORG_NAME}_TLSCA_FOLDER
mkdir -p ${!var}
var=${PEER_ORG_NAME}_ADMIN_FOLDER
mkdir -p ${!var}
var=${PEER_ORG_NAME}_USER_FOLDER
mkdir -p ${!var}


IFS=':' read -ra ADDR <<< "$CA_URL"
export PROXY_IP=${ADDR[0]}
export ${PEER_ORG_NAME}_CA_HOST=${ADDR[0]}
export ${PEER_ORG_NAME}_CA_PORT=${ADDR[1]}

var0=${PEER_ORG_NAME}_CA_HOST
var1=${PEER_ORG_NAME}_CA_PORT
NAME=${PEER_ORG_NAME}ca CA_HOST=${!var0} CA_PORT=${!var1} ./wait_for_pod.sh

var0=${PEER_ORG_NAME}_CA_FOLDER
echo $TLS_CERT | base64 -d -w 0 > ${!var0}/tls-ca-cert.pem


var=${PEER_ORG_NAME}_CA_FOLDER
var0=${PEER_ORG_NAME}_CA_PORT
var1=${PEER_ORG_NAME}_CA_HOST
var2=${PEER_ORG_NAME}_ECA_FOLDER
var3=${PEER_ORG_NAME}_ADMIN_FOLDER
var4=${PEER_ORG_NAME}_TLSCA_FOLDER
var5=${PEER_ORG_NAME}_USER_FOLDER

CSRHOSTS="${PROXY_IP},${PEER_ORG_NAME},127.0.0.1,localhost"
FABRIC_CA_CLIENT_HOME=${!var2} fabric-ca-client enroll -u https://admin:pass4chain@${!var1}:${!var0} --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var2} fabric-ca-client register --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --id.name peeradmin --id.secret pass4chain --id.type user
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var2} fabric-ca-client register --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --id.name wally --id.secret pass4chain --id.type user
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var3} fabric-ca-client enroll -u https://peeradmin:pass4chain@${!var1}:${!var0} --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var5} fabric-ca-client enroll -u https://wally:pass4chain@${!var1}:${!var0} --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var4} fabric-ca-client enroll -u https://admin:pass4chain@${!var1}:${!var0} --caname ${TLS_CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var4} fabric-ca-client register --caname ${TLS_CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --id.name peertls --id.secret pass4chain --id.type peer
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

peer_signed_cert=$(cat $work_dir/crypto-config/${org_name}/admin/msp/signcerts/cert.pem | base64 -w 0)
root_certs=$(cat $work_dir/crypto-config/${org_name}/ca/enrollment/msp/signcerts/cert.pem | base64 -w 0)
tls_root_certs=$(cat $work_dir/crypto-config/${org_name}/ca/tls/msp/signcerts/cert.pem | base64 -w 0)

echo $peer_signed_cert > $work_dir/crypto-config/${org_name}/peer_signed_cert
echo $TLS_CERT > $work_dir/crypto-config/${org_name}/ca_tls_cert
echo $root_certs > $work_dir/crypto-config/${org_name}/ca_admin_cert
echo $tls_root_certs > $work_dir/crypto-config/${org_name}/tls_ca_cert
