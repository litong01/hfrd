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
	"hfrd/modules/gosdk/common"
)

var channelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update channel",
	Long:  "update channel according to cli parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateChannels()
	},
}

const CHANNEL_UPDATE = "channel.update"

func updateCmd() *cobra.Command {
	flags := channelUpdateCmd.Flags()
	flags.StringVar(&channelNamePrefix, CH_NAME_PREFIX, "", "channe name prefix")
	flags.StringSliceVar(&channelNameList, CH_NAME_LIST, []string{},
		"channel name list, mutual exclusive with "+CH_NAME_PREFIX)
	flags.StringVar(&ordererOrgName, ORDERER_ORG, "", "orderer org  that will be used to update orderer-realated channel configurations")
	flags.StringVar(&ordererName, CH_ORDERER_NAME, "", "orderer name that will be used to update channel")
	flags.StringSliceVar(&peers, CH_PEERS, []string{}, "use which peer during update channel")
	flags.IntVar(&prefixOffset, PREFIX_OFFSET, 0, "prefix offset,used to adjust the start index")

	flags.StringVar(&batchTimeout, CH_BATCH_TIMEOUT, "", "update batch timeout")
	flags.Float64Var(&maxMessageCount, CH_BATCH_SIZE_MAX_MESSAGE_COUNT, 0, "update the max message count in single block")
	flags.Float64Var(&preferredMaxBytes, CH_BATCH_SIZE_PREFERRED_MAX_BATES, 0, "update the max bytes in single block")
	flags.StringVar(&ordererAddressesAction, CH_ORDERER_ADDRESSES_ACTION, "replace", "which action (add/remove/replace) will be executed when update orderer addresses")
	flags.StringSliceVar(&ordererAddresses, CH_ORDERER_ADDRESSES, []string{}, "update the orderer addresses in the channel")
	flags.StringSliceVar(&anchorPeers, ANCHOR_PEERS, []string{}, "")

	channelUpdateCmd.MarkFlagRequired(ORDERER_ORG)
	channelUpdateCmd.MarkFlagRequired(CH_ORDERER_NAME)
	channelUpdateCmd.MarkFlagRequired(CH_PEERS)

	return channelUpdateCmd
}

func updateChannels() error {
	common.Delay(viper.GetString(common.DELAY_TIME))
	connProfile := viper.GetString(common.CONN_PROFILE)
	base := common.NewBase()
	base.ConnectionProfile = connProfile
	base.IterationCount = viper.GetString(common.ITERATION_COUNT)
	base.SetCurrentIter(prefixOffset)
	iterationCount, err := strconv.Atoi(viper.GetString(common.ITERATION_COUNT))
	if err == nil {
		// iterationCount can be converted to integer
		base.IterationCount = strconv.Itoa(iterationCount + prefixOffset)
	}
	base.SetIterationInterval(viper.GetString(common.ITERATION_INTERVAL))
	base.RetryCount = viper.GetInt(common.RETRY_COUNT)

	viperConn, err := common.GetViperInstance(connProfile, "yaml")
	if err != nil {
		return errors.WithMessage(err, "get viper instance from connection profile error")
	}

	if err := viperConn.Unmarshal(&connection); err != nil {
		return errors.WithMessage(err, "unmarshall connection profiles from connection profile error")
	}

	org := viperConn.GetString(CLIENT_ORGANIZATION)
	if org == "" {
		return fmt.Errorf("client.organization is not provided in sdk config yaml")
	}

	c := ChannelConfig{base, channelNamePrefix, nil}

	if channelNamePrefix != "" && len(channelNameList) > 0 {
		return errors.New(CH_NAME_PREFIX + " and " + CH_NAME_LIST + " are mutual exclusive")
	}

	if channelNamePrefix != "" {
		common.Logger.Debug(fmt.Sprintf("CHANNEL_UPDATE channelNamePrefix:%s", channelNamePrefix))
		updateFunc := func(iterationIndex int) error {
			channelName := c.namePrefix + strconv.Itoa(iterationIndex)
			return c.SingleChannelUpdate(org, channelName)
		}
		_, _, err = common.IterateFunc(base, updateFunc, false)
		if err != nil {
			return err
		}
	} else if len(channelNameList) > 0 {
		common.Logger.Debug(fmt.Sprintf("CHANNEL_UPDATE channelNameList:%s", channelNameList))
		common.Logger.Debug(fmt.Sprintf("CHANNEL_UPDATE channelNameList length:%d", len(channelNameList)))

		for index, channelName := range channelNameList {
			c.SingleChannelUpdate(org, channelName)
			if index != len(channelNameList)-1 {
				base.Wait()
			}
		}
	} else {
		return errors.New("Either " + CH_NAME_PREFIX + " or " + CH_NAME_LIST + " should be provided")
	}
	defer c.PrintMetrics(CHANNEL_UPDATE)
	return nil
}

