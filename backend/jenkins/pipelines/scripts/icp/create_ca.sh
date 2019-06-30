#!/bin/bash

ARCH=${ARCH:-"amd64"}
STORAGE_CLASS=${STORAGE_CLASS:-"nfs-recycle"}
GLOBAL_NAMESPACE=${NAMESPACE:-"blockchain-dev"}
CA_IMAGE_REPO=${CA_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-fabric-ca"}
CA_TAG=${CA_TAG:-"1.2.1"}
CA_INIT_IMAGE_REPO=${CA_INIT_IMAGE_REPO:-"mycluster.icp:8500/$GLOBAL_NAMESPACE/ibmcom/ibp-init"}
CA_INIT_TAG=${CA_INIT_TAG:-"1.2.1"}
MULTIARCH=${MULTIARCH:-"false"}
CA_CPU=${CA_CPU:-"100m"}
CA_CPU_LIMIT="${CA_CPU_LIMIT:-$CA_CPU}"
CA_MEMORY=${CA_MEMORY:-"256M"}
CA_MEMORY_LIMIT="${CA_MEMORY_LIMIT:-$CA_MEMORY}"
HELM_REPO=${HELM_REPO:-"repo"}
PROD_VERSION=${PROD_VERSION:-"1.0.2"}

set -x
helm install ./HelmCharts/ibm-blockchain-platform-prod --version ${PROD_VERSION} -n ${NAME} \
--set ca.enabled=true \
--set ca.app.arch=${ARCH} \
--set ca.ca.caAdminSecret=${CA_ADMIN_SECRET} \
--set ca.tlsca.name=tlsca \
--set ca.name=eca \
--set ca.service.type=NodePort \
--set ca.image.repository=${CA_IMAGE_REPO} \
--set ca.image.tag=${CA_TAG} \
--set ca.image.initimage=${CA_INIT_IMAGE_REPO} \
--set ca.image.inittag=${CA_INIT_TAG} \
--set ca.proxyIP=${PROXY_IP} \
--set ca.dataPVC.storageClassName=${STORAGE_CLASS} \
--set global.license=accept  \
--set global.multiarch=${MULTIARCH} \
--set ca.image.imagePullSecret=${PULL_SECRET_NAME} \
--set ca.resources.requests.cpu=${CA_CPU} \
--set ca.resources.requests.memory=${CA_MEMORY} \
--set ca.resources.limits.cpu=${CA_CPU_LIMIT} \
--set ca.resources.limits.memory=${CA_MEMORY_LIMIT} \
--tls

RC=$?
set +x

if [ $RC != 0 ]; then
    echo "CA $NAME Deployment Failed"
    exit 1
fi
