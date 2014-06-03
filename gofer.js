/*
  gofer.js - a simple gopher client for Google Chrome.

  Copyright 2013 Isaac Roll

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

//parse - parses data returned by TcpClient for display
//todo: clean up the line headers, detect links

function parse(input) {
  var lines = input.split('\n');
  var number = length(lines);
  var text = lines.join('<br/>');
  var assembled = ['<H1>Lines returned: ', number, '</H1><br/>', text];
  var data = assembled.join();
  return data;
}

//connect - calls TcpClient. note that TcpClient appends CR to query

function connect(host, port, query) {
  opensocket = new TcpClient(host, port);
  opensocket.connect(function() {
    opensocket.sendMessage(query);
    opensocket.addResponseListener(function(response) {
      var output = parse(response);
      document.getElementById('output').insertAdjacentHTML('beforeEnd', output);
    });
  });
}

//power the input form in the viewer
var button = document.getElementById('connect');
button.addEventListener('click', function () {

  var host = document.getElementById('host').value;
  var port = parseInt(document.getElementById('port').value, 10);
  var query = document.getElementById('query').value;
  document.getElementById('output').innerHTML = ""; //clears the previous page
  connect(host, port, query);

});
