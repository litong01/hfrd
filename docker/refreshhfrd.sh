#!/bin/bash
# This script monitor hfrd repo. If there is any update, then
# it will rebuild all the docker images, and push them to docker
# hub, it will also restart the current running hfrd containers.
#
# This script assumes the following:
# 1. hfrd is extracted to ${HOME}/hl/src/hfrd directory
# 2. the working directory is in ${HOME}/refreshhfrd

# The following values must be set before you start the task
# The userid is the docker hub userid where hfrd images are
# pushed under. The password is the password for user hfrd.
# ip should be the server IP address.

userid="hfrd"
password=""
ip=""
if [[ $(uname -m) == 's390x' ]]; then
    tag="s390x"
else
    tag="amd64"
fi
if [ -z "$password" ]; then
  echo 'Password was not set, quitting'
  exit 1
fi
if [ -z "$ip" ]; then
  echo 'IP address was not set, quitting'
  exit 1
fi

cd ${HOME}/hl/src/hfrd
git fetch origin
# mycheck=$(git diff HEAD..origin/master --name-only | cut -d '/' -f 1 | sort -u)
mycheck=$(git log HEAD..origin/master --oneline)
if [  ! -z "$mycheck" ]; then
  netstatus=$(echo $mycheck | grep 'fatal: unable to access')

  if [ ! -z "$netstatus" ]; then
    echo 'Connection problem at '$(date) >> ${HOME}/refreshhfrd/refreshhfrd.log
  else
    echo 'Found some changes on '$(date)' , ready to build new images ...' >> ${HOME}/refreshhfrd/refreshhfrd.log

    echo "===Patch sets are===" >> ${HOME}/refreshhfrd/refreshhfrd.log
    echo "$mycheck" >> ${HOME}/refreshhfrd/refreshhfrd.log

    changedmodules=$(git diff HEAD..origin/master --name-only | cut -d '/' -f 1 | sort -u)
    echo "===Changed modules are===" >> ${HOME}/refreshhfrd/refreshhfrd.log
    echo "$changedmodules" >> ${HOME}/refreshhfrd/refreshhfrd.log

    git pull
    docker login -u ${userid} -p ${password}

    # Build hfrd/server:latest
    echo 'Building hfrd/server:latest container ...' >> ${HOME}/refreshhfrd/refreshhfrd.log
    make api-docker

    # Build hfrd/gosdk:latest
    echo 'Building hfrd/gosdk:latest container ...' >> ${HOME}/refreshhfrd/refreshhfrd.log
    make gosdk-docker

    # Build hfrd/jenkins container
    # For now we don't build jenkins automatically due to unexpected bugs in jenkins server duing restart
    # echo 'Building hfrd/jenkins:latest container ...' >> ${HOME}/refreshhfrd/refreshhfrd.log
    # cd ${HOME}/hl/src/hfrd/docker/jenkins-ansible
    # docker build -t hfrd/jenkins:${tag}-latest .

    ${HOME}/hl/src/hfrd/setup/hfrd.sh stop

    # Clean up jenkins src
    rm -rf ${HOME}/hfrd/jenkins/src/hfrd/backend/jenkins/pipelines
    mkdir -p ${HOME}/hfrd/jenkins/src/hfrd/backend/jenkins/pipelines
    cp -r ${HOME}/hl/src/hfrd/backend/jenkins/pipelines/* ${HOME}/hfrd/jenkins/src/hfrd/backend/jenkins/pipelines

    # Clean up containers and images
    # Do not remove jenkins container
    # docker rm $(docker ps -a -f status=exited -q)
    docker tag hfrd/server:${tag}-latest hfrd/server:latest
    docker tag hfrd/gosdk:${tag}-latest hfrd/gosdk:latest

    docker rmi -f $(docker images -f "dangling=true" -q)

    # Restart all the hfrd servers. Use false to disable the start in jenkins server
    ${HOME}/hl/src/hfrd/setup/hfrd.sh start ${ip} ${HOME}/hfrdconfig.xml false
    docker start jenkins
    # Now push all the hfrd images
    echo 'Push images to docker hub' >> ${HOME}/refreshhfrd/refreshhfrd.log
    docker push hfrd/server:${tag}-latest
    docker push hfrd/gosdk:${tag}-latest
    cd ${HOME}/hl/src/hfrd/docker/
    ./build_push_docker_manifest.sh ${HOME}/hl/src/hfrd/docker/
  fi
fi

# Schedule the next run in 5 minutes
cd ${HOME}/refreshhfrd
at now + 5 minute -f refreshhfrd.sh > /dev/null 2>&1
