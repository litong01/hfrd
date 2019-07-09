package channel

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/spf13/cobra"
	hfrdcommon "hfrd/modules/gosdk/common"
)

const (
	CLIENT_ORGANIZATION = "client.organization" // defined in fabric-sdk-go config yaml
	CH_NAME_PREFIX      = "channelNamePrefix"
	CH_NAME             = "channelName"
	CH_NAME_LIST        = "channelNameList"
	PREFIX_OFFSET       = "prefixOffset"
	CH_CONSORTIUM       = "channelConsortium"
	CH_PEER_ORGS        = "channelOrgs"
	CH_ORDERER_NAME     = "ordererName"
	CH_PEERS            = "peers"         // define the peers that will join the channel
	APPLICATION_CAPABILITY      = "applicationCapability" // use to define application capability

	// Orderer channel configurations
	ORDERER_ORG                      = "ordererOrgName"
	CH_BATCH_TIMEOUT                 = "batchTimeout"
	CH_BATCH_SIZE_MAX_MESSAGE_COUNT  = "maxMessageCount"
	CH_BATCH_SIZE_PREFERRED_MAX_BATES = "preferredMaxBytes"
	CH_ORDERER_ADDRESSES_ACTION      = "ordererAddressesAction"
	CH_ORDERER_ADDRESSES             = "ordererAddresses"

	// Peer channel Configurations
	ANCHOR_PEERS = "anchorPeers"

	// channel add new org config file path
	ORG_CONFIG_PATH = "orgConfigPath"
)

var (
	channelNamePrefix      string
	channelName            string
	channelNameList        []string
	prefixOffset           int
	channelConsortium      string
	ordererName            string
	channelOrgs            []string
	peers                  []string
	ordererOrgName         string
	batchTimeout           string
	maxMessageCount        float64
	preferredMaxBytes       float64
	ordererAddressesAction string
	ordererAddresses       []string
	anchorPeers            []string
	newOrgConfigPath       string
	applicationCapability          string
)

type ChannelConfig struct {
	*hfrdcommon.Base
	namePrefix string
	sdk        *fabsdk.FabricSDK // the fabric-sdk-go instance to interact with fabric network

}

var connection *hfrdcommon.ConnectionProfile

var channelCmd = &cobra.Command{
	Use:              "channel",
	Short:            "channel related functions , create | join | update | query | addorg",
	Long:             "channel related functions , create channel | join peer(s) on channel | update channel configurations | query channel configurations | add new org to exsiting channel",
	TraverseChildren: true,
}

func Cmd() *cobra.Command {
	channelCmd.AddCommand(createCmd())
	channelCmd.AddCommand(joinCmd())
	channelCmd.AddCommand(updateCmd())
	channelCmd.AddCommand(queryCmd())
	channelCmd.AddCommand(addorgCmd())
	return channelCmd
}
