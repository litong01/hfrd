package chaincode

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	corecomm "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/bccsp/factory"
	discoveryClient "github.com/hyperledger/fabric/discovery/client"
	discoveryProto "github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric/protos/gossip"
	"github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"hfrd/modules/gosdk/chaincode/utils"
	"hfrd/modules/gosdk/common"
	hfrdcommon "hfrd/modules/gosdk/common"
	"path/filepath"
	"regexp"
	"time"
)

var chaincodeCmd = &cobra.Command{
	Use:              "chaincode",
	Short:            "cc related operations, install | instantiate | invoke",
	Long:             "cc related operations, install on peers | instantiate on channels| invoke instantiated cc",
	TraverseChildren: true,
}

var (
	chaincodeNamePrefix   string
	channelNamePrefix     string
	prefixOffset          int
	channelNameList       []string
	chaincodeVersion      string
	path                  string
	peers                 []string
	channelName           string
	org                   string
	chaincodeName         string
	queryOnly             string
	chaincodeParams       string
	staticTransientMap    string
	dynamicTransientMapKs []string
	dynamicTransientMapVs []string
	threads               int
	policystr             string
	collectionsConfigPath string // collections-config file , used to define private data
	connection            *hfrdcommon.ConnectionProfile
	serviceDiscovery      bool
	fabricVersion         string
	prometheusTargetUrl   string
	lang                  string
)

const (
	PEERS_IN_CHANNEL    = "channels.%s.peers"   // defined in sdk-go config yaml file
	CLIENT_ORGANIZATION = "client.organization" // defined in fabric-sdk-go config yaml

	// cli parameters
	CC_NAME_PREFIX            = "chaincodeNamePrefix"
	CHAN_NAME_PREFIX          = "channelNamePrefix"
	PREFIX_OFFSET             = "prefixOffset"
	CHAN_NAME_LIST 			  = "channelNameList"
	CC_VERSION                = "chaincodeVersion"
	CC_PATH                   = "path"
	PEERS                     = "peers"
	CHANNEL_NAME              = "channelName"
	CC_NAME                   = "chaincodeName"
	QUERY_ONLY                = "queryOnly"
	CC_PARAMS                 = "chaincodeParams"
	CC_STATIC_TRANSIENTMAP    = "staticTransientMap"
	CC_DYNAMIC_TRANSIENTMAP_K = "dynamicTransientMapKs"
	CC_DYNAMIC_TRANSIENTMAP_V = "dynamicTransientMapVs"
	THREADS                   = "threads"
	POLICY_STR                = "policyStr"
	COLLECTION_CONFIG_PATH    = "collectionsConfigPath"
	SERVICE_DISCOVERY         = "serviceDiscovery"
	FABRIC_VERSION            = "fabricVersion"
	PROMETHEUS_TARGET_URL     = "prometheusTargetUrl"

	// keys in connection profile
	CP_PEERS = "peers"
	CP_ORGS  = "organizations"
)

// Cmd returns the cobra command for Chaincode
func Cmd() *cobra.Command {
	chaincodeCmd.AddCommand(installCmd())
	chaincodeCmd.AddCommand(instantiateCmd())
	chaincodeCmd.AddCommand(invokeCmd())
	return chaincodeCmd
}

type Chaincode struct {
	*hfrdcommon.Base
	namePrefix        string            // Chaincode NamePrefix for install/instantiate
	name              string            // Chaincode name for invoke/query
	args              []string          // Arguments for invoke/query
	transientMap      map[string][]byte // Chaincode transient map. Used in private data chaincodes
	version           string            // chaincode version
	path              string            // chaincode path on file system: relative to GOPATH env variable
	channel           string            // on which channel to instantiate the chaincode
	channelPrefix     string            // channel name prefix for instantiate cc operation
	sdk               *fabsdk.FabricSDK // the fabric-sdk-go instance to interact with fabric network
	client            *channel.Client   // the client to send chaincode invoke request
	invokeClient      *utils.Client     // our customized client to send chaincode invoke w/ ability to break down latency
	queryOnly         bool
	CollectionsConfig []*cb.CollectionConfig // Collections config used to instantiate pvt(private data) chaincode
}

