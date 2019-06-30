#!/bin/bash
manifest_dir=$1
# Step 1 : make sure manifest-tool exists https://github.com/estesp/manifest-tool
if [ ! -f "$GOPATH/src/github.com/estesp/manifest-tool/manifest-tool" ]; then
  echo 'Error: manifest-tool is not installed.Start to install'
  cd $GOPATH/src
  mkdir -p github.com/estesp
  cd github.com/estesp
  git clone https://github.com/estesp/manifest-tool
  cd manifest-tool && make binary
fi

# Step 2 : create and push manifest
cd $GOPATH/src/github.com/estesp/manifest-tool
./manifest-tool push from-spec ${manifest_dir}/manifest/hfrdserver.yaml
./manifest-tool push from-spec ${manifest_dir}/manifest/gosdk.yaml
./manifest-tool push from-spec ${manifest_dir}/manifest/jenkins.yaml
./manifest-tool push from-spec ${manifest_dir}/manifest/nfs-provisioner.yaml

