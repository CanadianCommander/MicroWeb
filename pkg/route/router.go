package route

import (
	"net/http"
	"net/url"
)

const (
	//DefaultRouterPriority is the default router priority, most routers should use this.
	DefaultRouterPriority = 0
)

/*
Router routes http requests to end points based on application logic.
*/
type Router interface {
	/*
	  Route routes a request via this routers rules. if the return value is false
	  the request will not be propagated down the priority stack
	*/
	Route(req *http.Request, res http.ResponseWriter) bool

	// CanRoute returns true if the router can / wants to route this url.
	CanRoute(routeURL *url.URL) bool

	/*
	  GetPriority gets this Routers priority (0 is highest priority). Higher priority
	  routers run before lower priority routers
	*/
	GetPriority() int
}
