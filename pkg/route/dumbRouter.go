package route

import (
	"net/http"
	"net/url"
)

/*
DumbRouter routes http requests to end points based on a supplied url -> function map.
*/
type DumbRouter struct {
	routes   map[string]func(req *http.Request, res http.ResponseWriter)
	priority int
}

//NewDumbRouter creates a new DumbRouter.
func NewDumbRouter() *DumbRouter {
	return &DumbRouter{make(map[string]func(req *http.Request, res http.ResponseWriter)), DefaultRouterPriority}
}

/*
Route calls the function associated with the request url
*/
func (dumbRouter *DumbRouter) Route(req *http.Request, res http.ResponseWriter) bool {
	routFunc, bOk := dumbRouter.routes[req.URL.String()]
	if bOk {
		routFunc(req, res)
	}

	return true
}

/*
CanRoute returns true if the url-function map has an entry for the url
*/
func (dumbRouter *DumbRouter) CanRoute(routeURL *url.URL) bool {
	_, bOk := dumbRouter.routes[routeURL.String()]
	return bOk
}

/*
GetPriority returns this routers priority
*/
func (dumbRouter *DumbRouter) GetPriority() int {
	return dumbRouter.priority
}

/*
SetPriority sets the priority of this router.
*/
func (dumbRouter *DumbRouter) SetPriority(pri int) {
	dumbRouter.priority = pri
}

/*
AddFunctionMapping adds a new url -> function mapping to the router
*/
func (dumbRouter *DumbRouter) AddFunctionMapping(funcURL string, function func(req *http.Request, res http.ResponseWriter)) {
	dumbRouter.routes[funcURL] = function
}

/*
RemoveFunctionMapping removes the function mapping for funcURL
*/
func (dumbRouter *DumbRouter) RemoveFunctionMapping(funcURL string) {
	delete(dumbRouter.routes, funcURL)
}
