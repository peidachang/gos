package httpd

import (
	"encoding/json"
	"fmt"
	"github.com/jiorry/gos/cache"
	"github.com/jiorry/gos/db"
	"github.com/jiorry/gos/log"
	"io"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"reflect"
	"strings"
)

type HttpServer struct {
	Addr       string
	Port       int
	EnableGzip bool
	PprofOn    bool

	StaticDir string
	CacheDir  string
	Timestamp string

	EnablePing   bool
	EnableUpload bool
	EnableApi    bool

	UseFcgi   bool
	lenStatic int
}

var (
	httpServer *HttpServer
	HomeUrl    string
	StaticUrl  string
	Theme      *ThemeData
	RunMode    string //"dev" or "prod"
)

func init() {
	httpServer = &HttpServer{
		Addr:         "",
		Port:         8800,
		StaticDir:    "static",
		CacheDir:     "static",
		PprofOn:      false,
		EnableGzip:   false,
		EnablePing:   false,
		EnableUpload: true,
		EnableApi:    true,
		UseFcgi:      false,
	}
	Theme = &ThemeData{}
	HomeUrl = "/"
	StaticUrl = "/"

	if _, err := os.Stat("var"); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("var", os.ModeDir)
		}
	}

	if _, err := os.Stat("var/log"); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("var/log", os.ModeDir)
		}
	}

	if _, err := os.Stat("var/cache"); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("var/cache", os.ModeDir)
		}
	}
}

// Initialize server
// The config file (code/app.conf) will be loaded.
func Init() {
	conf := LoadConfig("code/")
	appConf := conf["app"]
	httpConf := conf["http"]
	RunMode = conf.GetRunMode()

	names := []string{}
	log.Init("var/log/", names, RunMode)

	if appConf.IsSet("home_url") {
		HomeUrl = appConf.GetString("home_url")
	}
	if appConf.IsSet("static_url") {
		StaticUrl = appConf.GetString("static_url")
	}

	if httpConf.IsSet("static") {
		httpServer.StaticDir = httpConf.GetString("static")
	}

	if httpConf.IsSet("cache_dir") {
		httpServer.CacheDir = httpConf.GetString("cache_dir")
	}

	if httpConf.IsSet("timestamp") {
		httpServer.Timestamp = "?ts=" + httpConf.GetString("timestamp")
	}

	if RunMode == "dev" {
		log.Level = 10
		fmt.Println("Server is run in Development Mode")
	} else {
		log.Level = conf["log"].GetInt("level")
		fmt.Println("Server is run in Production Mode")
	}

	if httpConf.IsSet("port") {
		httpServer.Addr = httpConf.GetString("addr")
	}

	if httpConf.IsSet("port") {
		httpServer.Port = httpConf.GetInt("port")
	}

	httpServer.UseFcgi = httpConf.GetBool("fcgi")
	httpServer.lenStatic = len(httpServer.StaticDir)
	httpServer.PprofOn = httpConf.GetBool("pprof")
	httpServer.EnableGzip = httpConf.GetBool("gzip")
	httpServer.EnableApi = httpConf.GetBool("enable_api")
	httpServer.EnableUpload = httpConf.GetBool("enable_upload")
	httpServer.EnablePing = httpConf.GetBool("enable_ping")

	if conf.IsSet("theme") {
		ct := conf["theme"]
		if ct.IsSet("template") {
			Theme.Template = ct["template"]
		}
		if ct.IsSet("css") {
			Theme.Css = ct["css"]
		}
		if ct.IsSet("js") {
			Theme.Js = ct["js"]
		}
	}

	if conf.IsSet("db") {
		db.New("app", conf["db"])
	}
	if conf.IsSet("cache") {
		cache.Init(conf["cache"])
	}
}

// Start server
// You can set config [fcgi] option to true if you want run server under fastcgi mode.
// [http]
// fcgi=true
func Start() {
	addHander()

	addr := fmt.Sprintf("%s:%d", httpServer.Addr, httpServer.Port)
	if httpServer.UseFcgi {
		startFcig()
		log.App.Write("fastcgi start at ", addr)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.App.Fatalln(err)
		}
		fcgi.Serve(l, http.DefaultServeMux)
	} else {
		log.App.Write("server start at ", addr)
		http.ListenAndServe(addr, http.DefaultServeMux)
	}
}

func addHander() {
	if httpServer.PprofOn {
		http.HandleFunc("/debug/pprof", nil)
	}

	if httpServer.EnablePing {
		http.HandleFunc("/ping", pingHander)
	}

	if httpServer.EnableApi {
		http.HandleFunc("/api/", webserviceHander)
	}

	if httpServer.EnableUpload {
		http.HandleFunc("/upload/", uploadHander)
	}

	http.HandleFunc("/", serveHTTPHander)
}

func startFcig() {

}

func pingHander(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("ok"))
}

// func staticHander(rw http.ResponseWriter, req *http.Request) {
// 	log.App.Info("static:", req.URL.Path)
// 	http.ServeFile(rw, req, "."+req.URL.Path)
// }

func uploadHander(rw http.ResponseWriter, req *http.Request) {
	var routeMatched *RouteMatched
	if routeMatched = MatchFileuploadRoute(req.URL.Path); routeMatched == nil {
		http.Error(rw, "File Upload Page Not Found!", 404)
		return
	}

	if req.Method != "POST" {
		http.Error(rw, "Forbidden", 403)
		return
	}

	req.ParseMultipartForm(1 << 26)

	prt := reflect.New(routeMatched.ClassType)
	ctx := buildContext(rw, req, routeMatched)

	prt.MethodByName("SetContext").Call([]reflect.Value{reflect.ValueOf(ctx)})
	prt.MethodByName("InitData").Call(nil)
	prt.MethodByName("DoUpload").Call(nil)
}

