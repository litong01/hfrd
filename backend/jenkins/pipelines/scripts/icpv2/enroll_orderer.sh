#!/bin/bash -xe

org_name=$1
work_dir=$2
binary_url=$3
ca_name=${org_name}'ca'

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

export PATH=$PATH:$work_dir/bin
ORDERER_ORG_NAME=$org_name
BASE_FOLDER=$work_dir'/crypto-config'
ORDERER_FOLDER=${BASE_FOLDER}/${ORDERER_ORG_NAME}
ORDERER_CA_FOLDER=${ORDERER_FOLDER}/ca
ORDERER_ECA_FOLDER=${ORDERER_FOLDER}/ca/enrollment/
ORDERER_TLSCA_FOLDER=${ORDERER_FOLDER}/ca/tls/
ORDERER_ADMIN_FOLDER=${ORDERER_FOLDER}/admin

mkdir -p ${BASE_FOLDER}
mkdir -p ${ORDERER_FOLDER}
mkdir -p ${ORDERER_CA_FOLDER}
mkdir -p ${ORDERER_ECA_FOLDER}
mkdir -p ${ORDERER_TLSCA_FOLDER}
mkdir -p ${ORDERER_ADMIN_FOLDER}

IFS=':' read -ra ADDR <<< "$CA_URL"
export PROXY_IP=${ADDR[0]}
ORDERERORG_CA_HOST=${ADDR[0]}
ORDERERORG_CA_PORT=${ADDR[1]}

NAME=${ORDERER_ORG_NAME}ca CA_HOST=${ORDERERORG_CA_HOST} CA_PORT=${ORDERERORG_CA_PORT} ./wait_for_pod.sh

echo $TLS_CERT | base64 -d -w 0 > ${ORDERER_CA_FOLDER}/tls-ca-cert.pem

CSRHOSTS="${PROXY_IP},${org_name}-orderer,127.0.0.1,localhost"

FABRIC_CLIENT_RC=0
set -x
FABRIC_CA_CLIENT_HOME=${ORDERER_ECA_FOLDER} fabric-ca-client enroll -u https://admin:pass4chain@${ORDERERORG_CA_HOST}:${ORDERERORG_CA_PORT} --caname ${CA_NAME} --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_ECA_FOLDER} fabric-ca-client register --caname ${CA_NAME} --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --id.name ${org_name}-admin --id.secret pass4chain --id.type user
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_ECA_FOLDER} fabric-ca-client register --caname ${CA_NAME} --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --id.name ordereradmin --id.secret  pass4chain --id.type orderer
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_ADMIN_FOLDER} fabric-ca-client enroll -u https://${org_name}-admin:pass4chain@${ORDERERORG_CA_HOST}:${ORDERERORG_CA_PORT} --caname ${CA_NAME} --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_TLSCA_FOLDER} fabric-ca-client enroll -u https://admin:pass4chain@${ORDERERORG_CA_HOST}:${ORDERERORG_CA_PORT} --caname tlsca --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_TLSCA_FOLDER} fabric-ca-client register --caname ${TLS_CA_NAME} --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --id.name orderertls --id.secret pass4chain --id.type orderer
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

peer_signed_cert=$(cat $work_dir/crypto-config/${org_name}/admin/msp/signcerts/cert.pem | base64 -w 0)
root_certs=$(cat $work_dir/crypto-config/${org_name}/ca/enrollment/msp/signcerts/cert.pem | base64 -w 0)
tls_root_certs=$(cat $work_dir/crypto-config/${org_name}/ca/tls/msp/signcerts/cert.pem | base64 -w 0)

echo $peer_signed_cert > $work_dir/crypto-config/${org_name}/peer_signed_cert
echo $TLS_CERT > $work_dir/crypto-config/${org_name}/ca_tls_cert
echo $root_certs > $work_dir/crypto-config/${org_name}/ca_admin_cert
echo $tls_root_certs > $work_dir/crypto-config/${org_name}/tls_ca_cert

