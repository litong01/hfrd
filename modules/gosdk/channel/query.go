package channel

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"hfrd/modules/gosdk/channel/utilities"
	"hfrd/modules/gosdk/channel/utilities/configtxlator"
	"hfrd/modules/gosdk/common"
)

var channelQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "query channel",
	Long:  "query channel according to cli parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return queryChannels()
	},
}

const CHANNEL_QUERY = "channel.query"
const CHANNEL_HEIGHT = "channel.%s.height" // %s represents channel id

func queryCmd() *cobra.Command {
	flags := channelQueryCmd.Flags()
	flags.StringVar(&channelName, CH_NAME, "", "channel name")
	flags.StringSliceVar(&peers, CH_PEERS, []string{}, "use which peer to query channel")

	channelQueryCmd.MarkFlagRequired(CH_NAME)
	channelQueryCmd.MarkFlagRequired(CH_PEERS)

	return channelQueryCmd
}

func queryChannels() error {
	common.Delay(viper.GetString(common.DELAY_TIME))
	connProfile := viper.GetString(common.CONN_PROFILE)
	base := common.NewBase()
	base.ConnectionProfile = connProfile
	base.IterationCount = viper.GetString(common.ITERATION_COUNT)
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

	c := ChannelConfig{base, channelName, nil}

	queryFunc := func(iterationIndex int) error {
		return c.SingleChannelQuery(org, iterationIndex)
	}
	defer c.PrintMetrics(CHANNEL_QUERY)
	_, _, err = common.IterateFunc(base, queryFunc, false)
	if err != nil {
		return err
	}
	return nil
}

func (c *ChannelConfig) SingleChannelQuery(org string, iterationIndex int) error {
	var err error
	channelID := c.namePrefix
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
	channelClientCtx := c.sdk.ChannelContext(channelID, fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(org))
	ledgerClient, err := ledger.New(channelClientCtx)
	if err != nil {
		common.Logger.Error(fmt.Sprintf("get ledger client failed due to %s", err))
		return err
	}
	if ledgerInfo, err := ledgerClient.QueryInfo(); err == nil {
		height := int64(ledgerInfo.BCI.Height)
		if height > 0 {
			common.TrackGauge(fmt.Sprintf(CHANNEL_HEIGHT, channelID), height)
			common.Logger.Info(fmt.Sprintf("Channel:%s height %v", channelID, height))
		} else {
			common.Logger.Error(fmt.Sprintf("Channel:%s height %v could not be tracked", channelID,
				ledgerInfo.BCI.Height))
		}
	} else {
		return err
	}
	defer func(now time.Time) {
		// only track time when no error
		if err == nil {
			common.TrackCount(CHANNEL_QUERY, 1)
			common.TrackTime(now, CHANNEL_QUERY)
			common.Logger.Debug(fmt.Sprintf("Tracking CHANNEL_QUERY %s", time.Since(now)))
		} else {
			common.Logger.Info(fmt.Sprintf("Skipping tracking CHANNEL_QUERY: %s", err))
		}
	}(time.Now())
	configBlockBytes, err := utilities.QueryChannelConfigBlock(ledgerClient)
	if err != nil {
		return err
	}

	channelGroup, err := configtxlator.DecodeProto("common.Block", configBlockBytes)
	if err != nil {
		return errors.Errorf("Error in decode proto : %s", err)
	}
	batchTimeout, maxMessageCount, preferredMaxBytes, orgAnchorPeers, ordererAddresses := utilities.GetConfigFromChannelGroup(channelGroup, connection.Organizations[org].Mspid)
	common.Logger.Info(fmt.Sprintf("Query %s config: BatchTimeout:%s, BatchSize.MaxMessageCount:%f, BatchSize.PreferredMaxBytes:%f", c.namePrefix, batchTimeout, maxMessageCount, preferredMaxBytes))
	common.Logger.Info(fmt.Sprintf("OrdererAddresses: %s", ordererAddresses))
	common.Logger.Info(fmt.Sprintf(" Org:%s AnchorPeers: %#v", org, orgAnchorPeers))
	utilities.PrintMSPidFromChannelGroup(channelGroup)
	return nil
}
