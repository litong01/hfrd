package chaincode

import (
	"fmt"
	"hfrd/modules/gosdk/common"
	"os"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"hfrd/modules/gosdk/chaincode/packager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"path/filepath"
)

var chaincodeInstallCmd = &cobra.Command{
	Use:              "install",
	Short:            "install chaincode on peer(s)",
	Long:             "install chaincode on peer(s) as per parameters provided",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installChaincode()
	},
}

const CC_INSTALL = "chaincode.install"

func installCmd() *cobra.Command {
	flags := chaincodeInstallCmd.Flags()
	flags.StringVar(&chaincodeNamePrefix, CC_NAME_PREFIX, "", "chaincode name prefix")
	flags.IntVar(&prefixOffset, PREFIX_OFFSET, 0, "prefix offset,used to adjust the start index when install chaincode")
	flags.StringVar(&chaincodeVersion, CC_VERSION, "", "chaincode version")
	flags.StringVar(&path, CC_PATH, "", "chaincode path")
	flags.StringSliceVar(&peers, PEERS, []string{}, "on which peer to install cc")
	flags.StringVarP(&lang, "lang", "l", "golang", "chaincode language(golang, node)")

	chaincodeInstallCmd.MarkFlagRequired(CC_NAME_PREFIX)
	chaincodeInstallCmd.MarkFlagRequired(CC_VERSION)
	chaincodeInstallCmd.MarkFlagRequired(CC_PATH)
	chaincodeInstallCmd.MarkFlagRequired(PEERS)
	return chaincodeInstallCmd
}

func installChaincode() error {
	common.Delay(viper.GetString(common.DELAY_TIME))
	connProfile := viper.GetString(common.CONN_PROFILE)
	base := common.NewBase()
	base.ConnectionProfile = connProfile
	base.SetIterationInterval(viper.GetString(common.ITERATION_INTERVAL))
	base.RetryCount = viper.GetInt(common.RETRY_COUNT)

	// Read connection profile
	viperConn, err := common.GetViperInstance(connProfile, "yaml")
	if err != nil {
		return err
	}
	if err := viperConn.Unmarshal(&connection); err != nil {
		return errors.WithMessage(err, "unmarshall connection profiles from connection profile error")
	}

	var organizations common.Organizations
	if err := viperConn.Unmarshal(&organizations); err != nil {
		return errors.WithMessage(err, "unmarshall organizations from connection profile error")
	}

	// Initialize sdk with multiple config files
	configBackends, err := common.GetConfigBackends(common.CONFIG_BCCSP, connProfile)
	if err != nil {
		return err
	}
	sdk, err := fabsdk.New(configBackends)
	if err != nil {
		return err
	} else {
		common.Logger.Info("sdk initialized successfully!")
	}

	cc := &Chaincode{
		Base:       base,
		namePrefix: chaincodeNamePrefix,
		version:    chaincodeVersion,
		path:       path,
		sdk:        sdk,
	}

	defer cc.PrintMetrics(CC_INSTALL)
	// organizations loops
	for orgName, org := range organizations.Organizations {
		// peers loop
		for _, peer := range org.Peers {
			// filter the peers provided in cli params
			for _, peer1 := range peers {
				if peer == peer1 {
					// Adjust currIter and iterationCount according to prefixOffset
					base.SetCurrentIter(prefixOffset)
					iterationCount, err := strconv.Atoi(viper.GetString(common.ITERATION_COUNT))
					if err != nil {
						return err
					}
					base.IterationCount = strconv.Itoa(iterationCount + prefixOffset)
					installFunc := func(iterationIndex int) error {
						return cc.InstallChaincode(cc.namePrefix+strconv.Itoa(iterationIndex),
							cc.version, cc.path, peer, orgName)
					}
					_, _, err = common.IterateFunc(base, installFunc, false)
					if err != nil {
						return err
					}
					cc.ResetCurrentIter()
				}
			}
		}
	}
	return nil
}

// name: chaincode name
// version: chaincode version
// path: chaincode path: relative to GOPATH environment variable
// peer: on which peer to install the chaincode
// org: the Org MSP id to which the peer belongs to
func (cc *Chaincode) InstallChaincode(name, version, path, peer, org string) error {
	var err error
	defer func(now time.Time) {
		if err == nil {
			common.TrackCount(CC_INSTALL, 1)
			common.TrackTime(now, CC_INSTALL)
		}
	}(time.Now())
	var ccPkg *resource.CCPackage
	if (lang == "node") {
		ccPkg, err = packager.NewCCPackage(path, os.Getenv("GOPATH"))
	} else if (lang == "cds") {
		ccPkg, err = packager.NewCDSPackage(path, os.Getenv("GOPATH"))
		path = filepath.Dir(path)
	} else {  // currently we can support node, golang and cds. The default would be golang
		ccPkg, err = gopackager.NewCCPackage(path, os.Getenv("GOPATH"))
	}

	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("Error packaging cc from path %s",
			os.Getenv("GOPATH")+"/"+path))
	}
	installCCReq := resmgmt.InstallCCRequest{
		Name:    name,
		Path:    path,
		Version: version,
		Package: ccPkg,
	}
	resourceManagerClientContext := cc.sdk.Context(fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(org))
	resMgmtClient, err := resmgmt.New(resourceManagerClientContext)
	if err != nil {
		return err
	}
	_, err = resMgmtClient.InstallCC(installCCReq, resmgmt.WithTargetEndpoints(peer))
	if err != nil {
		return err
	}
	return nil
}
