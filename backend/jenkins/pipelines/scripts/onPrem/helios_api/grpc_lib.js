//------------------------------------------------------------------------------------------
//  only used for HFRD helios/server/libs/grpc_lib.js
//------------------------------------------------------------------------------------------
// fill in the options variable - ALWAYS call this before calling a fc wrangler function
// - it will find/filter for running peers and orderers and format tls certs correctly
grpc_lib.populatePeersOrderers = function (network_id, options, cb) {
    let network_doc = null;
    const skip_cache = (options.SKIP_CACHE === 'true' || options.SKIP_CACHE === true) ? true : false;	// convert to boolean

    // get the network doc if its not provided
    if (skip_cache || !options.network_doc) {
        logger.debug('[gRPC Lib] fetching network doc to get populate peers/orderers');
        crud.get_network_by_id({ db: dbNetworks, id: network_id, SKIP_CACHE: skip_cache }, (errCode, net_doc) => {
            if (errCode != null) {
                logger.error('[gRPC Lib] error getting network doc ' + network_id, errCode);
                if (cb) cb({ error: 'error getting network doc' }, null);
            } else {
                network_doc = net_doc;
                do_work();
            }
        });
    } else {
        network_doc = options.network_doc;					// if network doc is provided, just use it
        do_work();
    }

    // do the actual work here
    function do_work() {

        // --- Decide if we need to switch the node ids out with peer URL --- //
        let peer_node_ids = convertToNodeId(options, 'peer', network_doc);
        let orderer_node_ids = convertToNodeId(options, 'orderer', network_doc);
        if (peer_node_ids === -1 || orderer_node_ids === -1) {
            if (peer_node_ids === -1) {
                if (cb) cb({ parsed: '1 or more provided peer names does not exist' }, null);
            } else {
                if (cb) cb({ parsed: '1 or more provided orderer names does not exist' }, null);
            }
        } else {
            let nodes = [];
            if (!orderer_node_ids || orderer_node_ids.length === 0) {				// if no orderer urls are provided, load all of them
                const orderers = misc.getOrderers(network_doc);
                for (let i in orderers) {
                    orderer_node_ids.push(orderers[i].node_id);						// this is empty, fill it in with orderers
                }
            }
            if (!peer_node_ids || peer_node_ids.length === 0) {						// if no peer urls are provided, load all of them
                const peers = misc.getPeers(network_doc, options.enrollment.msp_id);
                for (let i in peers) {
                    peer_node_ids.push(peers[i].node_id);							// this is empty, fill it in with peers
                }
                logger.info("peer added :", peer_node_ids)
            }
            nodes = peer_node_ids.concat(orderer_node_ids);

            // Filter out the non-running nodes
            //filter_nodes(nodes, network_doc, function (_, running_nodes) {
            misc.formatNodes(nodes, network_doc, options.enrollment.msp_id, function (_, data) {
                logger.debug('[gRPC Lib] will use ' + data.peer_urls.length + ' peer(s) and ' + data.orderer_urls.length + ' orderer(s)');
                options.peers = data.peers;										// copy all the data to the options variable
                options.peer_urls = data.peer_urls;
                options.orderers = data.orderers;
                options.orderer_urls = data.orderer_urls;
                options.enrollment.orderer_url = data.orderer_urls[0];			// the first orderer is used during build enrollment
                options.enrollment.network_doc = network_doc;					// set this so enrollment doesn't need to fetch it again
                options.fabric_version = misc.getNetworkVersion(network_doc);
                options.fabric_experimental = misc.getExperimental(network_doc);
                options.orderer_msp_id = misc.getOrderersMspId(network_doc);
                if (cb) cb(null, options);
            });
            //});
        }
    }

    // only work with running nodes
    function filter_nodes(node_ids, net_doc, cb) {
        let running_ids = [];
        var statusOptions = {
            network_id: network_id,
            network_doc: net_doc,
            msp_id: options.enrollment.msp_id,
            node_ids_2_check: node_ids,
            req: options.req || options.enrollment.req
        };

        logger.debug('[gRPC Lib] getting running nodes for grpc request', net_doc._id);
        grpc_lib.getRunningNodes(statusOptions, (statusErrCode, statusResp) => {
            if (statusErrCode !== null) {
                logger.error('[gRPC Lib] error checking the node statuses ' + statusErrCode);
                cb(statusErrCode, statusResp);
            } else {
                for (const i in node_ids) {
                    const nodeId = node_ids[i];
                    if (nodeId && statusResp[nodeId] && statusResp[nodeId].status && statusResp[nodeId].status.toLowerCase() === 'running') {
                        running_ids.push(nodeId);
                    }
                }
                if (running_ids.length === 0) {
                    logger.error('[gRPC Lib] there are no running nodes... this is likely a problem');
                }
                cb(null, running_ids);
            }
        });
    }
};