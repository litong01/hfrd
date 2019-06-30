#!/bin/bash
# This script extra images from nexus, then retag them
# and push these images onto docker hub

userid=""
password=""
if [[ $(uname -m) == 's390x' ]]; then
    tag="s390x"
else
    tag="amd64"
fi
if [ -z "$password" ]; then
  echo 'Password was not set, quitting'
  exit 1
fi

sourcerepo="nexus3.hyperledger.org:10001/hyperledger/"
targetrepo="hyperledger/"

version="1.4.1-stable"
declare -a imageNames=("fabric-tools" "fabric-ccenv" "fabric-orderer" "fabric-peer" "fabric-ca")
docker login -u ${userid} -p ${password}
for val in ${imageNames[@]}; do
   echo "Pulling image ${val}:${tag}-${version} from nexus"
   docker pull ${sourcerepo}${val}:${tag}-${version}
   docker tag ${sourcerepo}${val}:${tag}-${version} hfrd/${val}:${tag}-${version}
   docker push hfrd/${val}:${tag}-${version}
   docker rmi hfrd/${val}:${tag}-${version}
   docker rmi ${sourcerepo}${val}:${tag}-${version}
done

helperVersion="0.4.15"
declare -a imageNames=("fabric-baseos" "fabric-couchdb")
for val in ${imageNames[@]}; do
   echo "Pulling image ${val}:${tag}-${helperVersion} from nexus"
   docker pull ${sourcerepo}${val}:${tag}-${helperVersion}
   docker tag ${sourcerepo}${val}:${tag}-${helperVersion} hfrd/${val}:${tag}-${version}
   docker push hfrd/${val}:${tag}-${version}
   docker rmi hfrd/${val}:${tag}-${version}
   docker rmi ${sourcerepo}${val}:${tag}-${helperVersion}
done
