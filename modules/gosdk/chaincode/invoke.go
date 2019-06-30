package chaincode

import (
	"fmt"
	"hfrd/modules/gosdk/chaincode/utils"
	"hfrd/modules/gosdk/common"
	"hfrd/modules/gosdk/utilities"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var chaincodeInvokeCmd = &cobra.Command{
	Use:              "invoke",
	Short:            "invoke chaincode on channel(s)",
	Long:             "invoke chaincode according the connection profile and parameters provided",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return invokeChaincode()
	},
}

const (
	CC_INVOKE      = "chaincode.invoke"
	CC_INVOKE_SENT = "chaincode.invoke.sent"
	CC_INVOKE_FAIL = "chaincode.invoke.fail"
)

// DynamicDiscoveryProviderFactory is configured with dynamic (endorser) selection provider
type DynamicDiscoveryProviderFactory struct {
	defsvc.ProviderFactory
}

type channelProvider struct {
	fab.ChannelProvider
	services map[string]*dynamicdiscovery.ChannelService
}

// CreateLocalDiscoveryProvider returns a new local dynamic discovery provider
func (f *DynamicDiscoveryProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return dynamicdiscovery.NewLocalProvider(config), nil
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *DynamicDiscoveryProviderFactory) CreateChannelProvider(config fab.EndpointConfig, options ...options.Opt) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config)
	if err != nil {
		return nil, err
	}
	return &channelProvider{
		ChannelProvider: chProvider,
		services:        make(map[string]*dynamicdiscovery.ChannelService),
	}, nil
}

func invokeCmd() *cobra.Command {
	flags := chaincodeInvokeCmd.Flags()
	flags.StringVar(&chaincodeName, CC_NAME, "", "chaincode name")
	flags.StringSliceVar(&peers, PEERS, []string{}, "send proposal to the peer list. "+
		"If peers list is empty, then test module will do service discovery to get the peers")
	flags.StringVar(&channelName, CHANNEL_NAME, "", "the channel to send proposal to")
	flags.StringVar(&chaincodeParams, CC_PARAMS, "", "chaincode invoke/query parameters")
	flags.StringSliceVar(&dynamicTransientMapKs, CC_DYNAMIC_TRANSIENTMAP_K, []string{}, "the array of keys in dynamic transient map ")
	flags.StringSliceVar(&dynamicTransientMapVs, CC_DYNAMIC_TRANSIENTMAP_V, []string{}, "the array of values in dynamic transient map")
	flags.StringVar(&staticTransientMap, CC_STATIC_TRANSIENTMAP, "", "the static transient map")
	flags.StringVar(&queryOnly, QUERY_ONLY, "false", "if set to true, gosdk will not send tx to orderer(s)")
	flags.IntVar(&threads, THREADS, 1, "how many threads to send proposal concurrently")
	flags.BoolVar(&serviceDiscovery, SERVICE_DISCOVERY, false, "enable service discovery")
	flags.StringVar(&fabricVersion, FABRIC_VERSION, "1.1", "Use fabricVersion to define different capabilities. Use format 1.x in fabricVersion")
	flags.StringVar(&prometheusTargetUrl, PROMETHEUS_TARGET_URL, "", "if set, hfrd will send metrics to this prometheus endpoint")
	chaincodeInvokeCmd.MarkFlagRequired(CC_NAME)
	chaincodeInvokeCmd.MarkFlagRequired(CHANNEL_NAME)
	chaincodeInvokeCmd.MarkFlagRequired(CC_PARAMS)
	chaincodeInvokeCmd.MarkFlagRequired(PEERS)
	return chaincodeInvokeCmd
}

