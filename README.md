# MicroWeb
micro web is a simple and efficient web server intended to let you deploy simpile sites FAST. This is not a library! Rather one 
would build the micro web executable and through go templates and plugins would customize the server (even possible to do this while the server is running with no downtime).
## Build 
- Place the source directory under your go path and run: 
`go build microWeb/cmd/microWeb` or `go install microWeb/cmd/microWeb`
- After the server is built modify `resource/default.cfg.json` as needed.
- Next build any required go plugins 
- Finally test the server with `./microWeb -c <config file path> -v verbose`