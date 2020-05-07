#!/bin/bash

START_STOP="$1"
PUBLIC_IP="$2"
CONFIG_FILE_PATH="$3"
IS_START_JENKINS=true
if [[ $4 == "false" ]]; then
  IS_START_JENKINS=false
fi

if [[ $(uname -m) == 's390x' ]]; then
    IMAGE_ARCH='-s390x'
    HTTPD_IMAGE_NAME='s390x/httpd:2.4.34-alpine'
    COUCHDB_IMAGE_NAME='hfrd/couchdb-s390x'
else
    IMAGE_ARCH=''
    HTTPD_IMAGE_NAME='httpd:2.4.34-alpine'
    COUCHDB_IMAGE_NAME='couchdb:2.3.1'
fi

rootdir=~/hfrd

function printHelp () {
    echo "Usage: ./hfrd.sh <start {public ip} {configFile path} | stop >"
    echo "For example:"
    echo "If you want to start hfrd service,you must provide 3 parameters."
    echo "      ./hfrd.sh start {host public ip} {config file path} "
    echo "      ./hfrd.sh start 9.42.130.26 /home/ibmadmin/hl/src/hfrd/setup/org.jenkinsci.plugins.configfiles.GlobalConfigFiles.xml "
    echo "If you want to stop hfrd service,you only need to run"
    echo "      ./hfrd.sh stop"
    exit 1
}
function doscript() {
  cat > $rootdir/jenkins/scriptApproval.xml << 'endmsg'
<?xml version='1.1' encoding='UTF-8'?>
<scriptApproval plugin="script-security@1.46">
  <approvedScriptHashes/>
  <approvedSignatures>
    <string>method hudson.plugins.git.GitSCM getUserRemoteConfigs</string>
    <string>method hudson.plugins.git.UserRemoteConfig getUrl</string>
    <string>method groovy.text.Template make java.util.Map</string>
    <string>method groovy.text.TemplateEngine createTemplate java.lang.String</string>
    <string>method hudson.plugins.git.GitSCM getUserRemoteConfigs</string>
    <string>method hudson.plugins.git.UserRemoteConfig getUrl</string>
    <string>new groovy.text.GStringTemplateEngine</string>
    <string>staticMethod java.lang.System getenv java.lang.String</string>
    <string>method org.apache.commons.collections.KeyValue getValue</string>
  </approvedSignatures>
  <aclApprovedSignatures/>
  <approvedClasspathEntries/>
  <pendingScripts/>
  <pendingSignatures/>
  <pendingClasspathEntries/>
</scriptApproval>
endmsg
  cat > $rootdir/jenkins/config.xml << 'endmsg'
<?xml version='1.1' encoding='UTF-8'?>
<hudson>
  <disabledAdministrativeMonitors/>
  <version>1.0</version>
  <installStateName>RUNNING</installStateName>
  <numExecutors>30</numExecutors>
  <mode>NORMAL</mode>
  <useSecurity>false</useSecurity>
  <authorizationStrategy class="hudson.security.AuthorizationStrategy$Unsecured"/>
  <securityRealm class="hudson.security.SecurityRealm$None"/>
  <disableRememberMe>false</disableRememberMe>
  <projectNamingStrategy class="jenkins.model.ProjectNamingStrategy$DefaultProjectNamingStrategy"/>
  <workspaceDir>${JENKINS_HOME}/workspace/${ITEM_FULL_NAME}</workspaceDir>
  <buildsDir>${ITEM_ROOTDIR}/builds</buildsDir>
  <jdks/>
  <viewsTabBar class="hudson.views.DefaultViewsTabBar"/>
  <myViewsTabBar class="hudson.views.DefaultMyViewsTabBar"/>
  <clouds/>
  <scmCheckoutRetryCount>0</scmCheckoutRetryCount>
  <views>
    <hudson.model.AllView>
      <owner class="hudson" reference="../../.."/>
      <name>all</name>
      <filterExecutors>false</filterExecutors>
      <filterQueue>false</filterQueue>
      <properties class="hudson.model.View$PropertyList"/>
    </hudson.model.AllView>
  </views>
  <primaryView>all</primaryView>
  <slaveAgentPort>50000</slaveAgentPort>
  <label></label>
  <nodeProperties/>
  <globalNodeProperties/>
</hudson>
endmsg

}

