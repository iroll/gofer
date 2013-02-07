//connect - calls TcpClient
function connect(host, port, query) {
  opensocket = new TcpClient(host, port);
  opensocket.connect(function() {
    opensocket.sendMessage(query);
    opensocket.addResponseListener(function(response) {
      var lines = response.split('\n');
      var output = lines.join('<br/>');
      document.getElementById('output').innerHTML = output;
    });
  });
}

//Make the simple connection form in viewer work
var button = document.getElementById('connect');
button.addEventListener('click', function () {

  var host = document.getElementById('host').value;
  var port = parseInt(document.getElementById('port').value, 10);
  var query = document.getElementById('query').value;
  connect(host, port, query);

});
