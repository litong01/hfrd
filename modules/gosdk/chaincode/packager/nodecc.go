package packager


import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"go/build"
	"path"
	"path/filepath"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/pkg/errors"
	"fmt"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)


// NewCCPackage creates new node chaincode package
func NewCCPackage(chaincodePath string, goPath string) (*resource.CCPackage, error) {

	if chaincodePath == "" {
		return nil, errors.New("chaincode path must be provided")
	}

	var projDir string
	gp := goPath
	if gp == "" {
		gp = defaultGoPath()
		if gp == "" {
			return nil, errors.New("GOPATH not defined")
		}
	}
	projDir = path.Join(gp, "src", chaincodePath)
	tarBytes, err := GetDeploymentPayload(projDir)
	if err != nil {
		return nil, err
	}
	ccPkg := &resource.CCPackage{Type: pb.ChaincodeSpec_NODE, Code: tarBytes}
	return ccPkg, nil
}

func GetDeploymentPayload(path string) ([]byte, error) {
	var err error
	// --------------------------------------------------------------------------------------
	// Write out our tar package
	// --------------------------------------------------------------------------------------
	payload := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(payload)
	tw := tar.NewWriter(gw)

	folder := path
	if folder == "" {
		return nil, errors.New("ChaincodeSpec's path cannot be empty")
	}

	// trim trailing slash if it exists
	if folder[len(folder)-1] == '/' {
		folder = folder[:len(folder)-1]
	}

	if err = WriteFolderToTarPackage(tw, folder, []string{"node_modules"}, nil, nil); err != nil {
		return nil, fmt.Errorf("Error writing Chaincode package contents: %s", err)
	}

	// Write the tar file out
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("Error writing Chaincode package contents: %s", err)
	}

	tw.Close()
	gw.Close()

	return payload.Bytes(), nil
}

// defaultGoPath returns the system's default GOPATH. If the system
// has multiple GOPATHs then the first is used.
func defaultGoPath() string {
	gpDefault := build.Default.GOPATH
	gps := filepath.SplitList(gpDefault)

	return gps[0]
}
