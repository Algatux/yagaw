package yagaw

import (
	"fmt"
	"iter"
	"maps"
	"net/http"
	"regexp"
	"strings"
)

type RequestHandlerPackage struct {
	Handler   HttpRequestHandler
	ParamList map[int]string
	Params    Params
}
type RequestHandlerMap map[HttpMethod]map[string]RequestHandlerPackage

type Router struct {
	routes RequestHandlerMap
}

// ----------- REQUEST ROUTING -----------
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	debugRequest(rw, req)
	handlerPkg := r.findReqHandler(req)
	response := handlerPkg.Handler(req, handlerPkg.Params)

	for key, header := range response.headers {
		rw.Header().Set(key, header)
	}
	rw.WriteHeader(response.status)
	fmt.Fprint(rw, response.body)
}

// ----------- PATTERN MATCHING -----------
func (r *Router) findReqHandler(req *http.Request) RequestHandlerPackage {
	// Direct match on Method, if not found fast exit to 404
	_, methodFound := r.routes[HttpMethod(req.Method)]
	if !methodFound {
		return RequestHandlerPackage{Handler: routeNotFoundHandler}
	}

	// Direct match on Not parametrized route, if not found fast exit to 404
	handlerPackage, routeFound := r.routes[HttpMethod(req.Method)][req.URL.Path]
	if routeFound {
		return handlerPackage
	}

	// Matching on parametrized routes
	key, matchFound := matchRoutePattern(maps.Keys(r.routes[HttpMethod(req.Method)]), req.URL.Path)
	if matchFound {
		// Extract the parametrized route and retrive parameters values
		handlerPackage := r.routes[HttpMethod(req.Method)][key]
		for i, param := range handlerPackage.ParamList {
			parts := strings.Split(req.URL.Path, "/")
			handlerPackage.Params[param] = parts[i+1]
		}

		return r.routes[HttpMethod(req.Method)][key]
	}

	// Still not found, drop the sponge
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
func (r *Router) RegisterRoute(method HttpMethod, path string, handler HttpRequestHandler) {
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

func routeNotFoundHandler(req *http.Request, _ Params) *HttpResponse {
	return NewHttpResponse(http.StatusNotFound).
		SetHeader("Content-Type", "text/plain").
		SetBody("404 - Page not found")
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
