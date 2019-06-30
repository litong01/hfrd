package resource

import "path/filepath"

const (
	// Query parameters related constants
	ENV            = "env"
	REQUEST_ID_REQ = "requestid"
	REQUEST_ID_RES = "Request-Id"
	UID            = "uid"
	LOCATION       = "loc"
	PLAN_NAME      = "name"

	// environment names
	BX_STAGING  = "bxstaging"    // Bluemix staging: supports STARTER and ENTERPRISE plans
	BX_PROD     = "bxproduction" // Bluemix production: supports STARTER and ENTERPRISE plans
	CM          = "cm"           // cluster manager: supports ENTERPRISE plan only
	DEFAULT_ENV = BX_STAGING

	// Plan names
	STARTER    = "sp" // Starter plan: available in BX_* environments
	ENTERPRISE = "ep" // Enterprise plan: available on BX_* and CM environments

	// Network configurations for cluster manager environment
	// ONLY required for cm environment
	DEFAULT_NUM_OF_ORGS  = 1 // number or organizations per network
	DEFAULT_NUM_OF_PEERS = 2 // number of peers per organization
	NUM_OF_ORGS_KEY      = "numOfOrgs"
	NUM_OF_PEERS_KEY     = "numOfPeers"
	LEDGER_TYPE_KEY      = "ledgerType"
	// See https://github.ibm.com/IBM-Blockchain/manager-v1/blob/49d29b036f9435da489d7b36ad18eeb30dfa8cac/bluemix-fabric-functions.js#L521
	LEVELDB   = "levelDB"
	COUCHDB   = "couch"
	QUEUEID   = "queueid"
	CHARTPATH = "chartpath"

	// certs tar support
	CERT_VERSION_KEY = "certsVersion"
	CERT_V1          = "v1"
	CERT_V2          = "v2"

	CONN_JSON = "connection.json"
	CONN_YAML = "connection.yml"
	IDENTITY  = "identity.json"
	KEYFILES  = "keyfiles"
	SEP       = string(filepath.Separator)
)

var (
	// Supported environments and plans
	SUPPORTED_ENV = map[string][]string{
		BX_STAGING: []string{STARTER, ENTERPRISE},
		BX_PROD:    []string{STARTER, ENTERPRISE},
		CM:         []string{ENTERPRISE},
	}
)
