package hfrdlogging

import (
	"gopkg.in/op/go-logging.v1"
)

var (
	logger *logging.Logger
)

const (
	defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
	defaultLevel  = logging.INFO

	// log modules
	MODULE_AUTH       = "auth"
	MODULE_UTILS      = "utils"
	MODULE_COUCH      = "couch"
	MODULE_FILTER     = "filter"
	MODULE_JKS        = "jenkins"
	MODULE_JWK        = "jwk"
	MODULE_MAIN       = "main"
	MODULE_ROUTER     = "router"
	MODULE_RESOURCE   = "resource"
	MODULE_ROUTINE    = "routine"
	MODULE_UIUTILS    = "uiutils"
	MODULE_UI         = "userinterface"
	MODULE_MODULETEST = "moduletest"
	MODULE_NETWORK    = "network"
)

func init() {
	logger = logging.MustGetLogger("logger")

	// logging setting with default values
	logging.SetFormatter(logging.MustStringFormatter(defaultFormat))
	logging.SetLevel(defaultLevel, "")
	//defaultOutput := os.Stdout
	//logging.SetBackend(logging.NewLogBackend(defaultOutput, "", 0))
	logger.Info("hfrdlogging initialized!")
}

// A simple wrapper of logging.MustGetLogger
func MustGetLogger(module string) *logging.Logger {
	return logging.MustGetLogger(module)
}

// Set hfrdlogging level of all modules
func SetLogLevel(level logging.Level) {
	logging.SetLevel(level, "")
}

// Todo: Expose logging functions if necessary