func (c *ChannelConfig) SingleChannelUpdate(org string, channelName string) error {
	var err error
	defer func(now time.Time) {
		if err == nil {
			common.TrackCount(CHANNEL_UPDATE, 1)
			common.TrackTime(now, CHANNEL_UPDATE)
		}
	}(time.Now())

	channelConfig, err := common.GetTempChannelConfigFile(channelName, peers)
	if err != nil {
		return errors.Errorf("Instantiate failed due to errors in GetChannelBackendYaml. %s", err)
	}
	// Initialize sdk with multiple config files
	configBackends, err := common.GetConfigBackends(common.CONFIG_BCCSP, channelConfig, c.Base.ConnectionProfile)
	if err != nil {
		return err
	}
	sdk, err := fabsdk.New(configBackends)
	if err != nil {
		return err
	}
	c.sdk = sdk

	updateOptions := &utilities.UpdateOptions{
		ChannelID:              channelName,
		LedgerClientOrg:        org,
		MspID:                  connection.Organizations[org].Mspid,
		OrdererOrgUpdate:       false,
		OrdererAddressesUpdate: false,
		PeerOrgUpdate:          false,
		OrdererOrgName:         ordererOrgName,
		OrdererName:            ordererName,
		BatchTimeout:           batchTimeout,
		MaxMessageCount:        maxMessageCount,
		OrdererAddressesAction: ordererAddressesAction,
		OrdererAddresses:       ordererAddresses,
		PreferredMaxBytes:       preferredMaxBytes,

		AnchorPeers: anchorPeers,
	}

	// Update channel configurations
	if updateOptions.BatchTimeout != "" || updateOptions.MaxMessageCount != 0 || updateOptions.PreferredMaxBytes != 0 || len(ordererAddresses) != 0 || len(anchorPeers) != 0 {
		err := UpdateChannelConfig(c.sdk, updateOptions)
		if err != nil {
			return err
		}
	} else {
		return errors.Errorf("Please provide at least one meaningful input argument")
	}
	return nil
}

func UpdateChannelConfig(sdk *fabsdk.FabricSDK, updateOptions *utilities.UpdateOptions) error {
	// Initiate ledger client
	channelClientCtx := sdk.ChannelContext(updateOptions.ChannelID, fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(updateOptions.LedgerClientOrg))
	ledgerClient, err := ledger.New(channelClientCtx)
	if err != nil {
		common.Logger.Error(fmt.Sprintf("get ledger client failed due to %s", err))
		return err
	}
	// Query latest channel config
	configBlockBytes, err := utilities.QueryChannelConfigBlock(ledgerClient)
	if err != nil {
		return err
	}
	// Build config update
	configUpdate, updateOptions, err := utilities.CreateUpdateEnvelope(updateOptions, configBlockBytes)
	if err != nil {
		return err
	}
	// Update channel
	if err = utilities.UpdateChannel(sdk, configUpdate, updateOptions); err != nil {
		return err
	}
	// Wait until the channel update finished
	if err = utilities.WaitUntilUpdateSucc(ledgerClient, updateOptions); err != nil {
		return err
	}
	return nil
}
