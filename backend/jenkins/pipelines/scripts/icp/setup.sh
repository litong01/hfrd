#!/bin/bash -xe
source /opt/src/scripts/icp/config.cf

starttime=$(date +%s)
export ARCH=${ARCH:-"amd64"}
GLOBALNAME=${NAME:-$(echo bc${RANDOM})}
GLOBAL_NAMESPACE=${NAMESPACE:-"blockchain-dev"}
export ORG_NAME_PREFIX=${NAME}'org'
export ORDERER_ORG_NAME=${NAME}'ordererorg'

BASE_FOLDER=${BASE_FOLDER:-${PWD}/${GLOBALNAME}}
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

for ((i=0;i<${NUM_ORGS};i++))
do
   PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
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
done

# Delete all secrets and folders
NAME=${GLOBALNAME} ./cleanup.sh

# Download helmcharts source code
# Replace IBMCode with IBM-Blockchain AND replace hfrd with HelmCharts

git clone -b ${HELM_BRANCH} --single-branch ${SCM_URL}

#Create Image Pull Secret
if [ -n "${DOCKERHUB_USER}" ] && [ -n "${DOCKERHUB_PASS}" ] && [ -n "${DOCKERHUB_EMAIL}" ]; then
    echo "Creating docker pull secret..."
    export PULL_SECRET_NAME=${GLOBALNAME}-pullsecret
    kubectl create secret docker-registry ${PULL_SECRET_NAME} --docker-server=https://index.docker.io/v1/ --docker-username=${DOCKERHUB_USER} --docker-password=${DOCKERHUB_PASS} --docker-email=${DOCKERHUB_EMAIL}
fi

export PROXY_IP=${PROXY_IP:-$(kubectl get nodes --namespace ${GLOBAL_NAMESPACE} -l "proxy=true" -o jsonpath="{.items[0].status.addresses[0].address}")}
echo "PROXY IP: ${PROXY_IP}"

# Setup Orderer CA
kubectl create secret generic ${GLOBALNAME}-ordererorg-ca-admin-secret --from-literal=ca-admin-name=${GLOBALNAME}-ordererorgadmin --from-literal=ca-admin-password=${GLOBALNAME}-passwd
if [ -n "${CA_IMAGE_REPO}" ] && [[ "${CA_IMAGE_REPO}" =~ "${GLOBAL_NAMESPACE}" ]]; then #If Image name contains namespace then this means this is a local image
    NAME=${GLOBALNAME}-ordererorg-ca CA_ADMIN_SECRET=${GLOBALNAME}-ordererorg-ca-admin-secret CA_IMAGE_REPO=${CA_IMAGE_REPO} CA_TAG=${CA_TAG} ./create_ca.sh
elif [ -n "${CA_IMAGE_REPO}" ]; then #If Image name is passed but doesn't contains namespace then this means this is an image from an external docker repo
    NAME=${GLOBALNAME}-ordererorg-ca CA_ADMIN_SECRET=${GLOBALNAME}-ordererorg-ca-admin-secret CA_IMAGE_REPO=${CA_IMAGE_REPO} CA_TAG=${CA_TAG} MULTIARCH="true" ./create_ca.sh
else  #Else use default images
    NAME=${GLOBALNAME}-ordererorg-ca CA_ADMIN_SECRET=${GLOBALNAME}-ordererorg-ca-admin-secret ./create_ca.sh
fi


if [ $? != 0 ]; then
    echo "CA $NAME Deployment Failed"
    exit 1
fi

ORDERERORG_CA_SERVICE=$(kubectl get services --namespace ${GLOBAL_NAMESPACE} -l "app=ibm-ibp, release=${GLOBALNAME}-ordererorg-ca" -o jsonpath="{.items[0].metadata.name}")
ORDERERORG_CA_PORT=$(kubectl get --namespace ${GLOBAL_NAMESPACE} -o jsonpath="{.spec.ports[0].nodePort}" services $ORDERERORG_CA_SERVICE)
ORDERERORG_CA_HOST=${PROXY_IP}

# Setup Orgs CA

