
  function multipartformpost(formid, btnid, postdest, event) {
    event.preventDefault();
    var form = $('#'+formid)[0];
    var data = new FormData(form);
  
    // If you want to add an extra field for the FormData
    // data.append("CustomField", "This is some extra data, testing");
  
    // disabled the submit button
    $('#'+btnid).prop("disabled", true);
    $.ajax({
      type: "POST",
      enctype: 'multipart/form-data',
      url: postdest,
      data: data,
      processData: false,
      contentType: false,
      cache: false,
      timeout: 600000,
      success: function (data) {
        $('#'+btnid).prop("disabled", false);
        alert("Your request has been accepted!")
        location.reload()
      },
      error: function (e) {
        $('#'+btnid).prop("disabled", false);
        alert(JSON.stringify(e, null, 4))
        location.reload()
      }
    });  
  }
  
  $(document).ready(function() {
  // show file name when users choose files to upload
      $("#service_config").change(function() {
          var file = $(this).val()
          file = file.replace("C:\\fakepath\\", "")
          $(this).next('.file-label').text(file)
      })
  })

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
        let certsServices = await $.get('/v1/' + uid + '/certsService')
        let certsJobs = await $.get('/v1/' + uid + '/certsPending' )
        console.log("updated network list/request at:", new Date())

        let ibpCertsServices = $("#ibp-certs-services")
        let ibpCertsPending = $("#ibp-certs-pending")
        ibpCertsServices.empty()
        ibpCertsPending.empty()
        $.each(certsServices, function(_, val) {
            ibpCertsServices.append(`
            <tr>
            <td>
                <span>
                   <a target="_blank" href="${apacheBase}/${uid}/${val.serviceId}">
                   ${val.serviceId}
                   </a>
                </span>
            </td> 
            <td>
                <span>
                ${val.env}
                </span>
            </td>
            <td>
                <span>
                    <a target="_blank" href="${jenkinsBase}/job/${val.name}/${val.jobId}/console" 
                        target="_blank" rel="noopener noreferrer">
                        <img src="/static/images/log.png" alt="log" height="24" width="24" />
                    </a>
                </span>
            </td>
          
            <td>
                <span>${val.status}</span>
            </td> 
            <td>
                <button type="submit" class="bx--btn bx--btn--danger" 
                onclick="deleteService('/v1/${uid}/ibpcerts?requestid=${val._id}&version=${val._rev}')">
                    Delete
                </button>
            </td>
        </tr>
            `)
        })
        $.each(certsJobs, function(_, val) {
            console.log("current certjob val="+JSON.stringify(val))
            ibpCertsPending.append(`
            <tr>
                <td>
                    <span>${val.serviceId}</span>
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
               
            </tr>
            `)
        })
    } catch (err) {
        console.error("error getting IBP networks and pending jobs list:", err)
    } finally {
        if(updateIbpListTick) {
            clearTimeout(updateIbpListTick)
        }
        updateIbpListTick = setTimeout(updateIbpList, 10000, uid, jenkinsBase, apacheBase)
    }
}

  