func invokeChaincode() error {
	common.Delay(viper.GetString(common.DELAY_TIME))
	queryOnlyBool, err := strconv.ParseBool(strings.ToLower(queryOnly))
	if err != nil {
		return errors.Errorf("Error in parsing parameter queryOnly.Should be 'true' or 'false',but got %s", queryOnly)
	}
	// Split chaincode parameter and transientMap(Used in private data chaincode) by #
	chaincodeParamArray := strings.Split(chaincodeParams, "#")
	common.Logger.Debug(fmt.Sprintf("chaincodPrams: %s", chaincodeParamArray))

	// Read connection profile
	connProfile := viper.GetString(common.CONN_PROFILE)
	viperConn, err := common.GetViperInstance(connProfile, "yaml")
	if err != nil {
		return err
	}
	org = viperConn.GetString(CLIENT_ORGANIZATION)
	if org == "" {
		return fmt.Errorf("client.organization is not provided in sdk config yaml")
	}
	if err := viperConn.Unmarshal(&connection); err != nil {
		return errors.WithMessage(err, "unmarshall connection profiles from connection profile error")
	}
	if len(peers) == 0 {
		return errors.New("peers list is required!")
	}
	// TODO: I added all peers returned from service discovery to channel config backend
	p := peers
	channelConfig, err := common.GetTempChannelConfigFile(channelName, p)
	if err != nil {
		return errors.Errorf("cc invoke failed due to errors in GetChannelBackendYaml. %s", err)
	}

	// Initialize sdk with multiple config files
	configBackends, err := common.GetConfigBackends(common.CONFIG_BCCSP, channelConfig, connProfile)
	if err != nil {
		return errors.WithMessage(err, "Unable to get config backends")
	}
	// reconfig configBackends with dynamic service discovery
	if fabricVersion != "1.1" {
		configBackends, err = configBackendsWithSD(connProfile, peers[0], channelName, org, common.ADMIN, true)
		if err != nil {
			return errors.WithMessage(err, "Unable to get config backends with dynamic service discovery")
		}
	}
	sdk, err := fabsdk.New(configBackends, fabsdk.WithServicePkg(&DynamicDiscoveryProviderFactory{}))
	if err != nil {
		return errors.WithMessage(err, "Error creating sdk")
	} else {
		common.Logger.Info("sdk initialized successfully!")
	}

	if threads < 1 {
		// default to 1 go routine
		threads = 1
	}
	var wg sync.WaitGroup
	var successCount uint64
	var failCount uint64
	errChan := make(chan error, threads)
	// TODO: hardcoded to invoke with ADMIN
	clientContext := sdk.ChannelContext(channelName, fabsdk.WithUser(common.ADMIN), fabsdk.WithOrg(org))
	client, err := channel.New(clientContext)
	if err != nil {
		return err
	}
	invokeClient, err := utils.New(clientContext)
	if err != nil {
		return err
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	start := time.Now()
	// go routine loop
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if prometheusTargetUrl != "" {
				viper.Set("PROMETHEUS_TARGET_URL", prometheusTargetUrl)
			}
			base := common.NewBase()
			base.IterationCount = viper.GetString(common.ITERATION_COUNT)
			base.RetryCount = viper.GetInt(common.RETRY_COUNT)
			base.SetIterationInterval(viper.GetString(common.ITERATION_INTERVAL))
			base.ConnectionProfile = connProfile
			hostname, _ := os.Hostname()
			base.Hostname = hostname + "-" + strconv.Itoa(index)
			if err != nil {
				errChan <- err
				common.Logger.Error("failed to create new channel client")
				common.Logger.Error(err.Error())
				return
			}
			cc := &Chaincode{
				Base:         base,
				name:         chaincodeName,
				channel:      channelName,
				args:         nil,
				sdk:          sdk,
				client:       client,
				invokeClient: invokeClient,
				queryOnly:    queryOnlyBool,
			}
			invokeFunc := func(iterationIndex int) error {
				// Get common parameters
				chaincodeArgs, transientStaticMap, transientDynamicMap, err := utilities.GenerateChaincodeParams(chaincodeParamArray, staticTransientMap, dynamicTransientMapKs, dynamicTransientMapVs, iterationIndex)
				if err != nil {
					return err
				}
				cc.args = chaincodeArgs
				if transientStaticMap != nil {
					cc.transientMap = transientStaticMap
				}
				if transientDynamicMap != nil {
					cc.transientMap = transientDynamicMap
				}
				return cc.InvokeChaincode()
			}
			// capture ^C os.SIGINT signal
			go func() {
				for sig := range c {
					if sig == os.Interrupt {
						common.Logger.Info("\nSIGINT signal received, will exit\n")
						cc.PrintMetrics(CC_INVOKE)
						os.Exit(1)
					}
				}
			}()
			common.InitializeMetrics(CC_INVOKE)
			success, fail, err := common.IterateFunc(base, invokeFunc, false)
			atomic.AddUint64(&successCount, success)
			atomic.AddUint64(&failCount, fail)
			common.TrackCount(CC_INVOKE_FAIL, int64(fail))
			cc.PrintMetrics(CC_INVOKE)
			errChan <- err
			return
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	common.Logger.Info(fmt.Sprintf("Test elapsed time: %s", elapsed))
errChanLoop:
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return err
			}
		default:
			// no error found
			break errChanLoop
		}
	}
	return nil
}

func (cc *Chaincode) InvokeChaincode() error {
	common.TrackCount(CC_INVOKE_SENT, 1)
	var err error
	var argsByte [][]byte
	for i := 1; i < len(cc.args); i++ {
		argsByte = append(argsByte, []byte(cc.args[i]))
	}
	defer func(start time.Time) {
		if err == nil {
			common.TrackCount(CC_INVOKE, 1)
			common.Logger.Debug(fmt.Sprintf("e2e tx latency: %s", time.Since(start)))
			common.TrackTime(start, CC_INVOKE)
		}
	}(time.Now())
	if cc.queryOnly {
		_, err = cc.client.Query(channel.Request{ChaincodeID: cc.name, Fcn: cc.args[0],
			Args: argsByte, TransientMap: cc.transientMap}, channel.WithTargetEndpoints(peers...))
	} else {
		if serviceDiscovery {
			common.Logger.Debug(fmt.Sprintf("using service discovery"))
			peers = []string{}
		}
		_, err = cc.invokeClient.Execute(channel.Request{ChaincodeID: cc.name, Fcn: cc.args[0],
			Args: argsByte, TransientMap: cc.transientMap}, utils.WithTargetEndpoints(peers...))
	}
	if err != nil {
		return errors.WithMessage(err, "failed to execute chaincode")
	}
	return nil
}
