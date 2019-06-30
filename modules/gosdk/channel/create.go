package channel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"hfrd/modules/gosdk/channel/utilities"
	"hfrd/modules/gosdk/common"

	"github.com/pkg/errors"
)

var channelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create channel",
	Long:  "create channel according to cli parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return createChannels()
	},
}

const CHANNEL_CREATE = "channel.create"

func createCmd() *cobra.Command {
	flags := channelCreateCmd.Flags()
	flags.StringVar(&channelNamePrefix, CH_NAME_PREFIX, "", "channel name prefix")
	flags.StringSliceVar(&channelNameList, CH_NAME_LIST, []string{},
		"channel name list, mutual exclusive with "+CH_NAME_PREFIX)
	flags.IntVar(&prefixOffset, PREFIX_OFFSET, 0, "prefix offset,used to adjust the start index when create channels")
	flags.StringVar(&channelConsortium, CH_CONSORTIUM, "", "channel consotium that used to create channel")
	flags.StringSliceVar(&channelOrgs, CH_PEER_ORGS, []string{}, "peer orgs that can join this channel")
	flags.StringVar(&ordererName, CH_ORDERER_NAME, "", "orderer name that will be used to create channel")
	flags.StringVar(&applicationCapability, APPLICATION_CAPABILITY, "V1_1", "Use applicationCapability to enable private data in application channels")

	channelCreateCmd.MarkFlagRequired(CH_CONSORTIUM)
	channelCreateCmd.MarkFlagRequired(CH_PEER_ORGS)
	channelCreateCmd.MarkFlagRequired(CH_ORDERER_NAME)

	return channelCreateCmd
}

func createChannels() error {
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

	// Initialize the connection profile
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
	if channelNamePrefix != "" && len(channelNameList) > 0 {
		return errors.New(CH_NAME_PREFIX + " and " + CH_NAME_LIST + " are mutual exclusive")
	}
	// channelNamePrefix
	if channelNamePrefix != "" {
		common.Logger.Debug(fmt.Sprintf("CHANNEL_CREATE channelNamePrefix:%s", channelNamePrefix))
		createFunc := func(iterationIndex int) error {
			channelName := c.namePrefix + strconv.Itoa(iterationIndex)
			return c.SingleChannelCreate(org, channelName)
		}
		_, _, err = common.IterateFunc(base, createFunc, false)
		if err != nil {
			return err
		}

	} else if len(channelNameList) > 0 {
		common.Logger.Debug(fmt.Sprintf("CHANNEL_CREATE channelNameList:%s", channelNameList))
		common.Logger.Debug(fmt.Sprintf("CHANNEL_CREATE channelNameList length:%d", len(channelNameList)))
		for index, channelName := range channelNameList {
			c.SingleChannelCreate(org, channelName)
			if index != len(channelNameList)-1 {
				base.Wait()
			}
		}
	} else {
		return errors.New("Either " + CH_NAME_PREFIX + " or " + CH_NAME_LIST + " should be provided")
	}

	defer c.PrintMetrics(CHANNEL_CREATE)
	return nil

}

func (c *ChannelConfig) SingleChannelCreate(org string, channelName string) error {
	var err error
	defer func(now time.Time) {
		if err == nil {
			common.TrackCount(CHANNEL_CREATE, 1)
			common.TrackTime(now, CHANNEL_CREATE)
		}
	}(time.Now())

	mspClient, err := mspclient.New(c.sdk.Context(), mspclient.WithOrg(org))
	if err != nil {
		common.Logger.Error(fmt.Sprintf("Error creating msp client: %s", err))
		return err
	}

	adminIdentity, err := mspClient.GetSigningIdentity(common.ADMIN)
	if err != nil {
		common.Logger.Error(fmt.Sprintf("failed to get admin signing identity: %s", err))
		return err
	}

	orgClientContext := c.sdk.Context(fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(org))
	orgClient, err := resmgmt.New(orgClientContext)
	if err != nil {
		common.Logger.Error(fmt.Sprintf("Error creating resource manager client: %s", err))
		return err
	}

	// Build channel creation request and send create proposal to ordering system
	var peerOrgMSPs []string
	for _, orgName := range channelOrgs {
		mspID := connection.Organizations[orgName].Mspid
		peerOrgMSPs = append(peerOrgMSPs, mspID)
	}
	chTxBytes, err := utilities.CreatChannelTxEnvelope(applicationCapability, channelName, channelConsortium, peerOrgMSPs...)
	if err != nil {
		return err
	}
	req := resmgmt.SaveChannelRequest{ChannelID: channelName, ChannelConfig: bytes.NewReader(chTxBytes),
		SigningIdentities: []msp.SigningIdentity{adminIdentity}}

	res, err := orgClient.SaveChannel(req, resmgmt.WithOrdererEndpoint(ordererName))
	if err != nil || res.TransactionID == "" {
		if strings.Contains(err.Error(), "error validating ReadSet") {
			common.Logger.Error(fmt.Sprintf("The channel:%s might already exist, skip the error: %s", channelName, err))
			return err
		} else {
			common.Logger.Error(fmt.Sprintf("failed to create channel with name %s: %s", channelName, err))
			return err
		}
	}
	return nil
}
