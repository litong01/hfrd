#!/bin/bash -xe

org_name=$1
work_dir=$2
binary_url=$3
ca_name=${org_name}'ca'

source $work_dir'/apis.ini' || true


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

if [ ! -f $work_dir'/bin/cloudctl' ]; then

  curl -kLo cloudctl-linux-amd64-v3.2.0-634 https://$icp_url/api/cli/cloudctl-linux-amd64
  mv cloudctl* $work_dir'/bin/cloudctl'
  chmod +x $work_dir'/bin/cloudctl'

  curl -kLo kubectl-linux-amd64-v1.13.5 https://$icp_url/api/cli/kubectl-linux-amd64
  mv kubectl* $work_dir'/bin/kubectl'
  chmod +x $work_dir'/bin/kubectl'
fi

export PATH=$PATH:$work_dir/bin

cloudctl login -a https://$icp_url --skip-ssl-validation -u $icp_user -p $icp_password -n $icp_namespace


BASE_FOLDER=$work_dir'/crypto-config'

PEER_ORG_NAME=${org_name}
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


IFS=':' read -ra ADDR <<< "$CA_URL"
export PROXY_IP=${ADDR[0]}
export ${PEER_ORG_NAME}_CA_HOST=${ADDR[0]}
export ${PEER_ORG_NAME}_CA_PORT=${ADDR[1]}

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
        - name: admin
          pass: pass4chain
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
       - ${PEER_ORG_NAME}ca
       - localhost
       - 127.0.0.1
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

var0=${PEER_ORG_NAME}_CA_HOST
var1=${PEER_ORG_NAME}_CA_PORT
NAME=${PEER_ORG_NAME}ca CA_HOST=${!var0} CA_PORT=${!var1} ./wait_for_pod.sh

ca_deploy_name="${PEER_ORG_NAME}ca"
CA_POD_NAME=$(kubectl get pods | grep "${ca_deploy_name}" | awk '{print $1}')
kubectl exec -it ${CA_POD_NAME} -- ash -c 'rm -rf /etc/hyperledger/fabric-ca-server/ca/tls/*'
kubectl exec -it ${CA_POD_NAME} -- ash -c 'rm -rf /etc/hyperledger/fabric-ca-server/msp'
kubectl exec -it ${CA_POD_NAME} -- ash -c 'rm -f /etc/hyperledger/fabric-ca-server/fabric-ca-server-config.yaml'
kubectl exec -it ${CA_POD_NAME} -- ash -c 'find /etc/hyperledger/fabric-ca-server/ -maxdepth 1 -type f -delete'

kubectl cp fabric-ca-server-config.yaml ${CA_POD_NAME}:/etc/hyperledger/fabric-ca-server/
kubectl delete pod ${CA_POD_NAME}

var0=${PEER_ORG_NAME}_CA_HOST
var1=${PEER_ORG_NAME}_CA_PORT
NAME=${PEER_ORG_NAME}ca CA_HOST=${!var0} CA_PORT=${!var1} ./wait_for_pod.sh

var0=${PEER_ORG_NAME}_CA_FOLDER
echo $TLS_CERT | base64 -d -w 0 > ${!var0}/tls-ca-cert.pem


var=${PEER_ORG_NAME}_CA_FOLDER
var0=${PEER_ORG_NAME}_CA_PORT
var1=${PEER_ORG_NAME}_CA_HOST
var2=${PEER_ORG_NAME}_ECA_FOLDER
var3=${PEER_ORG_NAME}_ADMIN_FOLDER
var4=${PEER_ORG_NAME}_TLSCA_FOLDER
var5=${PEER_ORG_NAME}_USER_FOLDER

CSRHOSTS="${PROXY_IP},${PEER_ORG_NAME},127.0.0.1,localhost"
FABRIC_CA_CLIENT_HOME=${!var2} fabric-ca-client enroll -u https://admin:pass4chain@${!var1}:${!var0} --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var2} fabric-ca-client register --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --id.name peeradmin --id.secret pass4chain --id.type user
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var2} fabric-ca-client register --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --id.name wally --id.secret pass4chain --id.type user
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var3} fabric-ca-client enroll -u https://peeradmin:pass4chain@${!var1}:${!var0} --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var5} fabric-ca-client enroll -u https://wally:pass4chain@${!var1}:${!var0} --caname ${CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var4} fabric-ca-client enroll -u https://admin:pass4chain@${!var1}:${!var0} --caname ${TLS_CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --csr.hosts ${CSRHOSTS}
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

FABRIC_CA_CLIENT_HOME=${!var4} fabric-ca-client register --caname ${TLS_CA_NAME} --tls.certfiles ${!var}/tls-ca-cert.pem --id.name peertls --id.secret pass4chain --id.type peer
FABRIC_CLIENT_RC=$(($FABRIC_CLIENT_RC + $?))

peer_signed_cert=$(cat $work_dir/crypto-config/${org_name}/admin/msp/signcerts/cert.pem | base64 -w 0)
root_certs=$(cat $work_dir/crypto-config/${org_name}/ca/enrollment/msp/signcerts/cert.pem | base64 -w 0)
tls_root_certs=$(cat $work_dir/crypto-config/${org_name}/ca/tls/msp/signcerts/cert.pem | base64 -w 0)

echo $peer_signed_cert > $work_dir/crypto-config/${org_name}/peer_signed_cert
echo $TLS_CERT > $work_dir/crypto-config/${org_name}/ca_tls_cert
echo $root_certs > $work_dir/crypto-config/${org_name}/ca_admin_cert
echo $tls_root_certs > $work_dir/crypto-config/${org_name}/tls_ca_cert
