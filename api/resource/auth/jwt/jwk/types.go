package jwk

import (
	"sync"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"encoding/base64"
	"crypto/rsa"
	"math/big"
	"crypto"
	"errors"
)

const (
	ibm_jwks_url = "https://iam.ng.bluemix.net/identity/keys"
	update_interval = 24 * time.Hour
)

var (
	KID_NOT_FOUND = errors.New("Key not found to verify the token")
	EXPONENT_NOT_SUPPORTED = errors.New("The exponent for key is not supported yet")
)

type IbmJWK struct {
	Kty string `json:"kty"`
	N string `json:"n"`
	E string `json:"e"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type IbmJWKS struct {
	Keys []IbmJWK `json:"keys"`
}

type IbmJwksManager struct {
	IbmJWKS
	stopChan chan struct{}
	once sync.Once
	sync.RWMutex
}

func (manager *IbmJwksManager) Init() {
	manager.once.Do(func() {
		manager.updateJWKS()
		// Periodically update JWKS from IBM Cloud Platform
		ticker := time.NewTicker(update_interval)
		go func() {
			for {
				select {
				case <- ticker.C:
					manager.updateJWKS()
				case <- manager.stopChan:
					logger.Info("Stop updating JWKS from IBM Cloud")
					ticker.Stop()
				}
			}
		}()
	})
}

func (manager *IbmJwksManager) Stop() {
	logger.Info("Stopping IBM JWKS manager")
	manager.once.Do(func() {
		close(manager.stopChan)
	})
}

// kid: key identifier; content: hash input string;
// signature: the signature of the hash of content
// Return error if signature verification fails
func (manager *IbmJwksManager) VerifySig(kid, content, signature string) error {
	manager.RLock()
	var jwk IbmJWK
	for _, item := range manager.Keys {
		if kid == item.Kid {
			jwk = item
		}
	}
	if jwk.Kid == "" {
		return KID_NOT_FOUND
	}
	// decode the base64 bytes for n
	nb, _ := base64.RawURLEncoding.DecodeString(jwk.N)
	e := 0
	// The default exponent is usually 65537, so just compare the
	// base64 for [1,0,1] or [0,1,0,1]
	if jwk.E == "AQAB" || jwk.E == "AAEAAQ" {
		e = 65537
	} else {
		return EXPONENT_NOT_SUPPORTED
	}
	manager.RUnlock()
	pk := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: e,
	}
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(content))
	sig, _ := base64.RawURLEncoding.DecodeString(signature)
	if err := rsa.VerifyPKCS1v15(pk, crypto.SHA256, hasher.Sum(nil),
		sig); err != nil {
		return err
	}
	return nil
}

func (manager *IbmJwksManager) updateJWKS() {
	resp, err := http.Get(ibm_jwks_url)
	if err != nil {
		logger.Warningf("Unable to get response from %s. Error: %s",
			ibm_jwks_url, err)
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warningf("Unable to get response from %s. Error: %s",
			ibm_jwks_url, err)
	}
	var ibmJWKS IbmJWKS
	if err := json.Unmarshal(data, &ibmJWKS); err != nil {
		logger.Warningf("Unable to unmarshall JWKS: %s", err)
	}
	if len(ibmJWKS.Keys) > 0 {
		manager.Lock()
		manager.Keys = ibmJWKS.Keys
		manager.Unlock()
		logger.Debugf("Successfully updated IBM JWKS which contains %d keys",
			len(manager.IbmJWKS.Keys))
	} else {
		logger.Warningf("No key found from %s", ibm_jwks_url)
	}
}
