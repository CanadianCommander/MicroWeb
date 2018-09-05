package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

//TestMain sets up the testing environment
func TestMain(t *testing.M) {
	//build test environment and start server
	serverContext, ctxCancel := context.WithCancel(context.Background())
	serverCmd := buildNStartMicroWeb(&serverContext)

	//give the web server a bit of time to startup
	time.Sleep(50 * time.Millisecond)

	returnCode := t.Run()

	//stop web server
	ctxCancel()
	serverCmd.Wait()

	//WARN defer will not work in this function!
	os.Exit(returnCode)
}

/*
Build micro web project and start the web server for testing.
Returns the exec command handle for the command used to create the server. One
should use this to wait on command termination before exiting the process as this can cause
the command to continue running throwing off future tests. This leak can even happen if you cancel
the context and call os.Exit() without waiting for the command to complete!
*/
func buildNStartMicroWeb(srvCtx *context.Context) *exec.Cmd {
	//build the project
	buildCmd := exec.Command("go", "build", "-o", "/tmp/microWeb.a", ".")
	buildErr := buildCmd.Run()
	if buildErr != nil {
		fmt.Printf("ERROR: Failed to compile MicroWeb! with error: %s\n", buildErr.Error())
		os.Exit(1)
	}

	//copy test file
	copyTestEnv := exec.Command("cp", "-r", os.Getenv("GOPATH")+"/src/github.com/CanadianCommander/MicroWeb/testEnvironment", "/tmp/")
	cpyErr := copyTestEnv.Run()
	if cpyErr != nil {
		fmt.Printf("ERROR: Could not deploy testing environment to /tmp/ with error: %s\n", cpyErr.Error())
		os.Exit(1)
	}

	//run server
	runServer := exec.CommandContext(*srvCtx, "/tmp/microWeb.a", "-c", "/tmp/testEnvironment/test.cfg.json")
	runServer.Start()

	return runServer
}
