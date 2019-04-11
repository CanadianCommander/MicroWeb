package route

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"testing"
)

func TestBasicRoute(t *testing.T) {
	rManager := NewRoutingManager()
	var resultList sort.IntSlice

	// pre routers
	preRouter := NewDumbRouter()

	// normal routers
	normRouterBe := NewDumbRouter()
	normRouter := NewDumbRouter()
	normRouterAf := NewDumbRouter()

	// post routers
	postRouter := NewDumbRouter()

	preRouter.AddFunctionMapping("/foobar.do", func(req *http.Request, res http.ResponseWriter) {
		resultList = append(resultList, 0)
	})

	normRouterBe.SetPriority(-1)
	normRouterBe.AddFunctionMapping("/foobar.do", func(req *http.Request, res http.ResponseWriter) {
		resultList = append(resultList, 1)
	})

	normRouter.AddFunctionMapping("/foobar.do", func(req *http.Request, res http.ResponseWriter) {
		resultList = append(resultList, 2)
	})

	normRouterAf.SetPriority(1)
	normRouterAf.AddFunctionMapping("/foobar.do", func(req *http.Request, res http.ResponseWriter) {
		resultList = append(resultList, 3)
	})

	postRouter.AddFunctionMapping("/foobar.do", func(req *http.Request, res http.ResponseWriter) {
		resultList = append(resultList, 4)
	})

	rManager.AddPreRouter(preRouter)
	rManager.AddNormalRouter(normRouter)
	rManager.AddNormalRouter(normRouterBe)
	rManager.AddNormalRouter(normRouterAf)
	rManager.AddPostRouter(postRouter)

	req := &http.Request{}
	req.URL, _ = url.Parse("/foobar.do")
	err := rManager.RouteRequest(req, nil)
	if err != nil {
		fmt.Printf("Error while routing: %s\n", err.Error())
		t.Fail()
		return
	}

	if !sort.IsSorted(resultList) {
		fmt.Print("Routers not tirggerd in correct order\n")
		t.Fail()
	}
}

type MethodRouterTestTarget struct {
	Msg string
}

func (mrtt *MethodRouterTestTarget) CallMe(req *http.Request, res http.ResponseWriter) bool {
	mrtt.Msg = "win"
	return true
}

type myResponseWriter struct {
}

func (mrw *myResponseWriter) Header() http.Header {
	return nil
}

func (mrw *myResponseWriter) Write(data []byte) (int, error) {
	return 0, nil
}

func (mrw *myResponseWriter) WriteHeader(statusCode int) {

}

func TestMethodRouter(t *testing.T) {
	rManager := NewRoutingManager()

	mrtt := &MethodRouterTestTarget{"fail"}
	methodRouter := NewMethodRouter(mrtt, "/foobar.do")
	rManager.AddRouter(methodRouter, NormalRouter)

	req := &http.Request{}
	req.Form = make(map[string][]string)
	req.Form.Add("method", "CallMe")
	req.URL, _ = url.Parse("/foobar.do")

	err := rManager.RouteRequest(req, &myResponseWriter{})
	if err != nil {
		fmt.Printf("Error while routing: %s\n", err.Error())
		t.Fail()
		return
	}

	if mrtt.Msg != "win" {
		fmt.Print("the callMe() method was not called!\n")
		t.Fail()
	}
}
