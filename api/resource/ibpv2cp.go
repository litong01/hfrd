package resource

import (
	"archive/tar"
	"compress/gzip"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/sw"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//  SUPPORT IBP V2.0 SaaS connection profile json and identity json

type ibp2Channel struct {
	Chaincodes []string
	Orderers   []string
	Peers      map[string]interface{}
}
type ibp2Client struct {
	Cryptoconfig map[string]string
	Organization string
}
type ibp2Order struct {
	TlsCACerts map[string]string `yaml:"tlsCACerts"`
	Url        string
}
type ibp2Organization struct {
	CryptoPath             string `yaml:"cryptoPath"`
	MspId                  string
	Peers                  []string
	CertificateAuthorities []string `yaml:"certificateAuthorities"`
}
type ibp2Peer struct {
	EventUrl    string            `yaml:"eventUrl"`
	TlsCACerts  map[string]string `yaml:"tlsCACerts"`
	Url         string
	GrpcOptions map[string]string `yaml:"grpcOptions"`
}
type ibp2CA struct {
	Url            string
	CaName         string              `yaml:"caName"`
	TlsCACerts     map[string]string   `yaml:"-"`
	TlsCACertsYaml map[string][]string `yaml:"tlsCACerts"` // THIS IS LIMITED BY fabric-sdk-go
}
type ibp2Conn struct {
	Channels               map[string]ibp2Channel
	Client                 ibp2Client
	Description            string
	Name                   string
	Orderers               map[string]ibp2Order
	Organizations          map[string]ibp2Organization
	Peers                  map[string]ibp2Peer
	Version                string
	CertificateAuthorities map[string]ibp2CA `yaml:"certificateAuthorities"`
}

type ibp2Id struct {
	Name        string
	Private_key string
	Cert        string
}

// Input
// path: the connection-profiles directory path
// will return error if any
// if no error, will generate a tar gz file named as tarGzName in the parent directory of "path"
// E.g. If you pass ("./connection-profiles", "ibpcerts.tar.gz") as input params and no error occurred,
//	then convertIBP2SaaS will generate "./ibpcerts.tar.gz" that could feed hfrd test modules
func convertIBP2SaaS(path, tarGzName string) error {
	// TODO: assume connection.json and identity.json exist in each org's directory
	var dirs []os.FileInfo
	var err error
	if dirs, err = ioutil.ReadDir(path); err != nil {
		return err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		mspId := dir.Name()
		cpp := path + SEP + mspId + SEP + CONN_JSON
		idp := path + SEP + mspId + SEP + IDENTITY
		cpJsonBytes, err := ioutil.ReadFile(cpp)
		if err != nil {
			return errors.WithMessage(err, "Read error:"+cpp)
		}
		var cpJson ibp2Conn
		if err = json.Unmarshal(cpJsonBytes, &cpJson); err != nil {
			return errors.WithMessage(err, "Unmarshal error:"+cpp)
		}
		// TODO: connection profile validation
		for index, _ := range cpJson.CertificateAuthorities {
			if pem := cpJson.CertificateAuthorities[index].TlsCACerts["pem"]; pem != "" {
				cpJson.CertificateAuthorities[index].TlsCACerts["pem"] = strings.Replace(pem, "\r\n", "\n", -1)

			}
		}

		// TODO: required by hfrd test modules
		if cpJson.Client.Cryptoconfig == nil {
			cpJson.Client.Cryptoconfig = make(map[string]string)
		}
		cpJson.Client.Cryptoconfig["path"] = "/fabric/keyfiles"
		for k, _ := range cpJson.Organizations {
			org := cpJson.Organizations[k]
			org.CryptoPath = fmt.Sprintf("%s/users/{username}@%s/msp", k, k)
			cpJson.Organizations[k] = org
		}
		channel := ibp2Channel{Peers: make(map[string]interface{}), Orderers: make([]string, 0)}
		for peerName, _ := range cpJson.Peers {
			channel.Peers[peerName] = make(map[string]interface{})
		}
		for ordererName, _ := range cpJson.Orderers {
			channel.Orderers = append(channel.Orderers, ordererName)
		}

		if cpJson.Channels == nil {
			cpJson.Channels = make(map[string]ibp2Channel)
		}
		cpJson.Channels[cpJson.Name] = channel

		idb, err := ioutil.ReadFile(idp)
		if err != nil {
			return errors.WithMessage(err, "Read error"+idp)
		}
		var id ibp2Id
		// TODO: identity json validation
		if err = json.Unmarshal(idb, &id); err != nil {
			return errors.WithMessage(err, "Unmarshal error:"+idp)
		}
		keyFilesDirPath := path + SEP + ".." + SEP + KEYFILES
		orgPath := keyFilesDirPath + SEP + mspId
		if err := os.MkdirAll(orgPath, os.ModePerm); err != nil {
			return err
		}
		// convert certificateAuthorities
		cpYaml := cpJson
		for k, v := range cpYaml.CertificateAuthorities {
			v.TlsCACertsYaml = make(map[string][]string)
			v.TlsCACertsYaml["pem"] = []string{v.TlsCACerts["pem"]}
			cpYaml.CertificateAuthorities[k] = v
		}

		cpYamlBytes, err := yaml.Marshal(cpYaml)
		// write connection profile after conversion
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(orgPath+SEP+CONN_YAML, cpYamlBytes, os.ModePerm); err != nil {
			return err
		}
		if err = ioutil.WriteFile(orgPath+SEP+CONN_JSON, cpJsonBytes, os.ModePerm); err != nil {
			return err
		}
		// make user directory
		userMspDir := orgPath + SEP + "users" + SEP + "Admin@%s" + SEP + "msp"
		keystoreDir := fmt.Sprintf(userMspDir+SEP+"keystore", mspId)
		signcertDir := fmt.Sprintf(userMspDir+SEP+"signcerts", mspId)
		if err = os.MkdirAll(keystoreDir, os.ModePerm); err != nil {
			return err
		}
		if err = os.MkdirAll(signcertDir, os.ModePerm); err != nil {
			return err
		}
		userKeyBase64 := id.Private_key
		userCertBase64 := id.Cert

		var ski []byte
		// write user sign cert file
		if userCert, err := base64.StdEncoding.DecodeString(userCertBase64); err != nil {
			return errors.WithMessage(err, "Decode user cert error!")
		} else {
			if err = ioutil.WriteFile(fmt.Sprintf(signcertDir+SEP+"Admin@%s-cert.pem", mspId),
				userCert, os.ModePerm); err != nil {
				return err
			}
			ski, err = getSKIFromX509PemCert(userCert)
			if err != nil {
				return err
			}
		}
		// write user private key
		if userKey, err := base64.StdEncoding.DecodeString(userKeyBase64); err != nil {
			return errors.WithMessage(err, "Decode user key error!")
		} else {
			if err = ioutil.WriteFile(keystoreDir+SEP+hex.EncodeToString(ski)+"_sk", userKey, os.ModePerm); err != nil {
				return err
			}
		}
	}
	if err = genTarGz(filepath.Join(path, "..", KEYFILES), tarGzName); err != nil {
		return err
	}
	return nil
}

func getSKIFromX509PemCert(cert []byte) ([]byte, error) {
	block, _ := pem.Decode(cert)
	if block == nil {
		return cert, errors.New("Unable to get block bytes from cert")
	}
	crt, err := x509.ParseCertificate(block.Bytes)
	memoryKeyStore := sw.NewInMemoryKeyStore()
	csp, err := sw.NewWithParams(256, "SHA2", memoryKeyStore)
	if err != nil {
		return cert, err
	}
	opts := bccsp.X509PublicKeyImportOpts{Temporary: true}
	k, err := csp.KeyImport(crt, &opts)
	if err != nil {
		return cert, err
	}
	return k.SKI(), nil
}

func genTarGz(inputPath, tarGzPath string) error {
	tarGzFile, err := os.Create(tarGzPath)
	if err != nil {
		return err
	}
	defer tarGzFile.Close()
	gzWriter := gzip.NewWriter(tarGzFile)
	defer gzWriter.Close()
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Errorf("Error warlking path(%s): %s\n", path, err)
			return err
		}
		// bypass directory
		if info.IsDir() {
			return nil
		}
		var pathInTar string
		if filepath.Dir(inputPath) != "." {
			pathInTar = path[len(filepath.Dir(inputPath))+1:]
		} else {
			pathInTar = path
		}
		if len(pathInTar) == 0 {
			return nil
		}
		if fr, err := os.Open(path); err != nil {
			return err
		} else {
			defer fr.Close()
			if fih, err := tar.FileInfoHeader(info, pathInTar); err != nil {
				return err
			} else {
				fih.Name = pathInTar
				if err = tarWriter.WriteHeader(fih); err != nil {
					return err
				}
			}
			if _, err := io.Copy(tarWriter, fr); err != nil {
				return err
			}
		}
		return nil
	}
	if err := filepath.Walk(inputPath, walkFunc); err != nil {
		return err
	}
	return nil
}
