package httpd

import (
	"github.com/jiorry/gos/log"
	"html"
	"reflect"
	"regexp"
	"strings"
)

var (
	routes   []*Route
	wsRoutes []*Route
	upRoutes []*Route
)

func init() {
	routes = make([]*Route, 0)
	wsRoutes = make([]*Route, 0)
	upRoutes = make([]*Route, 0)
}

type Route struct {
	ClassType reflect.Type
	Rule      string
	Pattern   *regexp.Regexp // for matching the url path
	Keys      []string
}

type RouteMatched struct {
	ClassType reflect.Type
	Params    map[string]string
}

func AddRoute(rule string, clas interface{}) {
	switch rule {
	case "/upload":
		log.App.Alert("/upload is used for default upload router")
	case "/ping":
		log.App.Alert("/ping is used for default ping router")
	case "/api":
		log.App.Alert("/api is used for default api router")
	}
	addRouteTo(rule, clas, 0)
}

func MatchRoute(path string) *RouteMatched {
	return matchRoute(path)
}

func AddWebServiceRoute(rule string, clas interface{}) {
	if rule == "" {
		rule = "/"
	}
	addRouteTo("/api"+rule, clas, 1)
}

func MatchWebServiceRoute(path string) *RouteMatched {
	return searchPathFrom(path, wsRoutes)
}

func AddFileUploadRoute(rule string, clas interface{}) {
	if rule == "" {
		rule = "/"
	}
	addRouteTo("/upload"+rule, clas, 2)
}

func MatchFileuploadRoute(path string) *RouteMatched {
	return searchPathFrom(path, upRoutes)
}

func addRouteTo(rule string, clas interface{}, itype int) {
	var regPath *regexp.Regexp
	var keys []string

	if strings.ContainsAny(rule, "[") {
		regReplace, _ := regexp.Compile("\\[\\w+\\]")
		matched := regReplace.FindAllString(rule, -1)
		// prepare keys
		keys = []string{}
		for _, value := range matched {
			keys = append(keys, value[1:len(value)-1])
		}

		//prepare regexp pattern
		p := regReplace.ReplaceAllString(rule, "(\\w+)")
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
		Rule:      rule,
		Pattern:   regPath,
		ClassType: typ}
	switch itype {
	case 0:
		routes = append(routes, r)
	case 1:
		wsRoutes = append(wsRoutes, r)
	case 2:
		upRoutes = append(upRoutes, r)
	}
}

func searchPathFrom(path string, fromRoutes []*Route) *RouteMatched {
	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	for _, route := range fromRoutes {
		if path == route.Rule {
			return &RouteMatched{ClassType: route.ClassType, Params: nil}
		}
	}
	return nil
}

func matchRoute(path string) *RouteMatched {
	var route *Route

	for _, route = range routes {
		if route.Pattern == nil {
			// string route
			if strings.HasSuffix(path, "/") {
				path = strings.TrimSuffix(path, "/")
			}
			if path == route.Rule {
				return &RouteMatched{ClassType: route.ClassType, Params: nil}
			}
		} else {
			// regexp route
			if matched := route.Pattern.FindAllStringSubmatch(path, -1); matched != nil {
				params := map[string]string{}
				i := 0
				for _, value := range route.Keys {
					i++
					params[value] = html.EscapeString(matched[0][i])
				}
				return &RouteMatched{ClassType: route.ClassType, Params: params}
			}
		}
	}

	return nil
}
