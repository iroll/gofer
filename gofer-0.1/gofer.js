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
//Lines returned by the gopher server consist of 4 tabbed elements:
//display_string, selector_string, host_name, and port_number.
//The first character of display_string is the item_type.

function parse(input) {
  var lines = input.split('\n');            //split the input blob into lines
  
  var index;                                //setting up a loop to parse lines
  var display_text = [];                    //what will eventually go to the screen
  for (index = 0; index < lines.length; ++index){
    var thisline = lines[index];            //grab a line
    var elements = thisline.split('\t');    //break it into elements
    var a = elements[0];
    var item_type = a.charAt(0);            //gotta break the item_type and display_string
    var display_string = a.substr(1);
    var selector_string = elements[1];
    var host_name = elements[2];
    var port_number = elements[3];
                                            //now to handle the item_type cases
    if (item_type == 0) {                   //type 0: link to a Text File Transaction
      var rawline = ['<span class="_0">', display_string, '</span><br>'];
    } else if (item_type == 1) {            //type 1: link to a Menu Transaction
      var rawline = ['<span class="_1">', display_string, '</span><br>'];
    } else {
      var rawline = [display_string, '<br \/>'];
    }
  
    var cookedline = rawline.join('');
    display_text.push(cookedline);
  }
  
  var data = display_text.join('');
  return data;
}


//callout - calls TcpClient and refreshes the screen

function callout(host, port, selector) {
  document.getElementById('output').innerHTML = ""; //clears the previous page
  document.getElementById('host').value = host;     //update the input fields with
  document.getElementById('port').value = port;     //values that were passed to callout
  document.getElementById('selector').value = selector;
  opensocket = new TcpClient(host, port);
  opensocket.connect(function() {
    opensocket.sendMessage(selector);
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
  var selector = document.getElementById('selector').value;
  callout(host, port, selector);

});

//power the links in the viewer
//span.onclick = callout(host,port,selector);











