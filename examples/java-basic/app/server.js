const http = require("node:http");

const message = process.env.MESSAGE || "Hello from Zerops";
const version = "v1";

http
  .createServer((req, res) => {
    res.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
    res.end(`${message} (${version})\n`);
  })
  .listen(3000, "0.0.0.0");
