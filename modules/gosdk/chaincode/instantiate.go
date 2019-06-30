package chaincode

import (
	"encoding/json"
	"fmt"
	"hfrd/modules/gosdk/chaincode/utils"
	"hfrd/modules/gosdk/common"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var chaincodeInstantiateCmd = &cobra.Command{
	Use:              "instantiate",
	Short:            "instantiate chaincode on channel(s)",
	Long:             "instantiate chaincode according the connection profile and parameters provided",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return instantiateChaincode()
	},
}

const CC_INSTANTIATE = "chaincode.instantiate"

func instantiateCmd() *cobra.Command {
	flags := chaincodeInstantiateCmd.Flags()
	flags.StringVar(&channelNamePrefix, CHAN_NAME_PREFIX, "", "channel name prefix")
	flags.IntVar(&prefixOffset, PREFIX_OFFSET, 0, "prefix offset,used to adjust the start index when instantiate chaincode")
	flags.StringSliceVar(&channelNameList, CHAN_NAME_LIST, []string{},
		"channel name list, mutual exclusive with "+CHAN_NAME_PREFIX)
	flags.StringVar(&chaincodeVersion, CC_VERSION, "", "chaincode version")
	flags.StringVar(&path, CC_PATH, "", "chaincode path")
	flags.StringSliceVar(&peers, PEERS, []string{}, "on which peer to instantiate cc")
	flags.StringVar(&chaincodeName, CC_NAME, "", "chaincode name")
	flags.StringVar(&policystr, POLICY_STR, "", "optional cc policy")
	flags.StringVar(&collectionsConfigPath, COLLECTION_CONFIG_PATH, "", "optional collections config file.You need to specify the full qualified path")
	flags.StringVar(&fabricVersion, FABRIC_VERSION, "1.1", "Use fabricVersion to define different capabilities. Use format 1.x in fabricVersion")

	chaincodeInstantiateCmd.MarkFlagRequired(CC_VERSION)
	chaincodeInstantiateCmd.MarkFlagRequired(CC_PATH)
	chaincodeInstantiateCmd.MarkFlagRequired(PEERS)
	chaincodeInstantiateCmd.MarkFlagRequired(CC_NAME)
	return chaincodeInstantiateCmd
}

func instantiateChaincode() error {
	common.Delay(viper.GetString(common.DELAY_TIME))
	connProfile := viper.GetString(common.CONN_PROFILE)
	// Read connection profile
	viperConn, err := common.GetViperInstance(connProfile, "yaml")
	if err != nil {
		return err
	}
	org = viperConn.GetString(CLIENT_ORGANIZATION)
	if org == "" {
		return fmt.Errorf("client.organization is not provided in sdk config yaml")
	}

	base := common.NewBase()
	base.ConnectionProfile = connProfile

	// Load collections config file if specified
	var collectionsPVT []utils.CollectionConfig
	var collsConfig []*cb.CollectionConfig
	if collectionsConfigPath != "" {
		if string(collectionsConfigPath[0]) != "/" {
			// Relative path
			collectionsConfigPath = "/fabric/src/" + collectionsConfigPath
		}
		body, err := ioutil.ReadFile(collectionsConfigPath)
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, &collectionsPVT)
		for _, collection := range collectionsPVT {
			collConfig, err := utils.NewCollectionConfig(collection.Name, collection.Policy, collection.RequiredPeerCount, collection.MaxPeerCount, collection.BlockToLive, collection.MemberOnlyRead)
			if err != nil {
				return err
			}
			collsConfig = append(collsConfig, collConfig)
		}
	}

	// Adjust currIter and iterationCount according to prefixOffset
	base.SetCurrentIter(prefixOffset)
	iterationCount, err := strconv.Atoi(viper.GetString(common.ITERATION_COUNT))
	if err != nil {
		return err
	}
	base.IterationCount = strconv.Itoa(iterationCount + prefixOffset)
	base.SetIterationInterval(viper.GetString(common.ITERATION_INTERVAL))
	base.RetryCount = viper.GetInt(common.RETRY_COUNT)
	cc := &Chaincode{
		Base:              base,
		name:              chaincodeName,
		version:           chaincodeVersion,
		path:              path,
		channelPrefix:     channelNamePrefix,
		CollectionsConfig: collsConfig,
	}

	defer cc.PrintMetrics(CC_INSTANTIATE)

	if channelNamePrefix != "" && len(channelNameList) > 0 {
		return errors.New(CHAN_NAME_PREFIX + " and " + CHAN_NAME_LIST + " are mutual exclusive")
	}


	if channelNamePrefix != ""{
		common.Logger.Debug(fmt.Sprintf("CHAINCODE_INSTANTIATE channelNamePrefix:%s", channelNamePrefix))
		instantiateFunc := func(iterationIndex int) error {
			channelName = cc.channelPrefix+strconv.Itoa(iterationIndex)
			channelConfig, err := common.GetTempChannelConfigFile(channelName, peers)
			if err != nil {
				return errors.Errorf("Instantiate failed due to errors in GetChannelBackendYaml. %s", err)
			}

			// Initialize sdk with multiple config files
			configBackends, err := common.GetConfigBackends(common.CONFIG_BCCSP, channelConfig, connProfile)
			if err != nil {
				return errors.WithMessage(err, "Unable to get config backends")
			}
			// reconfig configBackends with dynamic service discovery
			if fabricVersion != "1.1" {
				configBackends, err = configBackendsWithSD(connProfile, peers[0],
					channelName, org, common.ADMIN, false)
				if err != nil {
					return errors.WithMessage(err, "Unable to get config backends with dynamic service discovery")
				}
			}

			sdk, err := fabsdk.New(configBackends)
			if err != nil {
				return errors.WithMessage(err, "Unable to create sdk instance")
			} else {
				common.Logger.Info("sdk initialized successfully!")
			}
			cc.sdk = sdk
			return cc.InstantiateChaincode(cc.name,
				cc.version, cc.path, org, channelName, peers...)
		}
		_, _, err = common.IterateFunc(base, instantiateFunc, true)
		if err != nil {
			return err
		}

	}else if len(channelNameList) > 0 {
		common.Logger.Debug(fmt.Sprintf("CHAINCODE_INSTANTIATE channelNameList:%s", channelNameList))
		common.Logger.Debug(fmt.Sprintf("CHAINCODE_INSTANTIATE channelNameList length:%d", len(channelNameList)))
		for index, channelName := range channelNameList {
			common.Logger.Debug(fmt.Sprintf("CHAINCODE_INSTANTIATE for channel:%s", channelName))
			channelConfig, err := common.GetTempChannelConfigFile(channelName, peers)
			if err != nil {
				return errors.Errorf("Instantiate failed due to errors in GetChannelBackendYaml. %s", err)
			}
			// Initialize sdk with multiple config files
			configBackends, err := common.GetConfigBackends(common.CONFIG_BCCSP, channelConfig, connProfile)
			if err != nil {
				return errors.WithMessage(err, "Unable to get config backends")
			}
			// reconfig configBackends with dynamic service discovery
			if fabricVersion != "1.1" {
				configBackends, err = configBackendsWithSD(connProfile, peers[0],
					channelName, org, common.ADMIN, false)
				if err != nil {
					return errors.WithMessage(err, "Unable to get config backends with dynamic service discovery")
				}
			}
			sdk, err := fabsdk.New(configBackends)
			if err != nil {
				return errors.WithMessage(err, "Unable to create sdk instance")
			} else {
				common.Logger.Info("sdk initialized successfully!")
			}
			cc.sdk = sdk
			if err = cc.InstantiateChaincode(cc.name,
				cc.version, cc.path, org, channelName, peers...);err != nil{
					return err
			}
			if index != len(channelNameList)-1 {
				base.Wait()
			}
		}

	}else{
		return errors.New("Either " + CHAN_NAME_PREFIX + " or " + CHAN_NAME_LIST + " should be provided")
	}


	return nil
}