for ((i=0;i<${NUM_ORGS};i++))
do
    PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
    kubectl create secret generic ${GLOBALNAME}-${PEER_ORG_NAME}-ca-admin-secret --from-literal=ca-admin-name=${GLOBALNAME}-${PEER_ORG_NAME}admin --from-literal=ca-admin-password=${GLOBALNAME}-passwd
    if [ -n "${CA_IMAGE_REPO}" ] && [[ "${CA_IMAGE_REPO}" =~ "${GLOBAL_NAMESPACE}" ]]; then #If Image name contains namespace then this means this is a local image
        NAME=${GLOBALNAME}-${PEER_ORG_NAME}-ca CA_ADMIN_SECRET=${GLOBALNAME}-${PEER_ORG_NAME}-ca-admin-secret CA_IMAGE_REPO=${CA_IMAGE_REPO} CA_TAG=${CA_TAG} ./create_ca.sh
    elif [ -n "${CA_IMAGE_REPO}" ]; then #If Image name is passed but doesn't contains namespace then this means this is an image from an external docker repo
        NAME=${GLOBALNAME}-${PEER_ORG_NAME}-ca CA_ADMIN_SECRET=${GLOBALNAME}-${PEER_ORG_NAME}-ca-admin-secret CA_IMAGE_REPO=${CA_IMAGE_REPO} CA_TAG=${CA_TAG} MULTIARCH="true" ./create_ca.sh
    else  #Else use default images
        NAME=${GLOBALNAME}-${PEER_ORG_NAME}-ca CA_ADMIN_SECRET=${GLOBALNAME}-${PEER_ORG_NAME}-ca-admin-secret ./create_ca.sh
    fi

    if [ $? != 0 ]; then
        echo "CA $NAME Deployment Failed"
        exit 1
    fi

   export ${PEER_ORG_NAME}_CA_SERVICE=$(kubectl get services --namespace ${GLOBAL_NAMESPACE} -l "app=ibm-ibp, release=${GLOBALNAME}-${PEER_ORG_NAME}-ca" -o jsonpath="{.items[0].metadata.name}")
   var=${PEER_ORG_NAME}_CA_SERVICE
   export ${PEER_ORG_NAME}_CA_PORT=$(kubectl get --namespace ${GLOBAL_NAMESPACE} -o jsonpath="{.spec.ports[0].nodePort}" services ${!var})
   export ${PEER_ORG_NAME}_CA_HOST=$(kubectl get nodes --namespace ${GLOBAL_NAMESPACE} -l "proxy=true" -o jsonpath="{.items[0].status.addresses[0].address}")
done

NAME=${GLOBALNAME}-ordererorg-ca CA_HOST=${ORDERERORG_CA_HOST} CA_PORT=${ORDERERORG_CA_PORT} ./wait_for_ca.sh
for ((i=0;i<${NUM_ORGS};i++))
do
   PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
   var=${PEER_ORG_NAME}_CA_HOST
   var0=${PEER_ORG_NAME}_CA_PORT
   NAME=${GLOBALNAME}-${PEER_ORG_NAME}-ca CA_HOST=${!var} CA_PORT=${!var0} ./wait_for_ca.sh
done

