#!/bin/bash -e

start_time=$(date +%s)

source /opt/src/scripts/icp/config.cf


# Default to default Namespace
export NAMESPACE=${NAMESPACE:-"default"}
# Default to example-nfs storage class else pass from to the script
export STORAGE_CLASS=${STORAGE_CLASS:-"nfs-recycle"}
export PROD_VERSION=${PROD_VERSION:-"1.0.2"}
# Defaults to MultiArch as True
export MULTIARCH=${MULTIARCH:-"true"}

export HELM_BRANCH=${HELM_BRANCH:-release-1.0}

# Default to fyre(random value) as the release name
export NAME=${NAME:-$(echo "hfrd")}
# Default to amd64 architecture
export ARCH=${ARCH:-"s390x"}
# Defaults to Fabric v1.4
export FABRIC_VERSION=${FABRIC_VERSION:-"1.4.0"}
# Default to make 2 orgs
export NUM_ORGS=${NUM_ORGS:-"4"}
# 2 peers per org
export PEERS_PER_ORG=${PEERS_PER_ORG:-"2"}

#Default components to use .1 CPU and 256MB of memory to allow scheduling on test system for high org deployments.
export ORDERER_CPU=${ORDERER_CPU:-"250m"}
export ORDERER_CPU_LIMIT="1"
export ORDERER_MEMORY=${ORDERER_MEMORY:-"250Mi"}
export ORDERER_MEMORY_LIMIT="1Gi"
export CA_CPU=${CA_CPU:-"200m"}
export CA_CPU_LIMIT="500m"
export CA_MEMORY=${CA_MEMORY:-"200M"}
export CA_MEMORY_LIMIT="500M"
export PEER_CPU=${PEER_CPU:-"250m"}
export PEER_CPU_LIMIT="1"
export PEER_MEMORY=${PEER_MEMORY:-"300M"}
export PEER_MEMORY_LIMIT="1G"
#export COUCHDB_CPU=${COUCHDB_CPU:-"100m"}
#export COUCHDB_MEMORY=${COUCHDB_MEMORY:-"256M"}
#export COUCHDB_CPU=${COUCHDB_CPU:-"100m"}
#export COUCHDB_MEMORY=${COUCHDB_MEMORY:-"256M"}

# ibmcom/ibp-fabric-peer:1.4.0
# Default values for the Image Name and TAG
# CA Image and Tag
export CA_IMAGE_REPO=${CA_IMAGE_REPO:-"ibmcom/ibp-fabric-ca"}
export CA_TAG=${CA_TAG:-"1.4.0"}
export CA_INIT_IMAGE_REPO=${CA_INIT_IMAGE_REPO:-"ibmcom/ibp-init"}
export CA_INIT_TAG=${CA_INIT_TAG:-"1.4.0"}

#GRPC Images

export GRPC_IMAGE_REPO=${GRPC_IMAGE_REPO:-"ibmcom/ibp-grpcweb"}
export GRPC_IMAGE_TAG=${GRPC_IMAGE_TAG:-"1.4.0"}

# Orderer and Init Image and Tag
export ORDERER_IMAGE_REPO=${ORDERER_IMAGE_REPO:-"ibmcom/ibp-fabric-orderer"}
export ORDERER_TAG=${ORDERER_TAG:-"1.4.0"}
export ORDERER_INIT_IMAGE_REPO=${ORDERER_INIT_IMAGE_REPO:-"ibmcom/ibp-init"}
export ORDERER_INIT_TAG=${ORDERER_INIT_TAG:-"1.4.0"}

# Peer , Dind and Init Image and Tag
export PEER_IMAGE_REPO=${PEER_IMAGE_REPO:-"ibmcom/ibp-fabric-peer"}
export PEER_TAG=${PEER_TAG:-"1.4.0"}
export PEER_DIND_IMAGE_REPO=${PEER_DIND_IMAGE_REPO:-"ibmcom/ibp-dind"}
export PEER_DIND_TAG=${PEER_DIND_TAG:-"1.4.0"}
export PEER_INIT_IMAGE_REPO=${PEER_INIT_IMAGE_REPO:-"ibmcom/ibp-init"}
export PEER_INIT_TAG=${PEER_INIT_TAG:-"1.4.0"}

# Deaults to 'NotIfPresent'
# export IMAGE_PULL_POLICY=${IMAGE_PULL_POLICY:-"Never"} ## Other option to try IfNotPresent

# Install CLIs if not present (cloudctl, helm,kubectl)
curl -kLo cloudctl-linux-amd64-3.1.1-1078 https://$CLUSTER_IP:8443/api/cli/cloudctl-linux-amd64
mv cloudctl* cloudctl
chmod +x cloudctl

curl -kLo kubectl-linux-amd64-v1.11.3 https://$CLUSTER_IP:8443/api/cli/kubectl-linux-amd64
mv kubectl* kubectl
chmod +x kubectl

curl -kLo helm-linux-amd64-v2.9.1.tar.gz https://$CLUSTER_IP:8443/api/cli/helm-linux-amd64.tar.gz
tar -xzvf helm* && mv linux-amd64/helm helm
chmod +x helm
export PATH=$PATH:/opt/fabrictest

cloudctl login -a https://$CLUSTER_IP:8443 --skip-ssl-validation -u $USER -p $PASSWORD -n $NAMESPACE

cd /opt/src/scripts/icp
echo -e "\n\n  ---- creating multi-org network ----\n"
./setup.sh
echo -e "\n\n  ---- Updating system channel with ${NUM_ORGS} org(s) ----\n"
./add_orgs.sh
echo -e "\n\n  ---- Package the certs tar file ----\n"
cp -rf /opt/src/scripts/icp/${NAME} /opt/src/scripts/icp/keyfiles
export CERTS_PATH=/opt/src/scripts/icp/keyfiles
python generateCerts.py /opt/src/scripts/icp/config.cf $CERTS_PATH

tar -zcf icpcerts.tgz keyfiles/ && mv icpcerts.tgz keyfiles/ /opt/hfrd/contentRepo/${USER_ID}/${REQ_ID}/

echo -e "\n ========= HFRD ICP Network Creation  C O M P L E T E D  (Total exec time: $(($(date +%s)-start_time)) secs.) =========\n"