func webserviceHander(rw http.ResponseWriter, req *http.Request) {
	log.App.Info("webservice:", req.URL.Path)

	var routeMatched *RouteMatched
	if routeMatched = MatchWebServiceRoute(req.URL.Path); routeMatched == nil {
		http.Error(rw, "Api Not Found!", 404)
		return
	}
	prt := reflect.New(routeMatched.ClassType)

	ctx := buildContext(rw, req, routeMatched)

	if len(req.PostForm["json"]) == 0 {
		MyErr(0, "miss parameters!").Write(rw)
		return
	}

	data := &WSParams{}
	if err := json.Unmarshal([]byte(req.PostForm["json"][0]), data); err != nil {
		MyErr(0, err.Error()).Write(rw)
		//ctx.Exit(500, "decode data error")
		return
	}

	if prt.MethodByName(data.Method).Kind() == reflect.Invalid {
		MyErr(0, "invalid method:"+data.Method).Write(rw)
		//ctx.Exit(500, "invalid function call")
		return
	}
	prt.MethodByName("SetContext").Call([]reflect.Value{reflect.ValueOf(ctx)})

	var result []reflect.Value
	if data.Args == nil {
		result = prt.MethodByName(data.Method).Call(nil)
	} else {
		result = prt.MethodByName(data.Method).Call([]reflect.Value{reflect.ValueOf(data.Args)})
	}

	prt.MethodByName("Reply").Call(result)

}

func serveHTTPHander(rw http.ResponseWriter, req *http.Request) {
	log.App.Info(req.URL.Path)

	var routeMatched *RouteMatched
	if routeMatched = MatchRoute(req.URL.Path); routeMatched == nil {
		if strings.ContainsRune(req.URL.Path, '.') {
			http.ServeFile(rw, req, httpServer.StaticDir+req.URL.Path)
		} else {
			http.ServeFile(rw, req, httpServer.CacheDir+req.URL.Path+".html")
		}
		// http.Error(rw, "Page Not Found!", 404)
		return
	}

	ctx := buildContext(rw, req, routeMatched)
	prt := reflect.New(routeMatched.ClassType)
	prt.MethodByName("SetContext").Call([]reflect.Value{reflect.ValueOf(ctx)})

	doCache := false
	v := prt.MethodByName("CheckPageCache").Call(nil)

	switch int(v[0].Int()) {
	case CACHE_FOUND:
		return
	case CACHE_NOT_FOUND:
		doCache = true
	}

	isExit := func(val []reflect.Value) bool {
		return len(val) > 0 && !val[0].Bool()
	}

	if isExit(prt.MethodByName("Init").Call(nil)) {
		return
	}

	if req.Method == "POST" {
		if isExit(prt.MethodByName("Post").Call(nil)) {
			return
		}

	} else {
		if isExit(prt.MethodByName("Get").Call(nil)) {
			return
		}
	}

	if isExit(prt.MethodByName("After").Call(nil)) {
		return
	}

	if doCache {
		prt.MethodByName("CachePage").Call(nil)
		return
	}

	prt.MethodByName("RenderPage").Call(nil)
}

func buildContext(rw http.ResponseWriter, req *http.Request, routeMatched *RouteMatched) *Context {

	if req.Method == "POST" {
		req.ParseForm()
	}

	return &Context{ResponseWriter: rw, Request: req, RouterParams: routeMatched.Params}
}

func MyErr(code int, messages ...interface{}) *MyError {
	return &MyError{Code: code, Messages: messages}
}

type MyError struct {
	Code     int
	Messages []interface{}
}

func (this *MyError) Write(w io.Writer) *MyError {
	w.Write([]byte(this.String()))
	return this
}

func (this *MyError) Log(level string) *MyError {
	switch level {
	case "alert":
		log.App.Alert("MYERR", this.Code, fmt.Sprint(this.Messages...))
	case "crit":
		log.App.Crit("MYERR", this.Code, fmt.Sprint(this.Messages...))
	case "warn":
		log.App.Warn("MYERR", this.Code, fmt.Sprint(this.Messages...))
	case "notice":
		log.App.Notice("MYERR", this.Code, fmt.Sprint(this.Messages...))
	case "info":
		log.App.Info("MYERR", this.Code, fmt.Sprint(this.Messages...))
	case "debug":
		log.App.Debug("MYERR", this.Code, fmt.Sprint(this.Messages...))
	default:
		log.App.Error("MYERR", this.Code, fmt.Sprint(this.Messages...))
	}
	return this
}

func (this *MyError) Data() map[string]interface{} {
	m := make(map[string]interface{})
	m["code"] = this.Code
	m["error"] = true
	m["message"] = fmt.Sprint(this.Messages...)
	return m
}
func (this *MyError) Json() string {
	return fmt.Sprintf("{\"code\":%d,\"message\":\"%s\", \"iserror\": true}", this.Code, fmt.Sprint(this.Messages...))
}
func (this *MyError) String() string {
	return fmt.Sprintf("code %d:%s", this.Code, fmt.Sprint(this.Messages...))
}
func (this *MyError) Error() string {
	return this.String()
}

type MapData map[string]interface{}

func (this MapData) GetString(key string) string {
	return this[key].(string)
}
func (this MapData) GetInt64(key string) int64 {
	return this[key].(int64)
}
func (this MapData) GetFloat64(key string) float64 {
	return this[key].(float64)
}
func (this MapData) GetBool(key string) bool {
	return this[key].(bool)
}

type Object interface{}
