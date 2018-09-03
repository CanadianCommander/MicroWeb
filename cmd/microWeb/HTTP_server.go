package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

type HttpServer struct {
	server      *http.Server
	tcpListener net.Listener
}

func (svr *HttpServer) ServeHTTP() {
	if globalSettings.IsTLSEnabled() {
		logger.LogInfo("Serving HTTPS on: %s", svr.tcpListener.Addr().String())
		svr.server.ServeTLS(svr.tcpListener, globalSettings.GetCertFile(), globalSettings.GetKeyFile())
	} else {
		logger.LogInfo("Serving HTTP on: %s", svr.tcpListener.Addr().String())
		svr.server.Serve(svr.tcpListener)
	}
}

func CreateHTTPServer(port string, proto string, errLogger *log.Logger) (*HttpServer, error) {
	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/", HandleResourceRequest)

	readTimout, rtErr := time.ParseDuration(globalSettings.GetHttpReadTimeout())
	if rtErr != nil {
		logger.LogError("Could not parse read timeout: %s. defaulting to 1 second",
			globalSettings.GetHttpReadTimeout())
		readTimout, _ = time.ParseDuration("1s")
	}

	writeTimout, wtErr := time.ParseDuration(globalSettings.GetHttpResponseTimeout())
	if wtErr != nil {
		logger.LogError("Could not parse response timeout: %s. defaulting to 1 second",
			globalSettings.GetHttpResponseTimeout())
		writeTimout, _ = time.ParseDuration("1s")
	}

	var srv HttpServer
	srv.server = &http.Server{
		Addr:         port,
		Handler:      srvMux,
		ErrorLog:     errLogger,
		ReadTimeout:  readTimout,
		WriteTimeout: writeTimout}

	var netErr error
	srv.tcpListener, netErr = net.Listen(proto, port)
	if netErr != nil {
		logger.LogError("Failed to create TCP socket using protocol: %s on port: %s", proto, port)
		return nil, netErr
	}

	return &srv, nil
}