for ((i=0;i<${NUM_ORGS};i++))
do
    PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
    cat << EOF > fabric-ca-server-config.yaml
    #############################################################################
    #   This is a configuration file for the fabric-ca-server command.
    #
    #   COMMAND LINE ARGUMENTS AND ENVIRONMENT VARIABLES
    #   ------------------------------------------------
    #   Each configuration element can be overridden via command line
    #   arguments or environment variables.  The precedence for determining
    #   the value of each element is as follows:
    #   1) command line argument
    #      Examples:
    #      a) --port 443
    #         To set the listening port
    #      b) --ca.keyfile ../mykey.pem
    #         To set the "keyfile" element in the "ca" section below;
    #         note the '.' separator character.
    #   2) environment variable
    #      Examples:
    #      a) FABRIC_CA_SERVER_PORT=443
    #         To set the listening port
    #      b) FABRIC_CA_SERVER_CA_KEYFILE="../mykey.pem"
    #         To set the "keyfile" element in the "ca" section below;
    #         note the '_' separator character.
    #   3) configuration file
    #   4) default value (if there is one)
    #      All default values are shown beside each element below.
    #
    #   FILE NAME ELEMENTS
    #   ------------------
    #   The value of all fields whose name ends with "file" or "files" are
    #   name or names of other files.
    #   For example, see "tls.certfile" and "tls.clientauth.certfiles".
    #   The value of each of these fields can be a simple filename, a
    #   relative path, or an absolute path.  If the value is not an
    #   absolute path, it is interpretted as being relative to the location
    #   of this configuration file.
    #
    #############################################################################

    # Version of config file
    version: 1.4.0

    # Server's listening port (default: 7054)
    port: 7054

    # Enables debug logging (default: false)
    debug: false

    # Size limit of an acceptable CRL in bytes (default: 512000)
    crlsizelimit: 512000

    #############################################################################
    #  TLS section for the server's listening port
    #
    #  The following types are supported for client authentication: NoClientCert,
    #  RequestClientCert, RequireAnyClientCert, VerifyClientCertIfGiven,
    #  and RequireAndVerifyClientCert.
    #
    #  Certfiles is a list of root certificate authorities that the server uses
    #  when verifying client certificates.
    #############################################################################
    tls:
      # Enable TLS (default: false)
      enabled: false
      # TLS for the server's listening port
      certfile:
      keyfile:
      clientauth:
        type: noclientcert
        certfiles:

    #############################################################################
    #  The CA section contains information related to the Certificate Authority
    #  including the name of the CA, which should be unique for all members
    #  of a blockchain network.  It also includes the key and certificate files
    #  used when issuing enrollment certificates (ECerts) and transaction
    #  certificates (TCerts).
    #  The chainfile (if it exists) contains the certificate chain which
    #  should be trusted for this CA, where the 1st in the chain is always the
    #  root CA certificate.
    #############################################################################
    ca:
      # Name of this CA
      name:
      # Key file (is only used to import a private key into BCCSP)
      keyfile:
      # Certificate file (default: ca-cert.pem)
      certfile:
      # Chain file
      chainfile:

    #############################################################################
    #  The gencrl REST endpoint is used to generate a CRL that contains revoked
    #  certificates. This section contains configuration options that are used
    #  during gencrl request processing.
    #############################################################################
    crl:
      # Specifies expiration for the generated CRL. The number of hours
      # specified by this property is added to the UTC time, the resulting time
      # is used to set the 'Next Update' date of the CRL.
      expiry: 24h

    #############################################################################
    #  The registry section controls how the fabric-ca-server does two things:
    #  1) authenticates enrollment requests which contain a username and password
    #     (also known as an enrollment ID and secret).
    #  2) once authenticated, retrieves the identity's attribute names and
    #     values which the fabric-ca-server optionally puts into TCerts
    #     which it issues for transacting on the Hyperledger Fabric blockchain.
    #     These attributes are useful for making access control decisions in
    #     chaincode.
    #  There are two main configuration options:
    #  1) The fabric-ca-server is the registry.
    #     This is true if "ldap.enabled" in the ldap section below is false.
    #  2) An LDAP server is the registry, in which case the fabric-ca-server
    #     calls the LDAP server to perform these tasks.
    #     This is true if "ldap.enabled" in the ldap section below is true,
    #     which means this "registry" section is ignored.
    #############################################################################
    registry:
      # Maximum number of times a password/secret can be reused for enrollment
      # (default: -1, which means there is no limit)
      maxenrollments: -1

      # Contains identity information which is used when LDAP is disabled
      identities:
        - name: ${GLOBALNAME}-${PEER_ORG_NAME}admin
          pass: ${GLOBALNAME}-passwd
          type: client
          affiliation: ""
          attrs:
            hf.Registrar.Roles: "*"
            hf.Registrar.DelegateRoles: "*"
            hf.Revoker: true
            hf.IntermediateCA: true
            hf.GenCRL: true
            hf.Registrar.Attributes: "*"
            hf.AffiliationMgr: true

    #############################################################################
    #  Database section
    #  Supported types are: "sqlite3", "postgres", and "mysql".
    #  The datasource value depends on the type.
    #  If the type is "sqlite3", the datasource value is a file name to use
    #  as the database store.  Since "sqlite3" is an embedded database, it
    #  may not be used if you want to run the fabric-ca-server in a cluster.
    #  To run the fabric-ca-server in a cluster, you must choose "postgres"
    #  or "mysql".
    #############################################################################
    db:
      type: sqlite3
      datasource: fabric-ca-server.db
      tls:
         enabled: false
         certfiles:
         client:
           certfile:
           keyfile:

    #############################################################################
    #  LDAP section
    #  If LDAP is enabled, the fabric-ca-server calls LDAP to:
    #  1) authenticate enrollment ID and secret (i.e. username and password)
    #     for enrollment requests;
    #  2) To retrieve identity attributes
    #############################################################################
    ldap:
       # Enables or disables the LDAP client (default: false)
       # If this is set to true, the "registry" section is ignored.
       enabled: false
       # The URL of the LDAP server
       url: ldap://<adminDN>:<adminPassword>@<host>:<port>/<base>
       # TLS configuration for the client connection to the LDAP server
       tls:
          certfiles:
          client:
             certfile:
             keyfile:
       # Attribute related configuration for mapping from LDAP entries to Fabric CA attributes
       attribute:
          # 'names' is an array of strings containing the LDAP attribute names which are
          # requested from the LDAP server for an LDAP identity's entry
          names: ['uid','member']
          # The 'converters' section is used to convert an LDAP entry to the value of
          # a fabric CA attribute.
          # For example, the following converts an LDAP 'uid' attribute
          # whose value begins with 'revoker' to a fabric CA attribute
          # named "hf.Revoker" with a value of "true" (because the boolean expression
          # evaluates to true).
          #    converters:
          #       - name: hf.Revoker
          #         value: attr("uid") =~ "revoker*"
          converters:
             - name:
               value:
          # The 'maps' section contains named maps which may be referenced by the 'map'
          # function in the 'converters' section to map LDAP responses to arbitrary values.
          # For example, assume a user has an LDAP attribute named 'member' which has multiple
          # values which are each a distinguished name (i.e. a DN). For simplicity, assume the
          # values of the 'member' attribute are 'dn1', 'dn2', and 'dn3'.
          # Further assume the following configuration.
          #    converters:
          #       - name: hf.Registrar.Roles
          #         value: map(attr("member"),"groups")
          #    maps:
          #       groups:
          #          - name: dn1
          #            value: peer
          #          - name: dn2
          #            value: client
          # The value of the user's 'hf.Registrar.Roles' attribute is then computed to be
          # "peer,client,dn3".  This is because the value of 'attr("member")' is
          # "dn1,dn2,dn3", and the call to 'map' with a 2nd argument of
          # "group" replaces "dn1" with "peer" and "dn2" with "client".
          maps:
            groups:
              - name:
                value:

    #############################################################################
    # Affiliations section. Fabric CA server can be bootstrapped with the
    # affiliations specified in this section. Affiliations are specified as maps.
    # For example:
    #   businessunit1:
    #     department1:
    #       - team1
    #   businessunit2:
    #     - department2
    #     - department3
    #
    # Affiliations are hierarchical in nature. In the above example,
    # department1 (used as businessunit1.department1) is the child of businessunit1.
    # team1 (used as businessunit1.department1.team1) is the child of department1.
    # department2 (used as businessunit2.department2) and department3 (businessunit2.department3)
    # are children of businessunit2.
    # Note: Affiliations are case sensitive except for the non-leaf affiliations
    # (like businessunit1, department1, businessunit2) that are specified in the configuration file,
    # which are always stored in lower case.
   #############################################################################
    affiliations:
       org1:
         - department1
         - department2
       org2:
         - department1

    #############################################################################
    #  Signing section
    #
    #  The "default" subsection is used to sign enrollment certificates;
    #  the default expiration ("expiry" field) is "8760h", which is 1 year in hours.
    #
    #  The "ca" profile subsection is used to sign intermediate CA certificates;
    #  the default expiration ("expiry" field) is "43800h" which is 5 years in hours.
    #  Note that "isca" is true, meaning that it issues a CA certificate.
    #  A maxpathlen of 0 means that the intermediate CA cannot issue other
    #  intermediate CA certificates, though it can still issue end entity certificates.
    #  (See RFC 5280, section 4.2.1.9)
    #
    #  The "tls" profile subsection is used to sign TLS certificate requests;
    #  the default expiration ("expiry" field) is "8760h", which is 1 year in hours.
    #############################################################################
    signing:
        default:
          usage:
           - digital signature
          expiry: 8760h
        profiles:
          ca:
             usage:
              - cert sign
              - crl sign
             expiry: 43800h
             caconstraint:
               isca: true
               maxpathlen: 0
          tls:
             usage:
               - signing
               - key encipherment
               - server auth
               - client auth
               - key agreement
             expiry: 8760h

    ###########################################################################
    #  Certificate Signing Request (CSR) section.
    #  This controls the creation of the root CA certificate.
    #  The expiration for the root CA certificate is configured with the
    #  "ca.expiry" field below, whose default value is "131400h" which is
    #  15 years in hours.
    #  The pathlength field is used to limit CA certificate hierarchy as described
    #  in section 4.2.1.9 of RFC 5280.
    #  Examples:
    #  1) No pathlength value means no limit is requested.
    #  2) pathlength == 1 means a limit of 1 is requested which is the default for
    #     a root CA.  This means the root CA can issue intermediate CA certificates,
    #     but these intermediate CAs may not in turn issue other CA certificates
    #     though they can still issue end entity certificates.
    #  3) pathlength == 0 means a limit of 0 is requested;
    #     this is the default for an intermediate CA, which means it can not issue
    #     CA certificates though it can still issue end entity certificates.
    ###########################################################################
    csr:
      cn: fabric-ca-server
      names:
       - C: US
         ST: "North Carolina"
         L:
         O: ${PEER_ORG_NAME}
         OU: Fabric
      hosts:
       - ${NAME}-${PEER_ORG_NAME}-ca-ca
       - localhost
      ca:
        expiry: 131400h
        pathlength: 1

    #############################################################################
    # BCCSP (BlockChain Crypto Service Provider) section is used to select which
    # crypto library implementation to use
    #############################################################################
    bccsp:
      default: SW
      sw:
        hash: SHA2
        security: 256
        filekeystore:
          # The directory used for the software file-based keystore
          keystore: msp/keystore

    #############################################################################
    # Multi CA section
    #
    # Each Fabric CA server contains one CA by default.  This section is used
    # to configure multiple CAs in a single server.
    #
    # 1) --cacount <number-of-CAs>
    # Automatically generate <number-of-CAs> non-default CAs.  The names of these
    # additional CAs are "ca1", "ca2", ... "caN", where "N" is <number-of-CAs>
    # This is particularly useful in a development environment to quickly set up
    # multiple CAs. Note that, this config option is not applicable to intermediate CA server
    # i.e., Fabric CA server that is started with intermediate.parentserver.url config
    # option (-u command line option)
    #
    # 2) --cafiles <CA-config-files>
    # For each CA config file in the list, generate a separate signing CA.  Each CA
    # config file in this list MAY contain all of the same elements as are found in
    # the server config file except port, debug, and tls sections.
    #
    # Examples:
    # fabric-ca-server start -b admin:adminpw --cacount 2
    #
    # fabric-ca-server start -b admin:adminpw --cafiles ca/ca1/fabric-ca-server-config.yaml
    # --cafiles ca/ca2/fabric-ca-server-config.yaml
    #
    #############################################################################

    cacount:

    cafiles:

    #############################################################################
    # Intermediate CA section
    #
    # The relationship between servers and CAs is as follows:
    #   1) A single server process may contain or function as one or more CAs.
    #      This is configured by the "Multi CA section" above.
    #   2) Each CA is either a root CA or an intermediate CA.
    #   3) Each intermediate CA has a parent CA which is either a root CA or another intermediate CA.
    #
    # This section pertains to configuration of #2 and #3.
    # If the "intermediate.parentserver.url" property is set,
    # then this is an intermediate CA with the specified parent
    # CA.
    #
    # parentserver section
    #    url - The URL of the parent server
    #    caname - Name of the CA to enroll within the server
    #
    # enrollment section used to enroll intermediate CA with parent CA
    #    profile - Name of the signing profile to use in issuing the certificate
    #    label - Label to use in HSM operations
    #
    # tls section for secure socket connection
    #   certfiles - PEM-encoded list of trusted root certificate files
    #   client:
    #     certfile - PEM-encoded certificate file for when client authentication
    #     is enabled on server
    #     keyfile - PEM-encoded key file for when client authentication
    #     is enabled on server
    #############################################################################
    intermediate:
      parentserver:
        url:
        caname:

      enrollment:
        hosts:
        profile:
        label:

      tls:
        certfiles:
        client:
          certfile:
          keyfile:
