$(function() {
    $("#ibp").submit(function(event) {
        event.preventDefault()
        var env = $("select#env option:selected").val()
        var plan = $("select#plan option:selected").val()
        var data = {env: env, name: plan}
        let uid = getUid()
        if(!uid) {
            alert("no uid found!")
            return
        }
        let createServiceUrl = '/v1/' + uid + '/service'
        if(plan === "ep") {
            let location = $("select#location option:selected").val()
            if(!location) {
                alert("location id is required!")
                return
            }
            let numOfOrgs = parseInt($("select#numOfOrgs option:selected").val())
            let numOfPeers = parseInt($("select#numOfPeers option:selected").val())
            let ledgerType = $("select#ledgerType option:selected").val()
            data = {loc: location, name: plan, env: env,
                config: {
                    numOfOrgs: numOfOrgs,
                    numOfPeers: numOfPeers,
                    ledgerType: ledgerType
                }
            }
        }
        var spReq = $.post(createServiceUrl, JSON.stringify(data))
        spReq.done(function(data) {
            alert("Your request to create IBP network has been successfully received!" + JSON.stringify(data))
            updateIbpList(uid, jenkinsBase, apacheBase)
        })
        spReq.fail(function(err) {
            alert("Error creating IBP network:\n" + JSON.stringify(err, null, 4))
        })
    });

    var locReq
    $("select#plan, select#env").change(function() {
        if(locReq) {
            console.log("Aborting previous location request")
            locReq.abort()
            locReq = null
        }
        var plan = $("select#plan").val()
        if(plan === "ep") {
            $("#ep").show()
            $(".progress").show() // show progress bar besides location id select options
            $("#net").attr("disabled",  true) // disable 'create network' button
            $("select#location").empty()
            var env = $("select#env").val()
            // the url to get network locations
            var url
            switch(env) {
                case "bxstaging":
                    url = "https://ibmblockchain-dev-v2.stage1.ng.bluemix.net/api/v1/network-locations"
                    break;
                case "bxproduction":
                    url = "https://ibmblockchain-v2.ng.bluemix.net/api/v1/network-locations"
                    break;
                default:
                    //do nothing
                    console.log("invalid environment value:", env, ".Should be 'bxstaging' or 'bxproduction'")
                    return
            }
            console.log("network locations url:", url)
            locReq = $.get(url, function(data) {
                if(data.length == 0) { // no data returned, probably because the this req was aborted
                    console.log("locReq data empty")
                    return
                }
                console.log("response from ", url, "\n", JSON.stringify(data, null, 4))
                $.each(data, function(_, val) {
                    if(val.location_id) {
                        if(val.status === "available") {
                            $("select#location").append($('<option>', {
                                value: val.location_id,
                                text: val.location_id
                            }))
                        } else {
                            $("select#location").append('<option disabled="disabled">'+val.location_id+'</option>')
                        }
                    }
                })
                $(".progress").hide() // hide progress bar
                $("#net").attr("disabled", false)
            })
        } else {
            $("#ep").hide()
            $("#net").attr("disabled", false)
        }
    })
})

// extract user id from path
function getUid() {
    var path = window.location.pathname
    var regex = /\/v1\/(\w+)\//g
    var uid = regex.exec(path)
    if(uid) {
        return uid[1]
    }
    return ""
}

// delete service by service id and env
function deleteService(url) {
    $.ajax({
        url: url,
        type: 'DELETE',
        success: function(result) {
          alert("your request has been accepted!", result)
          updateIbpList(uid, jenkinsBase, apacheBase)
        },
        error: function(err) {
            alert("error deleting ibp service:", err)
        }
    });
}

// update available IBP networks list and also pending creating jobs list
let updateIbpListTick
async function updateIbpList(uid, jenkinsBase, apacheBase) {
    if(!uid || uid === "") {
        console.error("no uid avaialble for updating IBP network list")
        return
    }
    try {
        let services = await $.get('/v1/' + uid + '/services')
        let jobs = await $.get('/v1/' + uid + '/pending' )
        console.log("updated network list/request at:", new Date())

        let ibpServices = $("#ibp-services")
        let ibpPending = $("#ibp-pending")
        ibpServices.empty()
        ibpPending.empty()
        $.each(services, function(_, val) {
            ibpServices.append(`
            <tr>
            <td>
                <span>
                    <a target="_blank" href="${apacheBase}/${uid}/${val.networkId}">
                        ${val.networkId}
                    </a>
                </span>
            </td>
            <td>
                <span>
                    <a target="_blank" href="${jenkinsBase}/job/${val.jobName}/${val.jobId}/console" 
                        target="_blank" rel="noopener noreferrer">
                        <img src="/static/images/log.png" alt="log" height="24" width="24" />
                    </a>
                </span>
            </td>
            <td>
                <span>${val.planName}</span>
            </td>
            <td>
                <span>` + ((val.location)? `${val.location}` : `N/A`) + `</span>
            </td>
            <td>
                <span>${val.env}</span>
            </td>  
            <td>
                <span>${val.createdAt}</span>
            </td> 
            <td>
                <button type="submit" class="bx--btn bx--btn--danger" 
                onclick="deleteService('/v1/${uid}/service/${val.serviceId}?env=${val.env}')">
                    Delete
                </button>
            </td>
        </tr>
            `)
        })
        $.each(jobs, function(_, val) {
            ibpPending.append(`
            <tr>
                <td>
                    <span>N/A</span>
                </td>
                <td>
                    <span>
                        <a ` + ((val.jobId&&val.jobId !== "") ? `href="${jenkinsBase}/job/${val.name}/${val.jobId}/console" 
                        target="_blank" rel="noopener noreferrer">
                        <img src="/static/images/log.png" alt="log" height="24" width="24" />` : 
                        `> <img src="/static/images/greylog.png" height="24" width="24" />`) +
                        `</a>
                    </span>
                </td>
                <td>
                    <span>${val.planName}</span>
                </td>
                <td>
                    <span>` + ((val.location)? `${val.location}` : `N/A`) + `</span>
                </td>
                <td>
                    <span>${val.env}</span>
                </td>
            </tr>
            `)
        })
    } catch (err) {
        console.error("Error getting IBP networks and pending jobs list:\n" + JSON.stringify(err, null, 4))
    } finally {
        if(updateIbpListTick) {
            clearTimeout(updateIbpListTick)
        }
        updateIbpListTick = setTimeout(updateIbpList, 10000, uid, jenkinsBase, apacheBase)
    }
}
