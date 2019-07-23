package packager

import (
	"path"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/pkg/errors"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"os"
	"io/ioutil"
	"github.com/golang/protobuf/proto"
	"hfrd/modules/gosdk/common"
	"fmt"
)


// install chaincode with cds file

func NewCDSPackage(chaincodePath string, goPath string) (*resource.CCPackage, error) {
	if chaincodePath == "" {
		return nil, errors.New("chaincode path must be provided")
	}
	var cdsFile string
	gp := goPath
	if gp == "" {
		gp = defaultGoPath()
		if gp == "" {
			return nil, errors.New("GOPATH not defined")
		}
	}
	cdsFile = path.Join(gp, "src", chaincodePath)
	info, err := os.Stat(cdsFile)
	if os.IsNotExist(err) {
		return nil, errors.Errorf("cds file %s does not exist", cdsFile)
	}
	if info.IsDir() {
		return nil, errors.Errorf("The path %s is a directory. CDS file is required", cdsFile)
	}
	cdsBytes, err := ioutil.ReadFile(cdsFile)
	if err != nil {
		return nil, err
	}
	// unmarshall cds bytes
	cds := &pb.ChaincodeDeploymentSpec{}
	err = proto.Unmarshal(cdsBytes, cds)
	if err != nil {
		return nil, err
	}
	common.Logger.Info(fmt.Sprintf("CDS-Chaincode language: %s, path: %s",
		pb.ChaincodeSpec_Type_name[int32(cds.ChaincodeSpec.Type)], cds.ChaincodeSpec.ChaincodeId.Path))
	ccPkg := &resource.CCPackage{Type: cds.ChaincodeSpec.Type, Code: cds.CodePackage}
	return ccPkg, nil
}