EOF
    ca_deploy_name="${GLOBALNAME}-${PEER_ORG_NAME}-ca-fabric-ca-deployment"
    CA_POD_NAME=$(kubectl get pods | grep "${ca_deploy_name}" | awk '{print $1}')
    kubectl exec -it ${CA_POD_NAME} -- ash -c 'rm -rf /etc/hyperledger/fabric-ca-server/ca/tls/*'
    kubectl exec -it ${CA_POD_NAME} -- ash -c 'rm -rf /etc/hyperledger/fabric-ca-server/msp'
    kubectl exec -it ${CA_POD_NAME} -- ash -c 'find /etc/hyperledger/fabric-ca-server/ -maxdepth 1 -type f -delete'

    kubectl cp fabric-ca-server-config.yaml ${CA_POD_NAME}:/etc/hyperledger/fabric-ca-server/
    kubectl delete pod ${CA_POD_NAME}
done

for ((i=0;i<${NUM_ORGS};i++))
do
   PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
   var=${PEER_ORG_NAME}_CA_HOST
   var0=${PEER_ORG_NAME}_CA_PORT
   NAME=${GLOBALNAME}-${PEER_ORG_NAME}-ca CA_HOST=${!var} CA_PORT=${!var0} ./wait_for_ca.sh
done

