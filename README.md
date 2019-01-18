# MicroWeb [![Travis](https://travis-ci.com/CanadianCommander/MicroWeb.svg?branch=master)](https://travis-ci.com/CanadianCommander/MicroWeb) [![go report card](https://goreportcard.com/badge/github.com/CanadianCommander/MicroWeb)](https://goreportcard.com/report/github.com/CanadianCommander/MicroWeb)
micro web is a simple and efficient web server intended to let you deploy simple sites FAST. It allows you to server basic static content over HTTP/HTTPS or if you want to get a little bit more fancy you can break out a plugin and customize the behaviour.
## Features
- Simple and light HTTP/HTTPS server
- golang plugin system allowing for customization of server behaviour / web API development
- helper packages to assist in plugin development
- systemd integration

## OS Support
**Official:** Linux x86_64

**Unofficial**: MacOs,  *might work!*
## Build
- Download with `go get github.com/CanadianCommander/MicroWeb`
- Get dependencies with `make getdep`
- Build with `make` or `make build`
- Finally test the server with `./microweb.a -c <config file path> -v verbose`

## Install
- Download with `go get github.com/CanadianCommander/MicroWeb`
- Get dependencies with `make getdep`
- Install with `make install`
- Server binary is now installed in `/bin/`, configuration files in `/etc/microweb/`, and webroot in `/var/www/`. to manage the server use systemctl. Ex: `systemctl status microweb`, `systemctl start microweb`... etc. Finally to get the logs use `journalctl -u microweb`

## Documentation
[wiki](https://github.com/CanadianCommander/MicroWeb/wiki)
