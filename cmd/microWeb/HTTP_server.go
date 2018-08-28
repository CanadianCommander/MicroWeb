package main

import (
	"log"
	"microWeb/pkg/logger"
	"net"
	"net/http"
	"time"
)

type HttpServer struct {
	server      *http.Server
	tcpListener net.Listener
}

func (svr *HttpServer) ServeHTTP() {
	logger.LogInfo("Serving HTTP on: %s", svr.tcpListener.Addr().String())
	svr.server.Serve(svr.tcpListener)
}

func CreateHTTPServer(port string, proto string, handlerList []*http.Handler, errLogger *log.Logger) (*HttpServer, int) {
	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/", HandleTemplateRequest)

	readTimout, rtErr := time.ParseDuration(globalSettings.httpReadTimeout)
	if rtErr != nil {
		logger.LogError("Could not parse read timeout: %s. defaulting to 1 second",
			globalSettings.httpReadTimeout)
		readTimout, _ = time.ParseDuration("1s")
	}

	writeTimout, wtErr := time.ParseDuration(globalSettings.httpResponseTimeout)
	if wtErr != nil {
		logger.LogError("Could not parse response timeout: %s. defaulting to 1 second",
			globalSettings.httpResponseTimeout)
		writeTimout, _ = time.ParseDuration("1s")
	}

	var srv HttpServer
	srv.server = &http.Server{
		Addr:         port,
		Handler:      srvMux,
		ErrorLog:     errLogger,
		ReadTimeout:  readTimout,
		WriteTimeout: writeTimout}
	srv.tcpListener, _ = net.Listen(proto, port)

	return &srv, 0
}
