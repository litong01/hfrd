function updateTest(uid) {
    $.getJSON("/v1/"+uid+"/moduletest")
        .done(function( json ) {
            $.each(json.items, function(index, item){
                $('#a1'+item.id).attr("href", json.apachebase+"/"+uid+"/"+item.id)
                $('#a2'+item.id).text(item.status);
                if (item.jobid != "") {
                    $('#a3'+item.id).attr("href", json.consolebase + "/job/moduletest/"+item.jobid+"/console")
                }
                if (item.chartpath != "") {
                    $('#a4'+item.id).attr("href", item.chartpath);
                }
                $('#a5' + item.id).text(item.cdate)
            })
        })
        .fail(function( jqxhr, textStatus, error ) {
            var err = textStatus + ", " + error;
            console.log( "Request Failed: " + err );
        });
    setTimeout(updateTest, 5000, uid)
}

function updateNet(uid) {
    $.getJSON("/v1/"+uid+"/network")
        .done(function( json ) {
            $.each(json.items, function(index, item){
                $('#n1'+item.id).attr("href", json.apachebase+"/"+uid+"/"+item.id)
                $('#n2'+item.id).text(item.status);
                if (item.jobid != "") {
                    $('#n3'+item.id).attr("href", json.consolebase + "/job/k8snetwork/"+item.jobid+"/console")
                }
                $('#n4'+item.id).text(item.cdate)
            })
        })
        .fail(function( jqxhr, textStatus, error ) {
            var err = textStatus + ", " + error;
            console.log( "Request Failed: " + err );
        });
    setTimeout(updateNet, 10000, uid)
}

function senddelete(theurl) {
    $.ajax({
        url: theurl,
        type: 'DELETE',
        success: function(result) {
            alert("your request has been accepted!")
        }
    });
}

function modalpartformpost(formid,postdest) {
    var form = $('#' + formid)[0];
    var data = new FormData(form);

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
            alert("Your request has been accepted!")
            location.reload()
        },
        error: function (e) {
            alert(JSON.stringify(e, null, 4))
            location.reload()
        }
    });
}


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
    $("#kubeconfig, #config, #cert, #plan, #chaincode").change(function() {
        var file = $(this).val()
        file = file.replace("C:\\fakepath\\", "")
        $(this).next('.file-label').text(file)
    })
})