func configBackendsWithSD(cpPath, peer, channelName, org, identity string, endorserOnly bool) (core.ConfigProvider, error) {
	var configBackends core.ConfigProvider
	if cpPath == "" || peer == "" || org == "" || identity == "" {
		return configBackends, errors.New("Required param is empty")
	}
	viperConn, err := common.GetViperInstance(cpPath, "yaml")
	if err != nil {
		return configBackends, errors.WithMessage(err, "Error reading connection profile yaml")
	}
	channelConfig, err := common.GetTempChannelConfigFile(channelName, []string{peer})
	if err != nil {
		return configBackends, errors.Errorf("Instantiate failed due to errors in GetChannelBackendYaml. %s", err)
	}
	configBackends, err = common.GetConfigBackends(common.CONFIG_BCCSP, channelConfig, cpPath)
	if err != nil {
		return configBackends, err
	}
	common.Logger.Info(fmt.Sprintf("service discovery with peer: %s", peer))
	peersMap := viperConn.GetStringMap(CP_PEERS)
	// is peer a peer name?
	if peerUrl, ok := getPeerUrl(peersMap, peer); ok && len(peerUrl) > 0 {
		// then convert peer name to peer url
		peer = peerUrl
		common.Logger.Info(fmt.Sprintf("service discovery with peer url: %s", peer))
	}
	// do we need to do peer url substitution?
	if peerSubUrl, ok := getPeerSubstitutionUrl(viperConn.Get("entityMatchers.peer"), peer); ok {
		peer = peerSubUrl
		common.Logger.Info(fmt.Sprintf("service discovery with peer url substitution: %s", peer))
	}
	if err := factory.InitFactories(nil); err != nil {
		return configBackends, errors.WithMessage(err, "error initialize bccsp factories")
	}
	// TODO: this is a workaround of https://github.com/spf13/viper/issues/324 and
	// https://github.com/spf13/viper/pull/673
	viperConn.SetKeyDelim("\\\\")
	var connProf common.ConnectionProfile
	if err := viperConn.Unmarshal(&connProf); err != nil || &connProf == nil {
		return configBackends, errors.WithMessage(err, "Error unmarshalling connection profile")
	}
	// construct a map peerUrl-> peerName for peers defined in connection profile
	peersMapInCp := map[string]string{}
	for peerName, peer := range connProf.Peers {
		if len(peer.URL) == 0 || len(peerName) == 0 {
			continue
		}
		peersMapInCp[peer.URL] = peerName
	}
	common.Logger.Debug(fmt.Sprintf("peers in original connection profile: %#v", peersMapInCp))
	endorsers, err := getPeers(configBackends, org, identity, channelName, chaincodeName, peer, endorserOnly, peersMapInCp)
	if err != nil {
		return configBackends, errors.WithMessage(err, "unable to get endorsers")
	}
	if endorserOnly {
		common.Logger.Debug(fmt.Sprintf("[%s %s]discovered endorsers: %s", channelName, chaincodeName, endorsers))
	} else {
		common.Logger.Debug(fmt.Sprintf("[%s]discovered peers: %s", channelName, endorsers))
	}
	if len(endorsers) == 0 {
		return configBackends, errors.New("No endorser found!")
	}
	viperConn.Set(CP_PEERS, endorsers)
	organizations := hfrdcommon.Organizations{Organizations: make(map[string]hfrdcommon.Organization)}
	if err := viperConn.Unmarshal(&organizations); err != nil {
		return configBackends, errors.WithMessage(err, "Unable to unmarshal organization from viperconn")
	}

	// clean up the peer list for each org defined in connection profile
	for orgName, org := range organizations.Organizations {
		org.Peers = []string{}
		organizations.Organizations[orgName] = org
	}
endorsersLoop:
	for name, endorser := range endorsers {
		for orgName, org := range organizations.Organizations {
			if org.Mspid == endorser.MspId {
				org.Peers = append(org.Peers, name)
				organizations.Organizations[orgName] = org
				continue endorsersLoop
			}
		}
		organizations.Organizations[endorser.MspId] = hfrdcommon.Organization{
			Mspid: endorser.MspId,
			Peers: []string{name},
		}

	}

	viperConn.Set(CP_ORGS, organizations.Organizations)

	viperConn.SetKeyDelim("\\\\")
	var extension = filepath.Ext(cpPath)
	var newCpFileName = cpPath[0:len(cpPath)-len(extension)] + "-service-discovery.yaml"
	common.Logger.Debug(fmt.Sprintf("New connection profile file name: %s", newCpFileName))
	if err := viperConn.WriteConfigAs(newCpFileName); err != nil {
		return configBackends, errors.WithMessage(err, "Unable to write connection profile")
	}
	i := 0
	endorserNames := make([]string, len(endorsers))
	for k := range endorsers {
		endorserNames[i] = k
		i++
	}
	channelConfig, err = common.GetTempChannelConfigFile(channelName, endorserNames)
	if err != nil {
		return configBackends, errors.WithMessage(err, "Failed to get channel config")
	}

	// Initialize sdk with multiple config files
	configBackends, err = common.GetConfigBackends(common.CONFIG_BCCSP, channelConfig, newCpFileName)
	if err != nil {
		return configBackends, errors.WithMessage(err, "Unable to get config backends")
	}
	return configBackends, nil
}

