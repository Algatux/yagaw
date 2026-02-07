package yagaw

import (
	"fmt"
	"iter"
	"maps"
	"net/http"
	"regexp"
	"strings"
)

type Params map[string]any
type RequestHandlerPackage struct {
	Handler   RequestHandler
	ParamList map[int]string
	Params    Params
}
type RequestHandlerMap map[HttpRequestMethod]map[string]RequestHandlerPackage
type RequestHandler func(rw http.ResponseWriter, req *http.Request, params Params)

type HttpRequestMethod string

const (
	GET     HttpRequestMethod = `GET`
	HEAD    HttpRequestMethod = `HEAD`
	OPTIONS HttpRequestMethod = `OPTIONS`
	TRACE   HttpRequestMethod = `TRACE`
	PUT     HttpRequestMethod = `PUT`
	DELETE  HttpRequestMethod = `DELETE`
	POST    HttpRequestMethod = `POST`
	PATCH   HttpRequestMethod = `PATCH`
	CONNECT HttpRequestMethod = `CONNECT`
)

type Router struct {
	routes RequestHandlerMap
}

// ----------- REQUEST ROUTING -----------
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	debugRequest(rw, req)
	handler := r.findReqHandler(req)
	handler.Handler(rw, req, handler.Params)
}

// ----------- PATTERN MATCHING -----------
func (r *Router) findReqHandler(req *http.Request) RequestHandlerPackage {
	_, methodFound := r.routes[HttpRequestMethod(req.Method)]
	if !methodFound {
		return RequestHandlerPackage{Handler: routeNotFoundHandler}
	}

	handlerPackage, routeFound := r.routes[HttpRequestMethod(req.Method)][req.URL.Path]

	if routeFound {
		return handlerPackage
	}

	if !routeFound {
		key, matchFound := matchRoutePattern(maps.Keys(r.routes[HttpRequestMethod(req.Method)]), req.URL.Path)
		if matchFound {
			handlerPackage := r.routes[HttpRequestMethod(req.Method)][key]
			for i, param := range handlerPackage.ParamList {
				parts := strings.Split(req.URL.Path, "/")
				handlerPackage.Params[param] = parts[i+1]
			}

			return r.routes[HttpRequestMethod(req.Method)][key]
		}
	}

	return RequestHandlerPackage{Handler: routeNotFoundHandler}
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

// ----------- ROUTE REGISTRATION -----------
func (r *Router) RegisterRoute(method HttpRequestMethod, path string, handler RequestHandler) {
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]RequestHandlerPackage)
	}

	type paramSearch struct {
		start int
		end   int
		pos   int
		name  string
	}

	// Searching for url parameters patterns
	paramList := []paramSearch{}
	pathDepth := -1
	found := false
	foundAt := 0

	paramNameBuilder := strings.Builder{}
	for i, c := range path {
		switch c {
		case '/':
			pathDepth++
		case '{':
			found = true
			foundAt = i
		case '}':
			found = false
			paramList = append(paramList, paramSearch{
				start: foundAt,
				end:   i,
				pos:   pathDepth,
				name:  paramNameBuilder.String(),
			})
			paramNameBuilder.Reset()
		}
		if found && c != '{' {
			paramNameBuilder.WriteRune(c)
		}
	}

	pathBuilder := strings.Builder{}
	lastPos := 0
	reqParamList := map[int]string{}

	for _, param := range paramList {
		pathBuilder.WriteString(path[lastPos:param.start])
		pathBuilder.WriteString("([a-z0-9-_]+)")
		lastPos = param.end + 1
		reqParamList[param.pos] = param.name
	}
	pathBuilder.WriteString(path[lastPos:])
	newPath := pathBuilder.String()
	if len(paramList) > 0 {
		newPath = "^" + newPath + "$"
	}

	r.routes[method][newPath] = RequestHandlerPackage{Handler: handler, ParamList: reqParamList, Params: make(Params)}
}

func (r *Router) RegisteredRoutes() *RequestHandlerMap {
	return &r.routes
}

// ----------- DEFALUT HANDLERS -----------

func routeNotFoundHandler(rw http.ResponseWriter, req *http.Request, _ Params) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(rw, "404 - Page not found")
}

// ----------- HELPERS -----------
func debugRequest(_ http.ResponseWriter, req *http.Request) {
	Log.Debug("Received request:", req.Method, req.URL.Path)
}

// ----------- CONSTRUCTOR -----------
func NewRouter() *Router {
	return &Router{
		routes: make(RequestHandlerMap),
	}
}
