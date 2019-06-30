#!/bin/bash

ARCH=${ARCH:-"amd64"}
STORAGE_CLASS=${STORAGE_CLASS:-"nfs-recycle"}
GLOBAL_NAMESPACE=${NAMESPACE:-"blockchain-dev"}
PEER_IMAGE_REPO=${PEER_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-fabric-peer"}
PEER_TAG=${PEER_TAG:-"1.2.1"}
PEER_DIND_IMAGE_REPO=${PEER_DIND_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-dind"}
PEER_DIND_TAG=${PEER_DIND_TAG:-"1.2.1.1"}
PEER_INIT_IMAGE_REPO=${PEER_INIT_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-init"}
PEER_INIT_TAG=${PEER_INIT_TAG:-"1.2.1"}
MULTIARCH=${MULTIARCH:-"false"}
PEER_CPU=${PEER_CPU:-"100m"}
PEER_CPU_LIMIT="${PEER_CPU_LIMIT:-$PEER_CPU}"
PEER_MEMORY=${PEER_MEMORY:-"256M"}
PEER_MEMORY_LIMIT="${PEER_MEMORY_LIMIT:-$PEER_MEMORY}"
COUCHDB_CPU=${COUCHDB_CPU:-"100m"}
COUCHDB_MEMORY=${COUCHDB_MEMORY:-"256M"}
PEER_PROXY_CPU=${PEER_PROXY_CPU:-"200m"}
PEER_PROXY_CPU_LIMIT="${PEER_PROXY_CPU_LIMIT:-$PEER_PROXY_CPU}"
PEER_PROXY_MEMORY=${PEER_PROXY_MEMORY:-"200M"}
PEER_PROXY_MEMORY_LIMIT="${PEER_PROXY_MEMORY_LIMIT:-$PEER_PROXY_MEMORY}"
HELM_REPO=${HELM_REPO:-"repo"}
PROD_VERSION=${PROD_VERSION:-"1.0.2"}
GRPC_IMAGE_REPO=${GRPC_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-grpcweb"}
GRPC_IMAGE_TAG=${GRPC_IMAGE_TAG:-"1.4.0"}

set -x
helm install ./HelmCharts/ibm-blockchain-platform-prod --version ${PROD_VERSION} -n ${NAME} \
--set peer.enabled=true \
--set peer.app.arch=${ARCH} \
--set peer.app.secret=${MSP_SECRET} \
--set peer.env.orgMSP=${ORGMSP} \
--set peer.service.type=NodePort \
--set peer.dataPVC.storageClassName=${STORAGE_CLASS} \
--set peer.statedbPVC.storageClassName=${STORAGE_CLASS} \
--set peer.app.image=${PEER_IMAGE_REPO} \
--set peer.app.tag=${PEER_TAG} \
--set peer.app.grpcwebimage=${GRPC_IMAGE_REPO} \
--set peer.app.grpcwebtag=${GRPC_IMAGE_TAG} \
--set peer.app.dindimage=${PEER_DIND_IMAGE_REPO} \
--set peer.app.dindtag=${PEER_DIND_TAG} \
--set peer.app.initimage=${PEER_INIT_IMAGE_REPO} \
--set peer.app.inittag=${PEER_INIT_TAG} \
--set global.license=accept \
--set global.multiarch=${MULTIARCH} \
--set peer.app.imagePullSecret=${PULL_SECRET_NAME} \
--set peer.resources.dind.requests.cpu=${PEER_CPU} \
--set peer.resources.dind.requests.memory=${PEER_MEMORY} \
--set peer.peerResources.requests.cpu=${PEER_CPU} \
--set peer.peerResources.requests.memory=${PEER_MEMORY} \
--set peer.resources.proxy.requests.cpu=${PEER_PROXY_CPU} \
--set peer.resources.proxy.requests.memory=${PEER_PROXY_MEMORY} \
--set peer.resources.dind.limits.cpu=${PEER_CPU_LIMIT} \
--set peer.resources.dind.limits.memory=${PEER_MEMORY_LIMIT} \
--set peer.peerResources.limits.cpu=${PEER_CPU_LIMIT} \
--set peer.peerResources.limits.memory=${PEER_MEMORY_LIMIT} \
--set peer.resources.proxy.limits.cpu=${PEER_PROXY_CPU_LIMIT} \
--set peer.resources.proxy.limits.memory=${PEER_PROXY_MEMORY_LIMIT} \
--tls

RC=$?
set +x

if [ $RC != 0 ]; then
    echo "Peer $NAME Deployment Failed"
    exit 1
fi
