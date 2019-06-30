package jwk

import "hfrd/api/utils/hfrdlogging"

var logger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_JWK)

func NewIbmJwksManager() IbmJwksManager {
	return IbmJwksManager{stopChan: make(chan struct{})}
}
