package utils

import (
	reqContext "context"
	"time"

	ch "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/filter"
	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/pkg/errors"
)

// Client enables access to a channel on a Fabric network.
//
// A channel client instance provides a handler to interact with peers on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the channel client for each channel. Channel client supports non-admin functions only.
type Client struct {
	context      context.Channel
	membership   fab.ChannelMembership
	eventService fab.EventService
	greylist     *greylist.Filter
}

// RequestOption func for each Opts argument
type RequestOption func(ctx context.Client, opts *requestOptions) error

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// New returns a Client instance. Channel client can query chaincode, execute chaincode and register/unregister for chaincode events on specific channel.
func New(channelProvider context.ChannelProvider, opts ...ClientOption) (*Client, error) {

	channelContext, err := channelProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create channel context")
	}

	greylistProvider := greylist.New(channelContext.EndpointConfig().Timeout(fab.DiscoveryGreylistExpiry))

	if channelContext.ChannelService() == nil {
		return nil, errors.New("channel service not initialized")
	}

	eventService, err := channelContext.ChannelService().EventService()
	if err != nil {
		return nil, errors.WithMessage(err, "event service creation failed")
	}

	membership, err := channelContext.ChannelService().Membership()
	if err != nil {
		return nil, errors.WithMessage(err, "membership creation failed")
	}

	channelClient := newClient(channelContext, membership, eventService, greylistProvider)

	for _, param := range opts {
		err := param(&channelClient)
		if err != nil {
			return nil, errors.WithMessage(err, "option failed")
		}
	}

	return &channelClient, nil
}

// opts allows the user to specify more advanced options
type requestOptions struct {
	Targets       []fab.Peer // targets
	TargetFilter  fab.TargetFilter
	TargetSorter  fab.TargetSorter
	Retry         retry.Opts
	BeforeRetry   retry.BeforeRetryHandler
	Timeouts      map[fab.TimeoutType]time.Duration //timeout options for channel client operations
	ParentContext reqContext.Context                //parent grpc context for channel client operations (query, execute, invokehandler)
	CCFilter      invoke.CCFilter
}

// Execute prepares and executes transaction using request and optional request options
//  Parameters:
//  request holds info about mandatory chaincode ID and function
//  options holds optional request options
//
//  Returns:
//  the proposal responses from peer(s)
func (cc *Client) Execute(request ch.Request, options ...RequestOption) (ch.Response, error) {
	options = append(options, addDefaultTimeout(fab.Execute))
	options = append(options, addDefaultTargetFilter(cc.context, filter.EndorsingPeer))
	return cc.InvokeHandler(newExecuteHandler(), request, options...)
}

// WithTargetFilter specifies a per-request target peer-filter
func withTargetFilter(filter fab.TargetFilter) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		o.TargetFilter = filter
		return nil
	}
}

//WithTimeout encapsulates key value pairs of timeout type, timeout duration to Options
func withTimeout(timeoutType fab.TimeoutType, timeout time.Duration) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		if o.Timeouts == nil {
			o.Timeouts = make(map[fab.TimeoutType]time.Duration)
		}
		o.Timeouts[timeoutType] = timeout
		return nil
	}
}

// addDefaultTargetFilter adds default target filter if target filter is not specified
func addDefaultTargetFilter(chCtx context.Channel, ft filter.EndpointType) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		if len(o.Targets) == 0 && o.TargetFilter == nil {
			return withTargetFilter(filter.NewEndpointFilter(chCtx, ft))(ctx, o)
		}
		return nil
	}
}

// addDefaultTimeout adds default timeout if timeout is not specified
func addDefaultTimeout(tt fab.TimeoutType) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		if o.Timeouts[tt] == 0 {
			return withTimeout(tt, ctx.EndpointConfig().Timeout(tt))(ctx, o)
		}
		return nil
	}
}

// InvokeHandler invokes handler using request and optional request options provided
//  Parameters:
//  handler to be invoked
//  request holds info about mandatory chaincode ID and function
//  options holds optional request options
//
//  Returns:
//  the proposal responses from peer(s)
func (cc *Client) InvokeHandler(handler invoke.Handler, request ch.Request, options ...RequestOption) (ch.Response, error) {
	//Read execute tx options
	txnOpts, err := cc.prepareOptsFromOptions(cc.context, options...)
	if err != nil {
		return ch.Response{}, err
	}

	reqCtx, cancel := cc.createReqContext(&txnOpts)
	defer cancel()

	//Prepare context objects for handler
	requestContext, clientContext, err := cc.prepareHandlerContexts(reqCtx, request, txnOpts)
	if err != nil {
		return ch.Response{}, err
	}

	invoker := retry.NewInvoker(
		requestContext.RetryHandler,
		retry.WithBeforeRetry(
			func(err error) {
				if requestContext.Opts.BeforeRetry != nil {
					requestContext.Opts.BeforeRetry(err)
				}

				cc.greylist.Greylist(err)

				// Reset context parameters
				requestContext.Opts.Targets = txnOpts.Targets
				requestContext.Error = nil
				requestContext.Response = invoke.Response{}
			},
		),
	)

	complete := make(chan bool, 1)
	go func() {
		_, _ = invoker.Invoke( // nolint: gas
			func() (interface{}, error) {
				handler.Handle(requestContext, clientContext)
				return nil, requestContext.Error
			})
		complete <- true
	}()
	select {
	case <-complete:
		return ch.Response(requestContext.Response), requestContext.Error
	case <-reqCtx.Done():
		return ch.Response{}, status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"request timed out or been cancelled", nil)
	}
}

