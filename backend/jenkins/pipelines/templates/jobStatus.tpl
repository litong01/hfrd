{
  "status":"${status}",
  "artifacts": {
<% if (status == 'SUCCESS') { for (item in items) { %>\
    "<%= item[0] %>": "<%= item[1] %>"<%= (item==items.last())?'':',' %>
<% } } %>\
  }
}