export VERSION=${FABRIC_VERSION:-1.2.1}
HOST_ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')")
CA_BINARY_FILE=hyperledger-fabric-ca-${HOST_ARCH}-${VERSION}.tar.gz
BINARY_FILE=hyperledger-fabric-${HOST_ARCH}-${VERSION}.tar.gz

function downloadBinaries() {
    curl -f -s -C - $1 -o $2 || rc=$?
    if [ ! -z "$rc" ]; then
	    echo "Failed to download the binaries , RC=$rc"
	    exit 1
    fi
    tar xzf ${2}
}

pwd
if [ ! -f ./bin/fabric-ca-client ] || [ ! -f ${CA_BINARY_FILE} ]; then
    echo "===> Downloading version ${VERSION} platform specific fabric-ca binaries"
    CA_URL=https://nexus.hyperledger.org/content/repositories/releases/org/hyperledger/fabric-ca/hyperledger-fabric-ca/${HOST_ARCH}-${VERSION}/${CA_BINARY_FILE}
    downloadBinaries ${CA_URL} ${CA_BINARY_FILE}
fi

if [ ! -f ./bin/peer ] || [ ! -f ${BINARY_FILE} ]; then
	echo "===> Downloading version ${VERSION} platform specific fabric binaries"
	URL=https://nexus.hyperledger.org/content/repositories/releases/org/hyperledger/fabric/hyperledger-fabric/${HOST_ARCH}-${VERSION}/${BINARY_FILE}
    downloadBinaries ${URL} ${BINARY_FILE}
fi

kubectl cp $(kubectl get po | grep ${GLOBALNAME}-ordererorg-ca | awk '{print $1}'):/etc/hyperledger/fabric-ca-server/ca-cert.pem ${ORDERER_CA_FOLDER}/tls-ca-cert.pem
for ((i=0;i<${NUM_ORGS};i++))
do
   PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
   var=${PEER_ORG_NAME}_CA_FOLDER
   kubectl cp $(kubectl get po | grep ${GLOBALNAME}-${PEER_ORG_NAME}-ca | awk '{print $1}'):/etc/hyperledger/fabric-ca-server/ca-cert.pem ${!var}/tls-ca-cert.pem
done

CSRHOSTS="${PROXY_IP},${NAME}-orderer-orderer,127.0.0.1"

