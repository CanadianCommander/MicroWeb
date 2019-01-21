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
	server          *http.Server
	tcpListener     net.Listener
	redirectServers []HTTPServer
}

/*
ServeHTTP start the http server. BLOCKS until server exits
*/
func (svr *HTTPServer) ServeHTTP() {
	// start redirect servers
	for _, rServer := range svr.redirectServers {
		logger.LogInfo("Starting redirect server on port: %s", rServer.tcpListener.Addr().String())

		rServerCpy := rServer
		go func() {
			err := rServerCpy.server.Serve(rServerCpy.tcpListener)
			if err != nil {
				logger.LogError("Failed to start redirect server on port %s with error: %s",
					rServerCpy.tcpListener.Addr().String(), err.Error())
			}
		}()
	}

	// start primary server
	if mwsettings.GetSettingBool("tls/enableTLS") {
		logger.LogInfo("Serving HTTPS on: %s", svr.tcpListener.Addr().String())
		err := svr.server.ServeTLS(svr.tcpListener, mwsettings.GetSettingString("tls/certFile"), mwsettings.GetSettingString("tls/keyFile"))
		if err != nil {
			logger.LogError("Could not server HTTPS with error: %s", err.Error())
		}
	} else {
		logger.LogInfo("Serving HTTP on: %s", svr.tcpListener.Addr().String())
		err := svr.server.Serve(svr.tcpListener)
		if err != nil {
			logger.LogError("Could not server HTTP with error: %s", err.Error())
		}
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

	readTimout, rtErr := time.ParseDuration(mwsettings.GetSettingString("tune/httpReadTimeout"))
	if rtErr != nil {
		logger.LogError("Could not parse read timeout: %s. defaulting to 1 second",
			mwsettings.GetSettingString("tune/httpReadTimeout"))
		readTimout, _ = time.ParseDuration("1s")
	}

	writeTimout, wtErr := time.ParseDuration(mwsettings.GetSettingString("tune/httpResponseTimeout"))
	if wtErr != nil {
		logger.LogError("Could not parse response timeout: %s. defaulting to 1 second",
			mwsettings.GetSettingString("tune/httpResponseTimeout"))
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

	CreateRedirectServers(port, proto, errLogger, writeTimout, readTimout, &srv)

	return &srv, nil
}

// CreateRedirectServers creates zero - N redirect servers. These servers simply redirect
// HTTP requests to the URL: "redirectURL" + "port" + "what ever the original request was for"
func CreateRedirectServers(port string, proto string, errLogger *log.Logger, writeTimeout time.Duration, readTimeout time.Duration, server *HTTPServer) {
	rdirects := mwsettings.GetSetting("general/redirectPorts")
	rURL := mwsettings.GetSettingString("general/redirectURL")
	redirectMux := http.NewServeMux()
	redirectMux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Location", rURL+port+req.URL.String())
		res.WriteHeader(301)
	})

	if rdirects != nil {
		rdirects, bOk := rdirects.([]interface{})
		if !bOk {
			logger.LogWarning("Setting \"general/redirectPorts\" has incorrect value. should be list of strings.")
		} else {
			redirectServers := make([]HTTPServer, len(rdirects))
			for i, redirect := range rdirects {
				redirectPort := redirect.(string)

				redirectServers[i] = HTTPServer{}
				redirectServers[i].server = &http.Server{
					Addr:         redirectPort,
					Handler:      redirectMux,
					ErrorLog:     errLogger,
					ReadTimeout:  readTimeout,
					WriteTimeout: writeTimeout}

				var err error
				redirectServers[i].tcpListener, err = net.Listen(proto, redirectPort)
				if err != nil {
					logger.LogError("Failed to create redirect server on port %s with error: %s", redirectPort, err.Error())
				}
			}
			server.redirectServers = redirectServers
		}
	}
}
