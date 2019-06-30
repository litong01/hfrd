package common

import (
	"encoding/json"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

var rawJSON = []byte(`{
	"level": "error",
	"encoding": "json",
	"outputPaths": ["stdout"],
	"errorOutputPaths": ["stderr"],
	"initialFields": {},
	"encoderConfig": {
	  "messageKey": "msg",
	  "levelKey": "lvl",
	  "levelEncoder": "uppercase"
	}
      }`)

func InitLog(loggingLevel string) {
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	switch strings.ToUpper(loggingLevel) {
	case "DEBUG":
		cfg.Level.SetLevel(zap.DebugLevel)
	case "INFO":
		cfg.Level.SetLevel(zap.InfoLevel)
	case "ERROR":
		cfg.Level.SetLevel(zap.ErrorLevel)
	default:
		cfg.Level.SetLevel(zap.ErrorLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	Logger = logger
}
