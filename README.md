# MicroWeb [![go report card](https://goreportcard.com/badge/github.com/CanadianCommander/MicroWeb)](https://goreportcard.com/report/github.com/CanadianCommander/MicroWeb)
micro web is a simple and efficient web server intended to let you deploy simpile sites FAST. This is not a library! Rather one
would build the micro web executable and through go templates and plugins would customize the server (even possible to do this while the server is running with no downtime).
## Build
- Download with `go get github.com/CanadianCommander/MicroWeb`
- To build run
`go build github.com/CanadianCommander/MicroWeb/cmd/microweb`
- After the server is built modify `resource/default.cfg.json` as needed.
- Next build any required go plugins
- Finally test the server with `./microWeb -c <config file path> -v verbose`

## Install 
to install microweb on to your system (create configuration files under /etc/microweb and install unit file for systemd) run 

`sudo -E ./install/install.sh` then run the server with `sudo systemctl start microweb`

## Documentation 
[wiki](https://github.com/CanadianCommander/MicroWeb/wiki)