type endorser struct {
	Url        string
	MspId      string
	TlsCACerts map[string]string
}

func getPeers(configBackends core.ConfigProvider, org, user, channelName, chaincodeName,
	bootstrapPeerUrl string, endorserOnly bool, peersMapInCp map[string]string) (map[string]endorser, error) {
	endorsers := make(map[string]endorser)
	sdk, err := fabsdk.New(configBackends)
	if err != nil {
		return endorsers, err
	}
	chCtx := sdk.ChannelContext(channelName, fabsdk.WithUser(user), fabsdk.WithOrg(org))
	channel.New(chCtx)
	clientProvider := sdk.Context(fabsdk.WithUser(user), fabsdk.WithOrg(org))
	clientCtx, err := clientProvider()
	if err != nil {
		return endorsers, errors.WithMessage(err, "Error creating clientCtx")
	}
	if clientCtx == nil {
		return endorsers, errors.New("Empty clientCtx")
	}
	target := fab.PeerConfig{
		URL: bootstrapPeerUrl,
		GRPCOptions: map[string]interface{}{
			"allow-insecure": true,
		},
	}
	opts := comm.OptsFromPeerConfig(&target)
	opts = append(opts, comm.WithConnectTimeout(10*time.Second))

	conn, err := comm.NewConnection(clientCtx, target.URL, opts...)
	if err != nil {
		return endorsers, errors.WithMessage(err, "Error creating connection")
	}
	defer conn.Close()
	reqCtx, cancel := context.NewRequest(clientCtx, context.WithTimeout(10*time.Second))
	defer cancel()
	var req *discoveryClient.Request
	if !endorserOnly {
		req = discoveryClient.NewRequest().OfChannel(channelName).AddPeersQuery(&discoveryProto.ChaincodeCall{
			Name: chaincodeName, CollectionNames: []string{}})
	} else {
		req, err = discoveryClient.NewRequest().OfChannel(channelName).
			AddEndorsersQuery(&discoveryProto.ChaincodeInterest{
				Chaincodes: []*discoveryProto.ChaincodeCall{{Name: chaincodeName, CollectionNames: []string{}}}})
		if err != nil {
			return endorsers, errors.WithMessage(err, "Error creating discovery request")
		}
		req = req.AddConfigQuery()
	}

	// auth info
	authIdentity, err := clientCtx.Serialize()
	if err != nil {
		return endorsers, errors.WithMessage(err, "Error creating identity")
	}
	hash, err := corecomm.TLSCertHash(clientCtx.EndpointConfig())
	if err != nil {
		return endorsers, errors.WithMessage(err, "Error creating tls cert hash")
	}

	reqToBeSent := *req.Request
	reqToBeSent.Authentication = &discoveryProto.AuthInfo{ClientIdentity: authIdentity, ClientTlsCertHash: hash}
	payload, err := proto.Marshal(&reqToBeSent)
	if err != nil {
		return endorsers, errors.WithMessage(err, "Error marshalling reqToBeSent")
	}
	sig, err := discoveryClient.NewMemoizeSigner(func(msg []byte) ([]byte, error) {
		return clientCtx.SigningManager().Sign(msg, clientCtx.PrivateKey())
	}, 0).Sign(payload)
	if err != nil {
		return endorsers, errors.WithMessage(err, "Error signing payload")
	}
	connection := conn.ClientConn()
	cl := discoveryProto.NewDiscoveryClient(connection)
	res, err := cl.Discover(reqCtx, &discoveryProto.SignedRequest{Payload: payload, Signature: sig})

	if err != nil {
		return endorsers, errors.WithMessage(err,
			fmt.Sprintf("Error discovering endorsers for channel[%s] and cc[%s]", channelName, chaincodeName))
	}
	common.Logger.Debug("gathering all endorsers")
	if len(res.Results) > 1 && res.Results[0].GetError() == nil && res.Results[1].GetError() == nil {
		configResults := res.Results[1].GetConfigResult()
		if configResults == nil {
			return endorsers, errors.New("Unable to discover config for channel" + channelName)
		}
		mspConfig := configResults.Msps
		if mspConfig == nil {
			return endorsers, errors.New("Unable to get msp config for channel" + channelName)
		}
		if endorserOnly {
			ccQueryRes := res.Results[0].GetCcQueryRes()
			for _, descriptor := range ccQueryRes.Content {
				for _, endorsersGroup := range descriptor.EndorsersByGroups {
					for _, peer := range endorsersGroup.Peers {
						if e := formatPeer(peer, mspConfig); e != nil {
							if peerName, ok := peersMapInCp[e.Url]; ok {
								endorsers[peerName] = *e
							} else {
								endorsers[e.Url] = *e
							}
						}
					}

				}
			}
		} else {
			for _, v := range res.Results[0].GetMembers().PeersByOrg {
				for _, peer := range v.Peers {
					if e := formatPeer(peer, mspConfig); e != nil {
						if peerName, ok := peersMapInCp[e.Url]; ok {
							endorsers[peerName] = *e
						} else {
							endorsers[e.Url] = *e
						}
					}
				}
			}
		}

	} else {
		for i, result := range res.Results {
			if result.GetError() != nil {
				common.Logger.Warn(fmt.Sprintf("response result %d error: %s,", i, result.GetError().Content))
			}
		}
	}
	return endorsers, nil
}

