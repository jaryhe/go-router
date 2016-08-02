package rsm

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
)

type ServeMux struct {
	mu    sync.RWMutex
	m     map[string]muxEntry
	hosts bool // whether any patterns contain hostnames
}

type muxEntry struct {
	explicit bool
	h        Handler
	pattern  string
}

//my defined handler
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, map[string]interface{})
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux { return new(ServeMux) }

// DefaultServeMux is the default ServeMux used by Serve.
var DefaultServeMux = &defaultServeMux

var defaultServeMux ServeMux

// Does path match pattern?
func pathMatch(pattern, path string) (bool, map[string]interface{}) {
	var m map[string]interface{}

	if m == nil {
		m = make(map[string]interface{})
	}

	var a = strings.Split(path, "/")
	var b = strings.Split(pattern, "/")

	//fmt.Println(a)
	if len(a) != len(b) {
		return false, nil
	}

	for index := 0; index < len(a); index++ {
		//fmt.Println(index,a[index])
		if strings.HasPrefix(b[index], ":") {
			m[strings.Split(b[index], ":")[1]] = a[index]
			continue
		}
		if a[index] != b[index] {
			return false, nil
		}
	}
	return true, m
}
func (mux *ServeMux) match(path string) (h Handler, pattern string, m map[string]interface{}) {
	var n = 0

	//var parasmap map[string]interface{}

	for k, v := range mux.m {
		/*regPattern := regexp.MustCompile(k)
		if ok := regPattern.MatchString(path); !ok {
			continue
		}*/

		//fmt.Println(k)

		ret, parasmap := pathMatch(k, path)
		if !ret {
			continue
		}
		if h == nil || len(k) > n {
			n = len(k)
			h = v.h
			m = parasmap
			pattern = v.pattern
		}
	}
	return
}

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// Handler returns the handler to use for the given request,
// consulting r.Method, r.Host, and r.URL.Path. It always returns
// a non-nil handler. If the path is not in its canonical form, the
// handler will be an internally-generated handler that redirects
// to the canonical path.
//
// Handler also returns the registered pattern that matches the
// request or, in the case of internally-generated redirects,
// the pattern that will match after following the redirect.
//
// If there is no registered handler that applies to the request,
// Handler returns a ``page not found'' handler and an empty pattern.

func (mux *ServeMux) Handler(r *http.Request) (h Handler, pattern string, m map[string]interface{}) {
	if r.Method != "CONNECT" {
		if p := cleanPath(r.URL.Path); p != r.URL.Path {
			_, pattern, _ = mux.handler(r.Host, p)
			url := *r.URL
			url.Path = p
			return nil, "", nil
		}
	}

	return mux.handler(r.Host, r.URL.Path)
}

// handler is the main implementation of Handler.
// The path is known to be in canonical form, except for CONNECT methods.
func (mux *ServeMux) handler(host, path string) (h Handler, pattern string, m map[string]interface{}) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	fmt.Println(path)
	// Host-specific pattern takes precedence over generic ones
	if mux.hosts {
		h, pattern, m = mux.match(host + path)
	}
	if h == nil {
		h, pattern, m = mux.match(path)
	}
	if h == nil {
		//h, pattern = nil, ""
		h, pattern = nil, ""
		m = nil
	}
	return
}

// Helper handlers

// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be plain text.
func Error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, error)
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, _, m := mux.Handler(r)
	if h == nil {
		Error(w, "404 page not found", 404)
	} else {
		h.ServeHTTP(w, r, m)
	}
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (mux *ServeMux) Handle(pattern string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("http: nil handler")
	}
	if mux.m[pattern].explicit {
		panic("http: multiple registrations for " + pattern)
	}

	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}

	fmt.Println(pattern)
	mux.m[pattern] = muxEntry{explicit: false, h: handler, pattern: pattern}

	if pattern[0] != '/' {
		mux.hosts = true
	}
	addPath := pattern + "/"
	fmt.Println(addPath)
	mux.m[addPath] = muxEntry{explicit: false, h: handler, pattern: addPath}

}
