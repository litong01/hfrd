package common

const (
	// global common parameters required by gosdk cli
	CONN_PROFILE       = "connectionProfile"
	NAME               = "name" // test name
	ITERATION_COUNT    = "iterationCount"
	ITERATION_INTERVAL = "iterationInterval"
	DELAY_TIME         = "delayTime"
	RETRY_COUNT        = "retryCount"
	LOG_LEVEL          = "logLevel"

	ADMIN = "Admin"
	USER  = "User1"

	// BCCSP Default Configurations
	CONFIG_BCCSP = "config/config_crypto_bccsp.yaml"

	IGNORE_ERRORS     = "ignoreErrors"
	CONCURRENCY_LIMIT = "concurrencyLimit" // in most cases users only need to config this param for cc invoke
)
