![progress-banner](image.gif)

This repo was created as part of ["Build Your Own HTTP server" Challenge](https://app.codecrafters.io/courses/http-server/overview).

[HTTP](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) is the protocol that powers the web. This repo consists of HTTP/1.1 server written in Go that is capable of serving multiple clients.  

### Features
- Simple/No dependency HTTP Server
- Supports GET, POST methods
- Supports parsing Request Body, Headers, Path
- Response includes http-code, headers & body
- Supports GZIP compression for response


### Development
- Run the app: `make run`
- Build linux_amd64 binary: `make build`. The binary will be stored under `bin`
