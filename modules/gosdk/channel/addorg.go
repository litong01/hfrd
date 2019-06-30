package channel

import (
	"fmt"
	"time"

	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"hfrd/modules/gosdk/channel/utilities"
	"hfrd/modules/gosdk/channel/utilities/configtxlator"
	"hfrd/modules/gosdk/common"
)

var channelAddOrgCmd = &cobra.Command{
	Use:   "addorg",
	Short: "add org to channel",
	Long:  "add new org to exsiting channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		return addNewOrgToChannel()
	},
}

func addorgCmd() *cobra.Command {
	flags := channelAddOrgCmd.Flags()
	flags.StringVar(&channelNamePrefix, CH_NAME_PREFIX, "", "channel name prefix")
	flags.StringVar(&ordererOrgName, ORDERER_ORG, "", "orderer org  that will be used to update orderer-realated channel configurations")
	flags.StringVar(&ordererName, CH_ORDERER_NAME, "", "orderer name that will be used to update channel")
	flags.StringSliceVar(&peers, CH_PEERS, []string{}, "use which peer during update channel")
	flags.StringVar(&newOrgConfigPath, ORG_CONFIG_PATH, "", "new org config file path")

	channelAddOrgCmd.MarkFlagRequired(CH_NAME_PREFIX)
	channelAddOrgCmd.MarkFlagRequired(ORDERER_ORG)
	channelAddOrgCmd.MarkFlagRequired(CH_ORDERER_NAME)
	channelAddOrgCmd.MarkFlagRequired(CH_PEERS)
	channelAddOrgCmd.MarkFlagRequired(ORG_CONFIG_PATH)

	return channelAddOrgCmd
}

func addNewOrgToChannel() error {
	connProfile := viper.GetString(common.CONN_PROFILE)
	base := common.NewBase()
	base.ConnectionProfile = connProfile
	base.IterationCount = viper.GetString(common.ITERATION_COUNT)
	base.SetIterationInterval(viper.GetString(common.ITERATION_INTERVAL))
	base.RetryCount = viper.GetInt(common.RETRY_COUNT)

	viperConn, err := common.GetViperInstance(connProfile, "yaml")
	if err != nil {
		errors.WithMessage(err, "get viper instance from connection profile error")
	}

	if err := viperConn.Unmarshal(&connection); err != nil {
		errors.WithMessage(err, "unmarshall connection profiles from connection profile error")
	}

	org := viperConn.GetString(CLIENT_ORGANIZATION)
	if org == "" {
		return fmt.Errorf("client.organization is not provided in sdk config yaml")
	}

	c := ChannelConfig{base, channelNamePrefix, nil}

	orgList := viperConn.GetStringMap("organizations")
	orgConfigFilePath := newOrgConfigPath
	addOrgFunc := func(iterationIndex int) error {
		return c.ChannelAddOrg(org, orgList, ordererName, orgConfigFilePath, iterationIndex)
	}
	_, _, err = common.IterateFunc(base, addOrgFunc, false)
	if err != nil {
		return err
	}
	return nil
}

//ChannelAddOrg add a new org to channel
func (c *ChannelConfig) ChannelAddOrg(org string, orgListMap map[string]interface{}, orderer string, newOrgConfigFile string, iterationIndex int) error {
	defer common.TrackTime(time.Now(), "channel.addorg")

	channelID := c.namePrefix + strconv.Itoa(iterationIndex)
	channelConfig, err := common.GetTempChannelConfigFile(channelID, peers)
	if err != nil {
		return errors.Errorf("Instantiate failed due to errors in GetChannelBackendYaml. %s", err)
	}
	// Initialize sdk with multiple config files
	configBackends, err := common.GetConfigBackends(common.CONFIG_BCCSP, channelConfig, c.Base.ConnectionProfile)
	if err != nil {
		return nil
	}
	sdk, err := fabsdk.New(configBackends)
	if err != nil {
		return err
	}
	c.sdk = sdk

	// Initiate ledger client
	channelClientCtx := c.sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg(org))
	ledgerClient, err := ledger.New(channelClientCtx)
	if err != nil {
		common.Logger.Error(fmt.Sprintf("get ledger client failed due to %s", err))
	}

	configBlockBytes, err := utilities.QueryChannelConfigBlock(ledgerClient)
	if err != nil {
		return err
	}

	channelGroup, err := configtxlator.DecodeProto("common.Block", configBlockBytes)
	if err != nil {
		return errors.Errorf("Error in decode proto : %s", err)
	}
	//bypass way to decode 2 original channel group
	channelGroupOrig, err := configtxlator.DecodeProto("common.Block", configBlockBytes)
	if err != nil {
		return errors.Errorf("Error in decode proto : %s", err)
	}

	filename := newOrgConfigFile
	// bytes, err := ioutil.ReadFile(filename)
	// if err != nil {
	//     fmt.Println("ReadFile: ", err.Error())
	//     return err
	// }

	// if err := json.Unmarshal(bytes, &xxx); err != nil {
	//     fmt.Println("Unmarshal: ", err.Error())
	//     return err
	// }
	neworgChannelGroup, err := configtxlator.JSONtoConfig(filename)
	if err != nil {
		return errors.Errorf("Error in JSONtoConfig : %s", err)
	}

	channelGroup["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Application"].(map[string]interface{})["groups"].(map[string]interface{})["Org3MSP"] = neworgChannelGroup

	envByte, err := utilities.CreateNewOrgEnvelope(channelGroupOrig, channelGroup, channelID)
	if err != nil {
		return errors.Errorf("CreateNewOrgEnvelope: %s", err)
	}

	appGroup := channelGroupOrig["channel_group"].(map[string]interface{})["groups"].(map[string]interface{})["Application"].(map[string]interface{})["groups"].(map[string]interface{})
	var orgList []string
	for key, value := range orgListMap {
		_, ok := appGroup[value.(map[string]interface{})["mspid"].(string)]
		if ok {
			orgList = append(orgList, key)
		}
		common.Logger.Info(fmt.Sprintf("Total orgnizations in connection profile: %s", key))
	}
	common.Logger.Info(fmt.Sprintf("Orgnizations in current channel: %s", orgList))
	//collect signatures for peer orgs in current channel
	signIdentities, err := utilities.CollectSign(sdk, orgList)
	if err != nil {
		return errors.Errorf("CollectSign: %s", err)
	}
	err = utilities.AddOrgToChannel(sdk, channelID, envByte, signIdentities, org, orderer)
	if err != nil {
		return errors.Errorf("AddOrgToChannel transaction: %s", err)
	}

	return nil
}
