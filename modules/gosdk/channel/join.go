package channel

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"hfrd/modules/gosdk/common"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var channelJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "join channel",
	Long:  "join channel according to cli parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return joinChannels()
	},
}

const CHANNEL_JOIN = "channel.join"

func joinCmd() *cobra.Command {
	flags := channelJoinCmd.Flags()
	flags.StringVar(&channelNamePrefix, CH_NAME_PREFIX, "",
		"channel name prefix, mutual exclusive with "+CH_NAME_LIST)
	flags.StringSliceVar(&channelNameList, CH_NAME_LIST, []string{},
		"channel name list, mutual exclusive with "+CH_NAME_PREFIX)
	flags.IntVar(&prefixOffset, PREFIX_OFFSET, 0, "prefix offset,used to adjust the start index when join channels")
	flags.StringVar(&ordererName, CH_ORDERER_NAME, "", "orderer name that will used to join channel")
	flags.StringSliceVar(&peers, CH_PEERS, []string{}, "org peers that will join this channel(Currently will join all orgs peers into channel)")

	channelJoinCmd.MarkFlagRequired(CH_ORDERER_NAME)
	channelJoinCmd.MarkFlagRequired(CH_PEERS)

	return channelJoinCmd
}

func joinChannels() error {
	common.Delay(viper.GetString(common.DELAY_TIME))
	connProfile := viper.GetString(common.CONN_PROFILE)
	base := common.NewBase()
	base.ConnectionProfile = connProfile
	// Adjust currIter and iterationCount according to prefixOffset
	base.SetCurrentIter(prefixOffset)
	iterationCount, err := strconv.Atoi(viper.GetString(common.ITERATION_COUNT))
	if err != nil {
		return err
	}
	base.IterationCount = strconv.Itoa(iterationCount + prefixOffset)
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

	// Initialize sdk with multiple config files
	configBackends, err := common.GetConfigBackends(common.CONFIG_BCCSP, connProfile)
	if err != nil {
		return err
	}

	sdk, err := fabsdk.New(configBackends)
	if err != nil {
		return err
	}

	c := ChannelConfig{base, channelNamePrefix, sdk}
	orgClientContext := c.sdk.Context(fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(org))
	orgClient, err := resmgmt.New(orgClientContext)
	if err != nil {
		return err
	}
	if channelNamePrefix != "" && len(channelNameList) > 0 {
		return errors.New(CH_NAME_PREFIX + " and " + CH_NAME_LIST + " are mutual exclusive")
	}
	// channelNamePrefix
	if channelNamePrefix != "" {
		common.Logger.Debug(fmt.Sprintf("CHANNEL_JOIN channelNamePrefix:%s", channelNamePrefix))
		joinFunc := func(iterationIndex int) error {
			channelId := channelNamePrefix + strconv.Itoa(iterationIndex)
			unJoinedList, err := c.FilterJoinedPeers(channelId, connection, orgClient)
			if err != nil {
				return err
			} else if len(unJoinedList) == 0 {
				return nil
			}
			for _, peer := range unJoinedList {
				if err := c.SingleChannelJoin(org, peer, channelId); err != nil {
					return err
				}
			}
			return nil
		}
		_, _, err = common.IterateFunc(base, joinFunc, false)
		if err != nil {
			return err
		}
		// channelNameList
	} else if len(channelNameList) > 0 {
		common.Logger.Debug(fmt.Sprintf("CHANNEL_JOIN channelNameList:%s", channelNameList))
		common.Logger.Debug(fmt.Sprintf("CHANNEL_JOIN channelNameList length:%d", len(channelNameList)))
		for index, channelName := range channelNameList {
			unJoinedList, err := c.FilterJoinedPeers(channelName, connection, orgClient)
			if err != nil {
				return err
			} else if len(unJoinedList) == 0 {
				return nil
			}
			for _, peer := range unJoinedList {
				base.IterationCount = "1" // should always iterate 1 with channelNameList
				base.ResetCurrentIter()
				common.Logger.Debug(fmt.Sprintf("Joining %s to channel %s", org, channelName))
				joinFunc := func(iterationIndex int) error {
					return c.SingleChannelJoin(org, peer, channelName)
				}
				_, _, err = common.IterateFunc(base, joinFunc, false)
				if err != nil {
					return err
				}
			}
			if index != len(channelNameList)-1 {
				base.Wait()
			}
		}
	} else {
		return errors.New("Either " + CH_NAME_PREFIX + " or " + CH_NAME_LIST + " should be provided")
	}

	defer c.PrintMetrics(CHANNEL_JOIN)
	return nil

}

func (c *ChannelConfig) SingleChannelJoin(org, peer string, channelName string) error {
	var err error
	channelID := channelName
	orgClientContext := c.sdk.Context(fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(org))
	orgClient, err := resmgmt.New(orgClientContext)
	if err != nil {
		common.Logger.Error(fmt.Sprintf("Error creating resource manager client: %s", err))
		return err
	}
	defer func(now time.Time) {
		if err == nil {
			common.TrackTime(now, CHANNEL_JOIN)
		}
	}(time.Now())
	// Join peer to channel
	if err = orgClient.JoinChannel(channelID, resmgmt.WithTargetEndpoints(peer), resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(ordererName)); err != nil {
		common.Logger.Error(fmt.Sprintf("failed to join %s peers into channel: %s with error : %s", org, channelID, err))
		return err
	} else {
		common.TrackCount(CHANNEL_JOIN, 1)
		return nil
	}
}

func (c *ChannelConfig) FilterJoinedPeers(channelID string, connection *common.ConnectionProfile, orgClient *resmgmt.Client) ([]string, error) {
	var unJoinedList []string
	for _, peer := range peers {
		joinStatus, err := isJoinedChannel(orgClient, peer, channelID)
		if err != nil {
			return nil, err
		}
		if !joinStatus {
			unJoinedList = append(unJoinedList, peer)
		}
	}
	return unJoinedList, nil
}

func isJoinedChannel(orgClient *resmgmt.Client, peer string, channelID string) (bool, error) {
	joinedChannels, err := orgClient.QueryChannels(resmgmt.WithTargetEndpoints(peer))
	if err != nil {
		common.Logger.Error(fmt.Sprintf("Error query channels with %s for channel: %s: %s", peer, channelID, err))
		return false, err
	}
	for _, channel := range joinedChannels.Channels {
		if channel.ChannelId == channelID {
			common.Logger.Error(fmt.Sprintf("Peer %s already joined channel: %s", peer, channelID))
			return true, nil
		}
	}
	return false, nil
}
