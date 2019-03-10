package session

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

type helloObject struct {
	Msg string
}

func (hello *helloObject) MarshalBinary() (data []byte, err error) {
	return []byte(hello.Msg), nil
}

func (hello *helloObject) UnmarshalBinary(data []byte) error {
	hello.Msg = string(data)
	return nil
}

func (hello *helloObject) GetIdentifier() string {
	return "HELLO_WORLD"
}

func TestSessionRawBuffer(t *testing.T) {
	logger.LogToStd(logger.VDebug)

	keyString := "don't tell any one my key!"

	// create buffer from session
	out, err := NewSession(keyString)
	if err != nil {
		fmt.Printf("Failed to create session with error %s", err.Error())
		t.Fail()
		return
	}

	out.Add(&helloObject{"hello world"})

	sessionBuffer, err := out.GetBuffer()
	if err != nil {
		fmt.Printf("session buffer creation failed with error: %s", err.Error())
		t.Fail()
		return
	}

	// reconstruct session from buffer
	anotherHello := helloObject{""}
	in, err := NewSession(keyString)
	if err != nil {
		fmt.Printf("Failed to create session with error %s", err.Error())
		t.Fail()
		return
	}
	in.Add(&anotherHello)

	err = in.FromBuffer(sessionBuffer)
	if err != nil {
		fmt.Printf("re-creation from session buffer failed, with error: %s ", err.Error())
		t.Fail()
		return
	}

	if anotherHello.Msg != "hello world" {
		fmt.Printf("Unmarshal failure! got: [%s] expecting: [hello world] ", anotherHello.Msg)
		t.Fail()
	}
}

func TestSessionHttp(t *testing.T) {
	helloString := "I like cookies"
	cookieName := "myCookie"
	key := "123 ABC"

	// start webserver
	go func() {
		mySession, _ := NewSession(key)
		mySession.Add(&helloObject{helloString})

		http.HandleFunc("/getSession", func(w http.ResponseWriter, r *http.Request) {
			Save(cookieName, mySession, w)
			w.WriteHeader(200)
			w.Write([]byte("hi"))
		})

		http.HandleFunc("/checkSession", func(w http.ResponseWriter, r *http.Request) {
			checkSession, _ := NewSession(key)
			hello := helloObject{""}
			checkSession.Add(&hello)

			err := Load(cookieName, checkSession, r)
			if err != nil {
				fmt.Printf("session decode error: %s\n", err.Error())
				w.WriteHeader(418)
				return
			}

			if hello.Msg == helloString {
				w.WriteHeader(200)
				w.Write([]byte("check ok"))
			} else {
				fmt.Printf("incorrect msg: %s expecting: %s\n", hello.Msg, helloString)
				w.WriteHeader(418)
			}
		})

		http.ListenAndServe(":9042", nil)
	}()

	//build client obj
	cookieJar, _ := cookiejar.New(nil)
	client := http.Client{
		Jar: cookieJar,
	}

	// wait for http server to come up! (max 1 second)
	for i := 0; i < 101; i++ {
		_, err := client.Get("http://localhost:9042/getSession")
		if err == nil {
			break
		}
		if i >= 100 {
			fmt.Print("server failed to start\n")
			t.Fail()
			return
		}
		time.Sleep(10 * time.Millisecond) // wait 10 more ms
	}

	// check that the cookie cad be decoded.
	resp, _ := client.Get("http://localhost:9042/checkSession")
	if resp.StatusCode != 200 {
		fmt.Printf("Wrong status [%s] expecting 200\n", resp.Status)
		t.Fail()
	}
}
