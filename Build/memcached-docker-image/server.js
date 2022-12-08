var http = require('http');
var os = require('os');

var totalrequests = 0;
console.log('count: %s', os.hostname());
http.createServer(function(request, response) {
  
  var fs = require('fs');   
  var contents = fs.readFileSync('report.txt', 'utf8');
  console.log(contents);
  
  totalrequests =  Number(contents)
  response.writeHead(200);

  if (request.url == "/metrics") {
    response.end("# HELP http_requests_total The amount of requests served by the server in total\n# TYPE http_requests_total counter\ncustom_metric " + totalrequests + "\n");
   return;
  }
  response.end("Hello! My name is " + os.hostname() + ". I have served "+ totalrequests + " requests so far.\n");
  
}).listen(8080)


