#!/bin/bash

ARCH=${ARCH:-"amd64"}
STORAGE_CLASS=${STORAGE_CLASS:-"nfs-recycle"}
GLOBAL_NAMESPACE=${NAMESPACE:-"blockchain-dev"}
ORDERER_IMAGE_REPO=${ORDERER_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-fabric-orderer"}
ORDERER_TAG=${ORDERER_TAG:-"1.2.1"}
ORDERER_INIT_IMAGE_REPO=${ORDERER_INIT_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-init"}
ORDERER_INIT_TAG=${ORDERER_INIT_TAG:-"1.2.1"}
MULTIARCH=${MULTIARCH:-"false"}
ORDERER_CPU=${ORDERER_CPU:-"100m"}
ORDERER_CPU_LIMIT="${ORDERER_CPU_LIMIT:-$ORDERER_CPU}"
ORDERER_MEMORY=${ORDERER_MEMORY:-"256M"}
ORDERER_MEMORY_LIMIT="${ORDERER_MEMORY_LIMIT:-$ORDERER_MEMORY}"
HELM_REPO=${HELM_REPO:-"repo"}
PROD_VERSION=${PROD_VERSION:-"1.0.2"}
GRPC_IMAGE_REPO=${GRPC_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-grpcweb"}
GRPC_IMAGE_TAG=${GRPC_IMAGE_TAG:-"1.4.0"}
ORDERER_PROXY_CPU=${ORDERER_PROXY_CPU:-"200m"}
ORDERER_PROXY_CPU_LIMIT="${ORDERER_PROXY_CPU_LIMIT:-$ORDERER_PROXY_CPU}"
ORDERER_PROXY_MEMORY=${ORDERER_PROXY_MEMORY:-"200M"}
ORDERER_PROXY_MEMORY_LIMIT="${ORDERER_PROXY_MEMORY_LIMIT:-$ORDERER_PROXY_MEMORY}"

set -x
helm install ./HelmCharts/ibm-blockchain-platform-prod --version ${PROD_VERSION} -n ${NAME} \
--set orderer.enabled=true \
--set orderer.app.mspsecret=${MSP_SECRET} \
--set orderer.app.arch=${ARCH} \
--set orderer.ord.orgName=${ORG_NAME} \
--set orderer.ord.mspID=${ORDERER_MSP_ID} \
--set orderer.service.type=NodePort \
--set orderer.pvc.storageClassName=${STORAGE_CLASS} \
--set orderer.image.repository=${ORDERER_IMAGE_REPO} \
--set orderer.image.tag=${ORDERER_TAG} \
--set orderer.image.grpcwebimage=${GRPC_IMAGE_REPO} \
--set orderer.image.grpcwebtag=${GRPC_IMAGE_TAG} \
--set orderer.image.initimage=${ORDERER_INIT_IMAGE_REPO} \
--set orderer.image.inittag=${ORDERER_INIT_TAG} \
--set global.license=accept \
--set global.multiarch=${MULTIARCH} \
--set orderer.image.imagePullSecret=${PULL_SECRET_NAME} \
--set orderer.resources.requests.cpu=${ORDERER_CPU} \
--set orderer.resources.requests.memory=${ORDERER_MEMORY} \
--set orderer.resources.limits.cpu=${ORDERER_CPU_LIMIT} \
--set orderer.resources.limits.memory=${ORDERER_MEMORY_LIMIT} \
--set orderer.proxyResources.requests.cpu=${ORDERER_PROXY_CPU} \
--set orderer.proxyResources.requests.memory=${ORDERER_PROXY_MEMORY} \
--set orderer.proxyResources.limits.cpu=${ORDERER_PROXY_CPU_LIMIT} \
--set orderer.proxyResources.limits.memory=${ORDERER_PROXY_MEMORY_LIMIT} \
--tls

RC=$?
set +x

if [ $RC != 0 ]; then
    echo "Orderer $NAME Deployment Failed"
    exit 1
fi
