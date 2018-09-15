package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
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

	//build plugins
	buildPlugins("../../testEnvironment/plugins/")

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

/*
buildPlugins builds every plugin found in the directory denoted by pluginDir. This is necessary for testing
because plugins cannot differ in package version from the host application (the web server in this case).
*/
func buildPlugins(pluginDir string) {
	pluginDirList, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		fmt.Printf("Could not read plugin dir: %s with error: %s", pluginDir, err.Error())
		os.Exit(1)
	}

	for _, file := range pluginDirList {
		if file.IsDir() {
			buildPath := path.Join(pluginDir, file.Name())
			buildCmd := exec.Command("go", "build", "-o", path.Join(buildPath, file.Name()+".so"), "-buildmode=plugin", buildPath)
			buildError := buildCmd.Run()
			if buildError != nil {
				fmt.Printf("failed to build plugin at path %s with error %s", buildPath, buildError.Error())
				os.Exit(1)
			}
		}
	}
}

//test getting normal html
func TestHTTPNormalContent(t *testing.T) {
	err := doGet("http://localhost:8080/normal.html", 200, func(b []byte) {
		if matched, _ := regexp.MatchString("<h1> <i> <u> Normal HTML </u> </i> </h1>", string(b)); !matched {
			fmt.Printf("HTML file did not have expected string. It contained: %s", string(b))
			t.Fail()
		}
	})

	if err != nil {
		t.Fail()
	}
}

//test invoking an api plugin
func TestAPIPlugin(t *testing.T) {
	err := doGet("http://localhost:8080/api/", 200, func(b []byte) {
		if matched, _ := regexp.MatchString("HELLO FROM AN API FUNCTION!", string(b)); !matched {
			fmt.Printf("API response did not contain the expected string, it contained: %s\n", string(b))
			t.Fail()
		}
	})

	if err != nil {
		t.Fail()
	}
}

func TestDB(t *testing.T) {
	err := doGet("http://localhost:8080/api/add", 200, func(b []byte) {
		if matched, _ := regexp.MatchString("ADD", string(b)); !matched {
			fmt.Printf("API response did not contain the expected string, it contained: %s\n", string(b))
			t.Fail()
		}
	})

	if err != nil {
		t.Fail()
	}

	err = doGet("http://localhost:8080/api/get", 200, func(b []byte) {
		if matched, _ := regexp.MatchString("hello world", string(b)); !matched {
			fmt.Printf("API response did not contain the expected string, it contained: %s\n", string(b))
			t.Fail()
		}
	})

	if err != nil {
		t.Fail()
	}
}

func doGet(url string, validStatus int, validationFunc func([]byte)) error {
	var client = http.Client{}

	response, err := client.Get(url)
	if err != nil {
		fmt.Printf("Could send GET request with error: %s\n", err.Error())
		return errors.New("could not send GET request")
	}

	if response.StatusCode != validStatus {
		fmt.Printf("Wrong HTTP status on api request http://localhost:8080/api/. expected: %d got: %d\n", validStatus, response.StatusCode)
		return errors.New("Bad status code")
	}

	responseBody, _ := ioutil.ReadAll(response.Body)
	validationFunc(responseBody)
	return nil
}