// Given the peersMap from connection profile and the peer name
// Return peerUrl of the peerName if found and true
// OR
// Return empty string and false if not found
func getPeerUrl(peersMap interface{}, peerName string) (string, bool) {
	if peersMap, ok := peersMap.(map[string]interface{}); ok {
		if peerMap, ok := peersMap[peerName]; ok {
			if peerMap, ok := peerMap.(map[string]interface{}); ok {
				if peerUrl, ok := peerMap["url"]; ok {
					if peerUrl, ok := peerUrl.(string); ok {
						return peerUrl, true
					}
				}
			}
		}
	}
	return "", false
}

func getPeerSubstitutionUrl(peerMatchers interface{}, peerUrl string) (string, bool) {
	if peerMatchers, ok := peerMatchers.([]interface{}); ok {
		for _, peerMatcher := range peerMatchers {
			if peerMatcher, ok := peerMatcher.(map[interface{}]interface{}); ok {
				if pattern, ok := peerMatcher["pattern"]; ok {
					if pattern, ok := pattern.(string); ok && len(pattern) > 0 {
						if match, _ := regexp.MatchString(pattern, peerUrl); match {
							if urlSubExp, ok := peerMatcher["urlSubstitutionExp"]; ok {
								if urlSubExp, ok := urlSubExp.(string); ok {
									return urlSubExp, true
								}

							}
						}
					}
				}
			}
		}
	}
	return "", false
}

func getPeerEndpoint(env *gossip.SignedGossipMessage) string {
	if env == nil {
		return ""
	}
	aliveMsg, _ := env.ToGossipMessage()
	if aliveMsg == nil {
		return ""
	}
	if !aliveMsg.IsAliveMsg() || aliveMsg.GetAliveMsg().Membership == nil {
		return ""
	}
	return aliveMsg.GetAliveMsg().Membership.Endpoint
}

func formatPeer(peer *discoveryProto.Peer, mspConfig map[string]*msp.FabricMSPConfig) *endorser {
	sID := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(peer.Identity, sID); err != nil || len(sID.Mspid) == 0 {
		common.Logger.Warn(fmt.Sprintf("Unable to get peer identity: %s", err))
		return nil
	}
	env, _ := peer.MembershipInfo.ToGossipMessage()
	if endpoint := getPeerEndpoint(env); endpoint != "" && mspConfig[sID.Mspid] != nil {
		var pem string
		for _, pb := range mspConfig[sID.Mspid].TlsRootCerts {
			pem += string(pb)
		}
		return &endorser{
			Url:   endpoint,
			MspId: sID.Mspid,
			TlsCACerts: map[string]string{
				"pem": pem,
			},
		}
	}
	return nil
}