function starthfrd() {
    if [ "$CONFIG_FILE_PATH" = "" ] || [ "$PUBLIC_IP" = "" ]; then
        echo "Error: Syntax error when start hfrd service"
        echo "Please make sure you have provided all the required parameters"
        echo ""
        printHelp
    fi
    if [[ ! -f ${CONFIG_FILE_PATH} ]]; then
    	echo "Missing hfrd config file, cannot continue.Please make sure the config file path is correct"
    	exit 1
    fi

    mkdir -p $rootdir/couchdb/data $rootdir/contentRepo \
        $rootdir/var $rootdir/jenkins
    cp -f ${CONFIG_FILE_PATH} $rootdir/jenkins/org.jenkinsci.plugins.configfiles.GlobalConfigFiles.xml

    docker run -d --rm --name couchdb \
        -v $rootdir/couchdb/data:/opt/couchdb/data \
        -p 5984:5984 ${COUCHDB_IMAGE_NAME}

    while : ; do
        res=$(docker logs couchdb 2>&1 | grep 'Apache CouchDB has started')
        if [[ ! -z $res ]]; then
            break
        fi
        echo 'Waiting for couchdb to be ready...'
        sleep 3
    done

    docker run -d --rm --name hfrdapache -p 9696:80 \
      -v $rootdir/contentRepo:/usr/local/apache2/htdocs/ \
      ${HTTPD_IMAGE_NAME}
    if ${IS_START_JENKINS}; then
        echo "Start jenkins"
        doscript
        docker run -d --rm --name jenkins --env HOST_JENKINS="$rootdir/jenkins" \
            -v /var/run/docker.sock:/var/run/docker.sock \
            -v $rootdir/jenkins:/var/jenkins_home \
            -v $rootdir/contentRepo:/opt/hfrd/contentRepo \
            -p 8080:8080 -p 50000:50000 hfrd/jenkins:latest

        while : ; do
          res=$(docker logs jenkins 2>&1 | grep 'Jenkins is fully up and running')
          if [[ ! -z $res ]]; then
            if [ ! "$(ls -A $rootdir/jenkins/jobs/)" ]; then
              docker exec -it jenkins sudo sed -i -e "s/localhost/${PUBLIC_IP}/g" \
                /usr/share/hfrd/jjb/hfrd-jenkins-jobs.yaml
              docker exec -it jenkins sudo jenkins-jobs --conf \
                /usr/share/hfrd/jenkins.ini update /usr/share/hfrd/jjb
            fi
            break
          fi
          echo 'Waiting for jenkins server to be ready...'
          sleep 3
        done
    fi

    echo 'Generating hfrdserver configuration'
cat > $rootdir/var/config.json << 'endmsg'
{
  "jenkins": {
    "baseUrl": "http://admin:admin@TTTipTTT:8080",
    "crumbUrl": "/crumbIssuer/api/json",
    "jobUrl": "/job/{JobName}/api/xml?tree=builds[id,result,queueId]&xpath=\/\/build[queueId={QueueId}]",
    "jobStatusUrl": "/job/{JobName}/{JobId}/artifact/workdir/results/jobStatus.json",
    "serviceGetByServiceId": "job/{JobName}/{JobId}/artifact/workdir/results/package.tar",
    "buildUrl": "/job/{JobName}/buildWithParameters"
  },
  "log": {
    "level": "debug"
  },
  "auth": {
    "enabled": false,
    "type": "jwt"
  },
  "couchUrl": "http://TTTipTTT:5984",
  "allowOrigins": ["http://TTTipTTT:9090"],
  "contentRepo": "/opt/hfrd/contentRepo",
  "apacheBaseUrl": "http://TTTipTTT:9696"
}
endmsg

    sed -i -e "s/TTTipTTT/${PUBLIC_IP}/g" $rootdir/var/config.json

    docker run -d --rm --name hfrdserver \
      -v $rootdir/contentRepo:/opt/hfrd/contentRepo \
      -v $rootdir/var:/opt/hfrd/var \
      -p 9090:8080 hfrd/server:latest

    echo "API server http://${PUBLIC_IP}:9090"
    echo "Jenkins server http://${PUBLIC_IP}:8080"
}

function stophfrd(){
    docker rm -f hfrdserver
    docker rm -f couchdb
    docker rm -f hfrdapache
}

if [ "${START_STOP}" == "stop" ]; then
    stophfrd
elif [ "${START_STOP}" == "start" ]; then
    starthfrd
else
    printHelp
fi
