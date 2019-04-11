package route

import (
	"net/http"
	"net/url"
	"reflect"
)

/*
MethodRouter routes requests to methods on a receiver interface (target).
The method to which the request is routed is selected via a http parameter "method".
*/
type MethodRouter struct {
	target          interface{}
	priority        int
	methodParameter string
	URL             string
}

/*
NewMethodRouter constructs a new method router on the target struct at the given url.
*/
func NewMethodRouter(target interface{}, targetURL string) *MethodRouter {
	methodR := MethodRouter{}
	methodR.SetTarget(target)
	methodR.SetURL(targetURL)
	methodR.SetMethodParameter("method")
	return &methodR
}

/*
Route routes the incoming request to the requested method of the target class
*/
func (mrouter *MethodRouter) Route(req *http.Request, res http.ResponseWriter) bool {
	req.ParseForm()

	methodName, bOk := req.Form[mrouter.methodParameter]
	if bOk && len(methodName) > 0 {
		targetValue := reflect.ValueOf(mrouter.target)
		targetMethod := targetValue.MethodByName(methodName[0])

		if targetMethod.IsValid() {
			methodArgs := []reflect.Value{reflect.ValueOf(req), reflect.ValueOf(res)}
			results := targetMethod.Call(methodArgs)

			if len(results) > 0 {
				return results[0].Bool()
			}
		}
	}
	return true
}

/*
CanRoute checks if the url is exactly the same as the routers url
*/
func (mrouter *MethodRouter) CanRoute(routeURL *url.URL) bool {
	return routeURL.String() == mrouter.URL
}

/*
GetPriority returns routers priority setting
*/
func (mrouter *MethodRouter) GetPriority() int {
	return mrouter.priority
}

/*
SetPriority sets the routers priority
*/
func (mrouter *MethodRouter) SetPriority(priority int) {
	mrouter.priority = priority
}

/*
GetTarget returns the target interface of the method router
*/
func (mrouter *MethodRouter) GetTarget() interface{} {
	return mrouter.target
}

/*
SetTarget sets the target interface of the method router
*/
func (mrouter *MethodRouter) SetTarget(target interface{}) {
	mrouter.target = target
}

/*
GetURL returns the URL of this method router
*/
func (mrouter *MethodRouter) GetURL() interface{} {
	return mrouter.URL
}

/*
SetURL sets the URL of this method router
*/
func (mrouter *MethodRouter) SetURL(URL string) {
	mrouter.URL = URL
}

/*
SetMethodParameter sets the method parameter name which indicates what method to call on
the router target.
*/
func (mrouter *MethodRouter) SetMethodParameter(methodParam string) {
	mrouter.methodParameter = methodParam
}

/*
GetMethodParameter returns the method parameter name which indicates what method to call on
the router target.
*/
func (mrouter *MethodRouter) GetMethodParameter() string {
	return mrouter.methodParameter
}