FABRIC_CLIENT_RC=0
set -x
FABRIC_CA_CLIENT_HOME=${ORDERER_ECA_FOLDER} ./bin/fabric-ca-client enroll -u https://${GLOBALNAME}-ordererorgadmin:${GLOBALNAME}-passwd@${ORDERERORG_CA_HOST}:${ORDERERORG_CA_PORT} --caname eca --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_ECA_FOLDER} ./bin/fabric-ca-client register --caname eca --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --id.name ${GLOBALNAME}-orderer-admin --id.secret ${GLOBALNAME}-adminsecret --id.type user
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_ECA_FOLDER} ./bin/fabric-ca-client register --caname eca --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --id.name ${GLOBALNAME}-orderer-orderer --id.secret ${GLOBALNAME}-secret1 --id.type orderer
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_ADMIN_FOLDER} ./bin/fabric-ca-client enroll -u https://${GLOBALNAME}-orderer-admin:${GLOBALNAME}-adminsecret@${ORDERERORG_CA_HOST}:${ORDERERORG_CA_PORT} --caname eca --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_TLSCA_FOLDER} ./bin/fabric-ca-client enroll -u https://${GLOBALNAME}-ordererorgadmin:${GLOBALNAME}-passwd@${ORDERERORG_CA_HOST}:${ORDERERORG_CA_PORT} --caname tlsca --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${ORDERER_TLSCA_FOLDER} ./bin/fabric-ca-client register --caname tlsca --tls.certfiles ${ORDERER_CA_FOLDER}/tls-ca-cert.pem --id.name ${GLOBALNAME}-orderer-orderer --id.secret ${GLOBALNAME}-secret1 --id.type orderer
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

for ((i=0;i<${NUM_ORGS};i++))
do
   PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
   var=${PEER_ORG_NAME}_CA_FOLDER
   var0=${PEER_ORG_NAME}_CA_PORT
   var1=${PEER_ORG_NAME}_CA_HOST
   var2=${PEER_ORG_NAME}_ECA_FOLDER
   var3=${PEER_ORG_NAME}_ADMIN_FOLDER
   var4=${PEER_ORG_NAME}_TLSCA_FOLDER
   var5=${PEER_ORG_NAME}_USER_FOLDER
   CSRHOSTS="${PROXY_IP},${GLOBALNAME}-${PEER_ORG_NAME}-peer1,${GLOBALNAME}-${PEER_ORG_NAME}-peer2,127.0.0.1"
   FABRIC_CA_CLIENT_HOME=${!var2} ./bin/fabric-ca-client enroll -u https://${GLOBALNAME}-${PEER_ORG_NAME}admin:${GLOBALNAME}-passwd@${!var1}:${!var0} --caname eca --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
   FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

   FABRIC_CA_CLIENT_HOME=${!var2} ./bin/fabric-ca-client register --caname eca --tls.certfiles ${!var}/tls-ca-cert.pem --id.name ${GLOBALNAME}-${PEER_ORG_NAME}-peeradmin --id.secret ${GLOBALNAME}-adminsecret --id.type user
   FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

   FABRIC_CA_CLIENT_HOME=${!var2} ./bin/fabric-ca-client register --caname eca --tls.certfiles ${!var}/tls-ca-cert.pem --id.name wally --id.secret ${GLOBALNAME}-usersecret --id.type user
   FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

   FABRIC_CA_CLIENT_HOME=${!var3} ./bin/fabric-ca-client enroll -u https://${GLOBALNAME}-${PEER_ORG_NAME}-peeradmin:${GLOBALNAME}-adminsecret@${!var1}:${!var0} --caname eca --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
   FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

   FABRIC_CA_CLIENT_HOME=${!var5} ./bin/fabric-ca-client enroll -u https://wally:${GLOBALNAME}-usersecret@${!var1}:${!var0} --caname eca --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
   FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

   FABRIC_CA_CLIENT_HOME=${!var4} ./bin/fabric-ca-client enroll -u https://${GLOBALNAME}-${PEER_ORG_NAME}admin:${GLOBALNAME}-passwd@${!var1}:${!var0} --caname tlsca --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
   FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))


   for ((peer_num=0;peer_num<${PEERS_PER_ORG};peer_num++))
   do
      FABRIC_CA_CLIENT_HOME=${!var2} ./bin/fabric-ca-client register --caname eca --tls.certfiles ${!var}/tls-ca-cert.pem --id.name ${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num} --id.secret ${GLOBALNAME}-secret1 --id.type peer
      FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))
      FABRIC_CA_CLIENT_HOME=${!var4} ./bin/fabric-ca-client register --caname tlsca --tls.certfiles ${!var}/tls-ca-cert.pem --id.name ${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num} --id.secret ${GLOBALNAME}-secret1 --id.type peer
      FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))
   done
done


if [ $FABRIC_CLIENT_RC -gt 0 ]; then
    echo "Fabric CA Client failure"
    exit 1
fi

