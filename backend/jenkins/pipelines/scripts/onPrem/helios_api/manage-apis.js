//------------------------------------------------------------------------------------------
// Join Network - only used for HFRD helios/server/routes/manage-apis.js
//------------------------------------------------------------------------------------------
app.post('/api/v1/networks/:network_id/joinNetwork', middle.basic_auth_for_boatman, function (req, res, next) {
    serviceId = req.body.service_id
    // Step1 : get network doc
    crud.get_network_by_id({ db: dbNetworks, id: req.params.network_id }, function (err_code, net_doc) {
        if (err_code != null) {
            logger.error('[joinOrg] could not find network doc', err_code, network_doc);
            res.status(400).json({ error: 'could not validate existing network' });
        } else {
            // Step2: Build certs object
            for (var org in net_doc.orgs) {
                // check to see if we have already assigned it
                if (org !== 'OrdererOrg' && net_doc.orgs[org].service_id === serviceId) {
                    orgObject = net_doc.orgs[org];
                    // Step3: Add new org into system channel
                    grpc_lib.addMspId2SysChannel({ network_doc: net_doc, certs: orgObject.certs }, function (eCode, resp3) {
                        logger.info("channel update is done.will return")
                        // refresh_cached_network_doc(network_id);         // update cached version of the network document
                        // res.status(200).json(resp);                   // do not call here, else timeout might occur
                    });
                    res.status(200).json("Succeeded to join into network");
                }
            }
        }
    });
});

//------------------------------------------------------------------------------------------
// get or make ibm id doc for a specific network only used for HFRD
//------------------------------------------------------------------------------------------
app.post('/api/v1/networks/:network_id/getIBMIdDoc', apiAuth, function (req, res, next) {
    const ibmid = req.ibmid_guid
    const network_id = req.params.network_id
    async.parallel([
        // --- Get Network Doc --- //
        function (join) {
            crud.get_network_by_id({ db: dbNetworks, id: req.params.network_id }, function (e, network_doc) {
                if (e) {
                    logger.warn('network doc not found by network id' + req.params.network_id, e, network_doc);
                    join(null, null);        // don't pass back error, just pass null for the network doc
                } else {
                    join(null, network_doc);      // got network, pass it along
                }
            });
        },
        function (join) {
            // --- Update IBM ID Doc --- //
            const options = {
                ibm_guid: ibmid,
                network_id: network_id,
                repeat_write: true,  // set this to make sure we get the api key
                refresh_cache: true
            };
            misc_db.update_ibmid_doc(options, function (e, doc) {
                if (e) {
                    logger.error('could not update ibmid doc...', e);
                    res.status(500).json("Failed to update ibmid doc...");
                    return;
                }
                res.status(200).json("Successfully update ibmid doc...");
                return;
            });
        }
    ]);
});