//createReqContext creates req context for invoke handler
func (cc *Client) createReqContext(txnOpts *requestOptions) (reqContext.Context, reqContext.CancelFunc) {

	if txnOpts.Timeouts == nil {
		txnOpts.Timeouts = make(map[fab.TimeoutType]time.Duration)
	}

	//setting default timeouts when not provided
	if txnOpts.Timeouts[fab.Execute] == 0 {
		txnOpts.Timeouts[fab.Execute] = cc.context.EndpointConfig().Timeout(fab.Execute)
	}

	reqCtx, cancel := contextImpl.NewRequest(cc.context, contextImpl.WithTimeout(txnOpts.Timeouts[fab.Execute]),
		contextImpl.WithParent(txnOpts.ParentContext))
	//Add timeout overrides here as a value so that it can be used by immediate child contexts (in handlers/transactors)
	reqCtx = reqContext.WithValue(reqCtx, contextImpl.ReqContextTimeoutOverrides, txnOpts.Timeouts)

	return reqCtx, cancel
}

//prepareHandlerContexts prepares context objects for handlers
func (cc *Client) prepareHandlerContexts(reqCtx reqContext.Context, request ch.Request, o requestOptions) (*invoke.RequestContext, *invoke.ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	transactor, err := cc.context.ChannelService().Transactor(reqCtx)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to create transactor")
	}

	selection, err := cc.context.ChannelService().Selection()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to create selection service")
	}

	discovery, err := cc.context.ChannelService().Discovery()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to create discovery service")
	}

	peerFilter := func(peer fab.Peer) bool {
		if !cc.greylist.Accept(peer) {
			return false
		}
		if o.TargetFilter != nil && !o.TargetFilter.Accept(peer) {
			return false
		}
		return true
	}

	var peerSorter selectopts.PeerSorter
	if o.TargetSorter != nil {
		peerSorter = func(peers []fab.Peer) []fab.Peer {
			return o.TargetSorter.Sort(peers)
		}
	}

	clientContext := &invoke.ClientContext{
		Selection:    selection,
		Discovery:    discovery,
		Membership:   cc.membership,
		Transactor:   transactor,
		EventService: cc.eventService,
	}

	requestContext := &invoke.RequestContext{
		Request:         invoke.Request(request),
		Opts:            invoke.Opts(o),
		Response:        invoke.Response{},
		RetryHandler:    retry.New(o.Retry),
		Ctx:             reqCtx,
		SelectionFilter: peerFilter,
		PeerSorter:      peerSorter,
	}

	return requestContext, clientContext, nil
}

//prepareOptsFromOptions Reads apitxn.Opts from Option array
func (cc *Client) prepareOptsFromOptions(ctx context.Client, options ...RequestOption) (requestOptions, error) {
	txnOpts := requestOptions{}
	for _, option := range options {
		err := option(ctx, &txnOpts)
		if err != nil {
			return txnOpts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return txnOpts, nil
}

func newClient(channelContext context.Channel, membership fab.ChannelMembership, eventService fab.EventService, greylistProvider *greylist.Filter) Client {
	channelClient := Client{
		membership:   membership,
		eventService: eventService,
		greylist:     greylistProvider,
		context:      channelContext,
	}
	return channelClient
}

//WithTargets allows overriding of the target peers for the request
func withTargets(targets ...fab.Peer) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {

		// Validate targets
		for _, t := range targets {
			if t == nil {
				return errors.New("target is nil")
			}
		}

		o.Targets = targets
		return nil
	}
}

// WithTargetEndpoints allows overriding of the target peers for the request.
// Targets are specified by name or URL, and the SDK will create the underlying peer
// objects.
func WithTargetEndpoints(keys ...string) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		var targets []fab.Peer
		for _, url := range keys {
			peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), url)
			if err != nil {
				return err
			}
			peer, err := ctx.InfraProvider().CreatePeerFromConfig(peerCfg)
			if err != nil {
				return errors.WithMessage(err, "creating peer from config failed")
			}
			targets = append(targets, peer)
		}
		return withTargets(targets...)(ctx, opts)
	}
}