CSRHOSTS="${PROXY_IP},${NAME}-orderer-orderer,127.0.0.1"
NAME=${GLOBALNAME}-orderer-msp LOCATION=${ORDERER_CA_FOLDER}/secret.json CA_HOST=${ORDERERORG_CA_HOST} CA_PORT=${ORDERERORG_CA_PORT} ENROLL_ID=${GLOBALNAME}-orderer-orderer ENROLL_SECRET=${GLOBALNAME}-secret1 CACERT=${ORDERER_CA_FOLDER}/tls-ca-cert.pem ADMINCERT=${ORDERER_ADMIN_FOLDER}/msp/signcerts/*.pem CSRHOSTS=\"${CSRHOSTS}\" CA_NAME=eca TLSCA_NAME=tlsca ./create_msp_secret.sh

for ((i=0;i<${NUM_ORGS};i++))
do
   PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
   var=${PEER_ORG_NAME}_CA_FOLDER
   var0=${PEER_ORG_NAME}_CA_PORT
   var1=${PEER_ORG_NAME}_CA_HOST
   var3=${PEER_ORG_NAME}_ADMIN_FOLDER
   for ((peer_num=0;peer_num<${PEERS_PER_ORG};peer_num++))
   do
      CSRHOSTS="${PROXY_IP},${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num},127.0.0.1"
      NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}-msp LOCATION=${!var}/secret.json CA_HOST=${!var1} CA_PORT=${!var0} ENROLL_ID=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num} ENROLL_SECRET=${GLOBALNAME}-secret1 CACERT=${!var}/tls-ca-cert.pem ADMINCERT=${!var3}/msp/signcerts/*.pem CSRHOSTS=\"${CSRHOSTS}\" CA_NAME=eca TLSCA_NAME=tlsca ./create_msp_secret.sh
   done
done
sleep 3

ORDERER_MSP_ID=${ORDERER_ORG_NAME}
if [ -n "${ORDERER_IMAGE_REPO}" ] && [[ "${ORDERER_IMAGE_REPO}" =~ "${GLOBAL_NAMESPACE}" ]]; then #If Image name contains namespace then this means this is a local image
    NAME=${GLOBALNAME}-orderer MSP_SECRET=${GLOBALNAME}-orderer-msp ORG_NAME=${ORDERER_ORG_NAME} ORDERER_MSP_ID=${ORDERER_MSP_ID} ORDERER_IMAGE_REPO=${ORDERER_IMAGE_REPO} ORDERER_TAG=${ORDERER_TAG} ORDERER_INIT_IMAGE_REPO=${ORDERER_INIT_IMAGE_REPO} ORDERER_INIT_TAG=${ORDERER_INIT_TAG} ./create_orderer.sh
elif [ -n "${ORDERER_IMAGE_REPO}" ]; then #If Image name is passed but doesn't contains namespace then this means this is an image from an external docker repo
    NAME=${GLOBALNAME}-orderer MSP_SECRET=${GLOBALNAME}-orderer-msp ORG_NAME=${ORDERER_ORG_NAME} ORDERER_MSP_ID=${ORDERER_MSP_ID} ORDERER_IMAGE_REPO=${ORDERER_IMAGE_REPO} ORDERER_TAG=${ORDERER_TAG} ORDERER_INIT_IMAGE_REPO=${ORDERER_INIT_IMAGE_REPO} ORDERER_INIT_TAG=${ORDERER_INIT_TAG} MULTIARCH="true" ./create_orderer.sh
else  #Else use default images
    NAME=${GLOBALNAME}-orderer MSP_SECRET=${GLOBALNAME}-orderer-msp ORG_NAME=${ORDERER_ORG_NAME} ORDERER_MSP_ID=${ORDERER_MSP_ID} ./create_orderer.sh
fi

if [ $? != 0 ]; then
    echo "Orderer $NAME Deployment Failed"
    exit 1
fi

set +x

echo "Waiting for pod ${POD_NAME} to start completion. Status = ${POD_STATUS}, Readiness = ${IS_READY}"

SECONDS=0
while (( $SECONDS < 600 ));
do
    POD_NAME=${GLOBALNAME}-orderer-orderer
    POD_STATUS=$(kubectl get pods | grep "${POD_NAME}" | awk '{print $3}')
    IS_READY=$(kubectl get pods | grep "${POD_NAME}" | awk '{print $2}')
    TOTAL_PODS=$(echo $IS_READY | cut -d'/' -f2)
    if [ "${IS_READY}" == "${TOTAL_PODS}/${TOTAL_PODS}" ]; then
        break
    fi
    echo "Waiting for pod ${POD_NAME} to start completion. Status = ${POD_STATUS}, Readiness = ${IS_READY}"
    sleep 3
done

if [ $SECONDS -ge 600 ]
then
    echo "Timed out waiting for pod ${POD_NAME} to start completion"
    exit 1
fi

set -x

for ((i=0;i<${NUM_ORGS};i++))
do
   PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
   export ${PEER_ORG_NAME}_MSP="${PEER_ORG_NAME}"
   var=${PEER_ORG_NAME}_MSP
   for ((peer_num=0;peer_num<${PEERS_PER_ORG};peer_num++))
   do
       if [ -n "${PEER_IMAGE_REPO}" ] && [[ "${PEER_IMAGE_REPO}" =~ "${GLOBAL_NAMESPACE}" ]]; then #If Image name contains namespace then this means this is a local image
       NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num} MSP_SECRET=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}-msp ORGMSP=${!var} PEER_IMAGE_REPO=${PEER_IMAGE_REPO} PEER_TAG=${PEER_TAG} PEER_DIND_IMAGE_REPO=${PEER_DIND_IMAGE_REPO} PEER_DIND_TAG=${PEER_DIND_TAG} PEER_INIT_IMAGE_REPO=${PEER_INIT_IMAGE_REPO} ./create_peer.sh
       elif [ -n "${PEER_IMAGE_REPO}" ]; then #If Image name is passed but doesn't contains namespace then this means this is an image from an external docker repo
       NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num} MSP_SECRET=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}-msp ORGMSP=${!var} PEER_IMAGE_REPO=${PEER_IMAGE_REPO} PEER_TAG=${PEER_TAG} PEER_DIND_IMAGE_REPO=${PEER_DIND_IMAGE_REPO} PEER_DIND_TAG=${PEER_DIND_TAG} PEER_INIT_IMAGE_REPO=${PEER_INIT_IMAGE_REPO} MULTIARCH="true" ./create_peer.sh
       else  #Else use default images
       NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num} MSP_SECRET=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}-msp ORGMSP=${!var} ./create_peer.sh
       fi

       if [ $? != 0 ]; then
       echo "Peer $NAME Deployment Failed"
       exit 1
       fi

       SECONDS=0
       while (( $SECONDS < 600 ));
       do
           POD_NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}
           POD_STATUS=$(kubectl get pods | grep "${POD_NAME}" | awk '{print $3}')
           if [ "${POD_STATUS}" == "Running" ]; then
               break
           fi
           echo "Waiting for pod ${POD_NAME} to finish init. Status = ${POD_STATUS}"
           sleep 3
       done

       if [ $SECONDS -ge 600 ]
       then
           echo "Timed out waiting for pod ${POD_NAME} to finish init"
           exit 1
       fi

       # Set external endpoint for gossip across ORGS not included for parity
       POD_NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}
       if [ ${peer_num} -eq 1 ] && [ ${peer_num} -gt 1 ]
       then
           OTHER_POD_NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer2
       else
           OTHER_POD_NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer1
       fi
       PEER_PORT=$(kubectl get svc ${POD_NAME} | grep NodePort | awk -F '[[:space:]:/]+' '{print $6}')
       PEER_ADDRESS="${PROXY_IP}:${PEER_PORT}"
       kubectl patch deployment ${POD_NAME} -p '{"spec":{"strategy":{"rollingUpdate": null, "type": "Recreate"}}}'
       kubectl set env deployment ${POD_NAME} -c peer "CORE_PEER_ID=${POD_NAME}" "CORE_PEER_ADDRESS=${POD_NAME}:7051" "CORE_PEER_GOSSIP_EXTERNALENDPOINT=${PEER_ADDRESS}" "CORE_PEER_GOSSIP_BOOTSTRAP=${OTHER_POD_NAME}:7051" #"FABRIC_LOGGING_SPEC=grpc=debug:gossip=debug:info"

       SECONDS=0
       while (( $SECONDS < 600 ));
       do
           POD_NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}
           POD_STATUS=$(kubectl get pods | grep "${POD_NAME}" | awk '{print $3}')
           if [ "${POD_STATUS}" == "Running" ]; then
               break
           fi
           echo "Waiting for pod ${POD_NAME} to finish init. Status = ${POD_STATUS}"
           sleep 3
       done

       if [ $SECONDS -ge 600 ]
       then
           echo "Timed out waiting for pod ${POD_NAME} to finish init"
           exit 1
       fi
   done
done

set +x

for ((i=0;i<${NUM_ORGS};i++))
do
    PEER_ORG_NAME=${ORG_NAME_PREFIX}${i}
    for ((peer_num=0;peer_num<${PEERS_PER_ORG};peer_num++))
    do
       SECONDS=0
       while (( $SECONDS < 600 ));
       do
           POD_NAME=${GLOBALNAME}-${PEER_ORG_NAME}-peer${peer_num}
           POD_STATUS=$(kubectl get pods | grep "${POD_NAME}" | awk '{print $3}')
           IS_READY=$(kubectl get pods | grep "${POD_NAME}" | awk '{print $2}')
           TOTAL_PODS=$(echo $IS_READY | cut -d'/' -f2)
           if [ "${IS_READY}" == "${TOTAL_PODS}/${TOTAL_PODS}" ]; then
               break
           fi
           echo "Waiting for pod ${POD_NAME} to start completion. Status = ${POD_STATUS}, Readiness = ${IS_READY}"
           sleep 3
       done

       if [ $SECONDS -ge 600 ]
       then
           echo "Timed out waiting for pod ${POD_NAME} to start completion"
           exit 1
       fi
    done
done
echo -e "\n\nTotal execution time for launching blockchain network: $(($(date +%s)-starttime)) secs ...\n"
