package gos

import (
	"net/http"
	"net/http/pprof"
)

func pprofDefaultHander(rw http.ResponseWriter, req *http.Request) {
	pprof.Index(rw, req)
}

func pprofProfileHander(rw http.ResponseWriter, req *http.Request) {
	pprof.Profile(rw, req)
}

func pprofCmdlineHander(rw http.ResponseWriter, req *http.Request) {
	pprof.Cmdline(rw, req)
}

func pprofSymbolHander(rw http.ResponseWriter, req *http.Request) {
	pprof.Symbol(rw, req)
}
