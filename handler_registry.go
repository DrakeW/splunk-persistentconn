package persistentconn

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Handler is a handler function that takes a persistentconn request and returns a response or error
type Handler func(Request) (Response, error)

// NoMatchingHandler is the default handler returned when no matching path
// is found from a request
func NoMatchingHandler(re Request) (Response, error) {
	return Response{
		StatusCode: http.StatusNotFound,
		Body:       "The requested path is not found.",
	}, nil
}

// route represents a registered route that has a corresponding handler
type route struct {
	Pattern *regexp.Regexp
	Handler Handler
	Methods []string
}

// newRoute creates a new route object
func newRoute(pathPattern string, handler Handler, allowedMethods []string) *route {
	re := translatePatternToRegexp(pathPattern)
	return &route{
		Pattern: re,
		Handler: handler,
		Methods: allowedMethods,
	}
}

// translatePatternToRegexp translates a path pattern in the format of "pc1/:<name>/pc2"
// where "pc" stands for path component and can be any arbitary string, and ":name" will be replaced
// based on the request's path. E.g. if request is hitting "pc1/hello/pc2", the param name=hello will
// be stored in the context of the request
func translatePatternToRegexp(pathPattern string) *regexp.Regexp {
	parts := strings.Split(pathPattern, "/")
	regexpStrParts := make([]string, len(parts))
	for idx, p := range parts {
		if strings.HasPrefix(p, ":") {
			p = fmt.Sprintf(`(?P<%s>[\S|^\/]+)`, p[1:])
		}
		regexpStrParts[idx] = p
	}
	regexpStr := strings.Join(regexpStrParts, "/")
	re := regexp.MustCompile(regexpStr)
	return re
}

// handlerRegistry is where all routes are stored
type handlerRegistry struct {
	routes []*route
}

// gethandler gets the handler based on the input reqeust's path info
func (rg *handlerRegistry) getHandler(req Request) Handler {
	handler := NoMatchingHandler
	for _, rt := range rg.routes {
		if matches := rt.Pattern.FindStringSubmatch(req.Path); len(matches) > 0 && contains(rt.Methods, req.Method) {
			// TODO: added matched paramter to request or context or whatever
			return rt.Handler
		}
	}
	return handler
}

// register func registers a path with a handler
func (rg *handlerRegistry) register(path string, handler Handler, allowedMethods []string) {
	route := newRoute(path, handler, allowedMethods)
	rg.routes = append(rg.routes, route)
}
