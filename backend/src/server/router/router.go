package router

import (
	"net/http"
	"regexp"
)

type Handler func(*Context)

type Router struct {
	middle []Handler
	routes []Route
}

type Route struct {
	regex *regexp.Regexp
	chain []Handler
}

func NewRouter() *Router {
	router := &Router{}
	router.middle = make([]Handler, 0)
	router.routes = make([]Route, 0)
	return router
}

func (r *Router) Use(h Handler) {
	r.middle = append(r.middle, h)
}

func (r *Router) For(path string, handler Handler) {
	chain := make([]Handler, 0)
	chain = append(chain, r.middle...)
	chain = append(chain, handler)

	x := Route{}
	x.regex = routeRegexp(path)
	x.chain = chain
	r.routes = append(r.routes, x)
}

func (r *Router) resolve(path string) *Route {
	for _, route := range r.routes {
		if route.regex.MatchString(path) {
			return &route
		}
	}
	return nil
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	route := r.resolve(path)
	if route == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	context := &Context{}
	context.Req = req
	context.Out = rw
	context.Vars = regexGroups(path, route.regex)
	context.index = -1
	context.chain = route.chain
	context.Next()
}
