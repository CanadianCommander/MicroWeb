package route

import (
	"net/http"
	"runtime/debug"
	"sort"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

const (
	//PreRouter is a router that occures before normal and post routers
	PreRouter = iota
	//NormalRouter is a router that occures after PreRouters but before PostRouters
	NormalRouter
	//PostRouter is a router that occures after all Pre and Normal routers.
	PostRouter
)

/*
RoutingManager manages routers and routing operation. simply add routers to the manager then
feed in http requests and the manager will pass the message on to the correct routers.
*/
type RoutingManager struct {
	preRouters  []Router
	routers     []Router
	postRouters []Router
}

//NewRoutingManager creates a new routing Manager.
func NewRoutingManager() *RoutingManager {
	return &RoutingManager{make([]Router, 0), make([]Router, 0), make([]Router, 0)}
}

/*
RouteRequest routes an http request through the registered routers
*/
func (manager *RoutingManager) RouteRequest(req *http.Request, res http.ResponseWriter) error {
	//catch any panics
	defer func() {
		if r := recover(); r != nil {
			logger.LogError("Caught panic while routing: %s \n", string(debug.Stack()))
		}
	}()

	// sort routers based on priority
	targetSlice := manager.preRouters
	routerSortFunc := func(i, j int) bool {
		return targetSlice[i].GetPriority() < targetSlice[j].GetPriority()
	}
	targetSlice = manager.preRouters
	sort.Slice(manager.preRouters, routerSortFunc)
	targetSlice = manager.routers
	sort.Slice(manager.routers, routerSortFunc)
	targetSlice = manager.postRouters
	sort.Slice(manager.postRouters, routerSortFunc)

	//go through routers
	routerSet := [][]Router{manager.preRouters, manager.routers, manager.postRouters}
	for _, set := range routerSet {
		for _, r := range set {
			if r.CanRoute(req.URL) {
				if !r.Route(req, res) {
					// cancel further propagation
					return nil
				}
			}
		}
	}

	return nil
}

/*
AddPreRouter adds a pre router to the manager.
A pre router, is a router that occures before normal and post routers.
*/
func (manager *RoutingManager) AddPreRouter(router Router) {
	manager.AddRouter(router, PreRouter)
}

/*
AddNormalRouter  adds a normal router to the manager.
A normal router, is a router that occures after PreRouters but before PostRouters.
*/
func (manager *RoutingManager) AddNormalRouter(router Router) {
	manager.AddRouter(router, NormalRouter)
}

/*
AddPostRouter adds a post router to the manager.
A post router, is a router that occures affter all Pre and Normal routers.
*/
func (manager *RoutingManager) AddPostRouter(router Router) {
	manager.AddRouter(router, PostRouter)
}

/*
AddRouter adds a router to the manager with the given router type.
*/
func (manager *RoutingManager) AddRouter(router Router, routerType int) {
	switch routerType {
	case PreRouter:
		manager.preRouters = append(manager.preRouters, router)
	case NormalRouter:
		manager.routers = append(manager.routers, router)
	case PostRouter:
		manager.postRouters = append(manager.postRouters, router)
	}
}
