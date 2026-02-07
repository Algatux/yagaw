package yagaw

import (
	"fmt"
	"iter"
	"maps"
	"net/http"
	"regexp"
)

type ReqHandlerMap map[Method]map[string]ReqHandler
type ReqHandler func(rw http.ResponseWriter, req *http.Request)

type Method string

const (
	GET  Method = "GET"
	POST Method = "POST"
)

type Router struct {
	routes ReqHandlerMap
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	debugRequest(rw, req)
	handler, err := r.findReqHandler(req)
	if err != nil {
		Log.FatalError(err)
	}
	handler(rw, req)
}

func (r *Router) findReqHandler(req *http.Request) (ReqHandler, error) {
	_, methodFound := r.routes[Method(req.Method)]
	if !methodFound {
		return routeNotFoundHandler, nil
	}
	handler, routeFound := r.routes[Method(req.Method)][req.URL.Path]
	if !routeFound {
		handler = routeNotFoundHandler
		path, matchFound := matchRoutePattern(maps.Keys(r.routes[Method(req.Method)]), req.URL.Path)
		if matchFound {
			return r.routes[Method(req.Method)][path], nil
		}
	}
	return handler, nil
}

func matchRoutePattern(keysIter iter.Seq[string], path string) (string, bool) {
	for k := range keysIter {
		re := regexp.MustCompile(fmt.Sprintf("(?i)%s", k))
		record := re.FindString(path)
		if len(record) != 0 {
			return k, true
		}
	}
	return "", false
}

func (r *Router) RegisterRoute(method Method, path string, handler ReqHandler) {
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]ReqHandler)
	}

	re := regexp.MustCompile(`(?i)({[a-z0-9-_]+})`)
	if re.FindStringIndex(path) == nil {
		r.routes[method][path] = handler
	}

	newPath := re.ReplaceAllString(path, `([a-z0-9-_]+)`)
	r.routes[method][newPath] = handler
}

func (r *Router) RegisteredRoutes() *ReqHandlerMap {
	return &r.routes
}

func NewRouter() *Router {
	return &Router{
		routes: make(ReqHandlerMap),
	}
}

func debugRequest(_ http.ResponseWriter, req *http.Request) {
	Log.Debug("Received request:", req.Method, req.URL.Path)
}

func routeNotFoundHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(rw, "404 - Page not found")
}

// var re = regexp.MustCompile(`(?i)/test/({[a-z0-9-_]+})/({[a-z0-9-_]+})/test2/({[a-z0-9-_]+})/tttt/aa/({[a-z0-9-_]+})`)
//     var str = `/test/{p1}/{p2}/test2/{p-3}/tttt/aa/{p_4}`

//     for i, match := range re.FindAllString(str, -1) {
//         fmt.Println(match, "found at index", i)
//     }
