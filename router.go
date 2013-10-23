package httpd

import (
	"bytes"
	"github.com/jiorry/gos/websock"
	"github.com/jiorry/libs/log"
	"html"
	"reflect"
	"regexp"
	"strings"
)

var routes []*Route = make([]*Route, 0)
var apiRoutes []*Route = make([]*Route, 0)
var upRoutes []*Route = make([]*Route, 0)
var wsRoutes []*Route = make([]*Route, 0)

type Route struct {
	ClassType     reflect.Type
	Rule          []byte
	Pattern       *regexp.Regexp // for matching the url path
	Keys          []string
	beforeFilters []func(ctx *Context) bool
	afterFilters  []func(ctx *Context) bool
}

type RouteMatched struct {
	ClassType reflect.Type
	Params    map[string]string
}

func AddRoute(rule string, clas interface{}) *Route {
	switch rule {
	case "/upload":
		log.App.Alert("/upload is used for default upload router")
	case "/ping":
		log.App.Alert("/ping is used for default ping router")
	case "/api":
		log.App.Alert("/api is used for default api router")
	case "/ws":
		log.App.Alert("/ws is used for default websocket router")
	}
	return addRouteTo(rule, clas, 0)
}

func MatchRoute(path []byte) *RouteMatched {
	return matchRoute(path, 0)
}

func AddWebSocketRoute(rule string, clas interface{}) {
	httpServer.EnableWebSocket = true
	r := addRouteTo("/ws"+rule, clas, 3)
	s := websock.NewServer(r.ClassType.String())
	go s.Start()
}

func MatchWebSocketRoute(path []byte) *RouteMatched {
	return searchPathFrom(path, wsRoutes)
}

func AddWebApiRoute(rule string, clas interface{}) {
	httpServer.EnableApi = true
	addRouteTo("/api"+rule, clas, 1)
}

func MatchWebApiRoute(path []byte) *RouteMatched {
	return searchPathFrom(path, apiRoutes)
}

func AddFileUploadRoute(rule string, clas interface{}) {
	httpServer.EnableUpload = true
	addRouteTo("/upload"+rule, clas, 2)
}

func MatchFileuploadRoute(path []byte) *RouteMatched {
	return searchPathFrom(path, upRoutes)
}

func addRouteTo(rule string, clas interface{}, itype int) *Route {
	var regPath *regexp.Regexp
	var keys []string

	if strings.ContainsAny(rule, ":") {
		regReplace, _ := regexp.Compile(":\\w+")
		matched := regReplace.FindAllString(rule, -1)
		// prepare keys
		keys = []string{}
		for _, value := range matched {
			keys = append(keys, value[1:len(value)])
		}

		//prepare regexp pattern
		p := regReplace.ReplaceAllString("^"+rule+"$", "(\\w+)")
		regPath, _ = regexp.Compile(p)
	} else {
		if strings.HasSuffix(rule, "/") {
			rule = strings.TrimSuffix(rule, "/")
		}
	}

	//prepare class struct
	typ := reflect.TypeOf(clas)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	r := &Route{
		Keys:      keys,
		Rule:      []byte(rule),
		Pattern:   regPath,
		ClassType: typ}
	switch itype {
	case 0:
		routes = append(routes, r)
	case 1:
		apiRoutes = append(apiRoutes, r)
	case 2:
		upRoutes = append(upRoutes, r)
	case 3:
		wsRoutes = append(wsRoutes, r)
	}
	return r
}

func searchPathFrom(path []byte, fromRoutes []*Route) *RouteMatched {
	if bytes.HasSuffix(path, B_SLASH) {
		path = bytes.TrimSuffix(path, B_SLASH)
	}
	for _, route := range fromRoutes {
		if bytes.Equal(path, route.Rule) {
			return &RouteMatched{ClassType: route.ClassType, Params: nil}
		}
	}
	return nil
}

func matchRoute(path []byte, itype int) *RouteMatched {
	var route *Route
	var items []*Route
	switch itype {
	case 3:
		items = wsRoutes
	default:
		items = routes
	}

	for _, route = range items {
		if route.Pattern == nil {
			// fix route
			if bytes.HasSuffix(path, B_SLASH) {
				path = bytes.TrimSuffix(path, B_SLASH)
			}
			if bytes.Equal(path, route.Rule) {
				return &RouteMatched{ClassType: route.ClassType, Params: nil}
			}
		} else {
			// regexp route
			if matched := route.Pattern.FindAllSubmatch(path, -1); matched != nil {
				params := make(map[string]string)
				i := 0
				for _, value := range route.Keys {
					i++
					params[value] = html.EscapeString(string(matched[0][i]))
				}
				return &RouteMatched{ClassType: route.ClassType, Params: params}
			}
		}
	}

	return nil
}
