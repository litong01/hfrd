package jenkins

const (
	// job names
	NETWORK       = "network"
	CONNECTION    = "connection"
	TEST          = "test"
	NETWORK_ICP   = "network-icp"
	NETWORK_CM    = "network-cm"
	CONNECTION_CM = "connection-cm"
	K8SNETWORK    = "k8snetwork"
	MODULETEST    = "moduletest"
	IBPCERTS      = "ibpcerts"
	ADDORG        = "addorg"

	// job status
	SUCCESS    = "SUCCESS"
	FAIL       = "FAIL"
	INPROGRESS = "INPROGRESS"

	// job parameters
	SERVICE_ID = "serviceid"
	METHOD     = "method"
	LOCATION   = "loc"
	GET        = "GET"
	POST       = "POST"
	DELETE     = "DELETE"
	TESTCONFIG = "testconfig"

	// k8snetwork and module test
	REQUESTID   = "requestid"
	UID         = "uid"
	ACTION      = "action"
	APACHE_BASE = "specpath" // apache base url. e.g. http://hfrdrestsrv.rtp.raleigh.ibm.com

	// add organizations into existing channels
	ORGNAME      = "orgName"
	ORGCERTSFILE = "orgCertsFile"
	CHANNELS     = "channels"
)
