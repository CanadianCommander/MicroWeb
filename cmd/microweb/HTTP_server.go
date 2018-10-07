package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

// HTTPServer contains an, http server + tcp connection all wrapped up in to one struct
type HTTPServer struct {
	server      *http.Server
	tcpListener net.Listener
}

/*
ServeHTTP start the http server. BLOCKS until server exits
*/
func (svr *HTTPServer) ServeHTTP() {
	if mwsettings.GlobalSettings.IsTLSEnabled() {
		logger.LogInfo("Serving HTTPS on: %s", svr.tcpListener.Addr().String())
		svr.server.ServeTLS(svr.tcpListener, mwsettings.GlobalSettings.GetCertFile(), mwsettings.GlobalSettings.GetKeyFile())
	} else {
		logger.LogInfo("Serving HTTP on: %s", svr.tcpListener.Addr().String())
		svr.server.Serve(svr.tcpListener)
	}
}

/*
CreateHTTPServer creates a new http server on the given port,
using the given protocol and outputing errors to the given error logger. The created
server is returned on success, else (nil, error) is returned
*/
func CreateHTTPServer(port string, proto string, errLogger *log.Logger) (*HTTPServer, error) {
	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/", HandleRequest)

	readTimout, rtErr := time.ParseDuration(mwsettings.GlobalSettings.GetHTTPReadTimeout())
	if rtErr != nil {
		logger.LogError("Could not parse read timeout: %s. defaulting to 1 second",
			mwsettings.GlobalSettings.GetHTTPReadTimeout())
		readTimout, _ = time.ParseDuration("1s")
	}

	writeTimout, wtErr := time.ParseDuration(mwsettings.GlobalSettings.GetHTTPResponseTimeout())
	if wtErr != nil {
		logger.LogError("Could not parse response timeout: %s. defaulting to 1 second",
			mwsettings.GlobalSettings.GetHTTPResponseTimeout())
		writeTimout, _ = time.ParseDuration("1s")
	}

	var srv HTTPServer
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