func (cc *Chaincode) InstantiateChaincode(name, version, path, org, channel string, peers ...string) error {
	var err error
	defer func(now time.Time) {
		if err == nil {
			common.TrackCount(CC_INSTANTIATE, 1)
			common.TrackTime(now, CC_INSTANTIATE)
		}
	}(time.Now())
	resourceManagerClientContext := cc.sdk.Context(fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(org))
	resMgmtClient, err := resmgmt.New(resourceManagerClientContext)
	if err != nil {
		return err
	}
	// Set up chaincode policy
	var ccPolicy *cb.SignaturePolicyEnvelope
	if policystr != "" {
		common.Logger.Info(fmt.Sprintf("chaincode instantiate: Policy str is %s", policystr))
		//ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP"})
		//ccPolicy, err := cauthdsl.FromString("OR ('Org1MSP.member','Org2MSP.member')")
		//ccPolicy, err := cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
		ccPolicy, err = cauthdsl.FromString(policystr)
		if err != nil {
			return err
		}
		common.Logger.Info(fmt.Sprintf("chaincode instantiate: ccPolicy is %s", ccPolicy))
	} else {
		common.Logger.Info(fmt.Sprintf("policyStr flag not set, will use default endorsement policy"))
		ccPolicy = &cb.SignaturePolicyEnvelope{}
	}
	var instantiateCCReq resmgmt.InstantiateCCRequest
	instantiateCCReq = resmgmt.InstantiateCCRequest{
		Name:       name,
		Path:       path,
		Version:    version,
		Args:       [][]byte{}, // TODO: pass cc instantiate init args as parameter
		Policy:     ccPolicy,
		CollConfig: cc.CollectionsConfig,
	}

	resp, err := resMgmtClient.InstantiateCC(channel, instantiateCCReq,
		resmgmt.WithTargetEndpoints(peers...))
	if err != nil || resp.TransactionID == "" {
		if err != nil && strings.Contains(err.Error(), fmt.Sprintf("chaincode with name '%s' already exists",
			instantiateCCReq.Name)) {
			common.Logger.Info(fmt.Sprintf("chaincode: %s already instantiated on channel: %s", instantiateCCReq.Name, channel))
		} else {
			return err
		}
	}
	return nil
}
