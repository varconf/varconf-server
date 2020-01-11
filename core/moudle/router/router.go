package router

import (
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"
	ANY     = "*"
)

const (
	DEFAULT = "DEFAULT"
)

const (
	BIND  = "BIND"
	ROOT  = "ROOT"
	INDEX = "INDEX"
)

type Handler func(http.ResponseWriter, *http.Request, *Context)

type Interceptor interface {
	PreHandleFunc(http.ResponseWriter, *http.Request, *Context) bool
	PostHandleFunc(http.ResponseWriter, *http.Request, *Context)
}

type Context struct {
	Data map[string]interface{}
}

type Resolver func(http.ResponseWriter, *http.Request, error)

type PathPattern struct {
	path   string
	regex  *regexp.Regexp
	params map[int]string
}

type HandlerAdapter struct {
	pathPattern *PathPattern
	method      string
	handler     Handler
	bind        interface{}
}

type InterceptorAdapter struct {
	pathPattern    *PathPattern
	ignorePatterns []*PathPattern
	interceptor    Interceptor
}

type ResolverAdapter struct {
	error    string
	resolver Resolver
}

type Router struct {
	addr                string
	tag                 string
	listener            net.Listener
	logger              *log.Logger
	handlerAdapters     []*HandlerAdapter
	interceptorAdapters []*InterceptorAdapter
	resolverAdapters    map[string]*ResolverAdapter
}

func NewRouter() *Router {
	return &Router{
		addr:                ":8888",
		tag:                 "varconf",
		logger:              log.New(os.Stdout, "", log.Ldate|log.Ltime),
		handlerAdapters:     make([]*HandlerAdapter, 0),
		interceptorAdapters: make([]*InterceptorAdapter, 0),
		resolverAdapters:    make(map[string]*ResolverAdapter),
	}
}

func (_self *Router) Run() error {
	listener, err := net.Listen("tcp", _self.addr)
	if err != nil {
		_self.logger.Fatal("Listen: ", err)
		return err
	}

	_self.listener = listener
	_self.logger.Println("Listening on http://" + _self.addr)

	err = http.Serve(_self.listener, _self)
	if err != nil {
		_self.logger.Fatal("ListenAndServe: ", err)
	}
	return err
}

func (_self *Router) RunTLS(certFile, keyFile string) error {
	listener, err := net.Listen("tcp", _self.addr)
	if err != nil {
		_self.logger.Fatal("Listen: ", err)
		return err
	}

	_self.listener = listener
	_self.logger.Println("Listening on https://" + _self.addr)

	err = http.ServeTLS(_self.listener, _self, certFile, keyFile)
	if err != nil {
		_self.logger.Fatal("ListenAndServe: ", err)
	}
	return err
}

func (_self *Router) Stop() error {
	if _self.listener != nil {
		return _self.listener.Close()
	}
	return nil
}

func (_self *Router) SetAddress(args ...interface{}) {
	host := "0.0.0.0"
	port := "8888"

	if len(args) == 1 {
		switch arg := args[0].(type) {
		case string:
			addrs := strings.Split(args[0].(string), ":")
			if len(addrs) == 1 {
				host = addrs[0]
			} else if len(addrs) >= 2 {
				port = addrs[1]
			}
		case int:
			port = strconv.Itoa(arg)
		}
	} else if len(args) >= 2 {
		if arg, ok := args[0].(string); ok {
			host = arg
		}
		if arg, ok := args[1].(int); ok {
			port = strconv.Itoa(arg)
		}
	}

	_self.addr = host + ":" + port
}

func (_self *Router) SetTag(tag string) {
	_self.tag = tag
}

func (_self *Router) SetLogger(logger *log.Logger) {
	_self.logger = logger
}

func (_self *Router) Connect(path string, handlerFunc Handler) {
	_self.AddRoute(CONNECT, path, handlerFunc)
}

func (_self *Router) Delete(path string, handlerFunc Handler) {
	_self.AddRoute(DELETE, path, handlerFunc)
}

func (_self *Router) Get(path string, handlerFunc Handler) {
	_self.AddRoute(GET, path, handlerFunc)
}

func (_self *Router) Head(path string, handlerFunc Handler) {
	_self.AddRoute(HEAD, path, handlerFunc)
}

func (_self *Router) Options(path string, handlerFunc Handler) {
	_self.AddRoute(OPTIONS, path, handlerFunc)
}

func (_self *Router) Patch(path string, handlerFunc Handler) {
	_self.AddRoute(PATCH, path, handlerFunc)
}

func (_self *Router) Post(path string, handlerFunc Handler) {
	_self.AddRoute(POST, path, handlerFunc)
}

func (_self *Router) Put(path string, handlerFunc Handler) {
	_self.AddRoute(PUT, path, handlerFunc)
}

func (_self *Router) Trace(path string, handlerFunc Handler) {
	_self.AddRoute(TRACE, path, handlerFunc)
}

func (_self *Router) Any(path string, handlerFunc Handler) {
	_self.AddRoute(ANY, path, handlerFunc)
}

func (_self *Router) Static(path, root, index string) {
	bind := make(map[string]interface{})

	bind[ROOT] = root
	bind[INDEX] = index

	// parse pathPattern
	pathPattern, err := _self.parsePattern(path)
	if err != nil {
		return
	}

	// add route
	adapter := &HandlerAdapter{}
	adapter.pathPattern = pathPattern
	adapter.method = GET
	adapter.handler = _self.serveFile
	adapter.bind = bind

	_self.handlerAdapters = append(_self.handlerAdapters, adapter)
}

func (_self *Router) AddRoute(method, path string, handlerFunc Handler) {
	// parse pathPattern
	pathPattern, err := _self.parsePattern(path)
	if err != nil {
		return
	}

	// add route
	adapter := &HandlerAdapter{}
	adapter.pathPattern = pathPattern
	adapter.method = method
	adapter.handler = handlerFunc
	adapter.bind = nil

	_self.handlerAdapters = append(_self.handlerAdapters, adapter)
}

func (_self *Router) AddFilter(path string, ignores []string, interceptor Interceptor) {
	// parse pathPattern
	pathPattern, err := _self.parsePattern(path)
	if err != nil {
		return
	}

	// parse ignores
	ignorePatterns := make([]*PathPattern, 0)
	for _, ignore := range ignores {
		ignorePattern, err := _self.parsePattern(ignore)
		if err != nil {
			return
		}
		ignorePatterns = append(ignorePatterns, ignorePattern)
	}

	adapter := &InterceptorAdapter{}
	adapter.interceptor = interceptor
	adapter.pathPattern = pathPattern
	adapter.ignorePatterns = ignorePatterns

	_self.interceptorAdapters = append(_self.interceptorAdapters, adapter)
}

func (_self *Router) AddResolver(err error, resolver Resolver) {
	// add route
	adapter := &ResolverAdapter{}
	adapter.resolver = resolver
	if err == nil {
		adapter.error = DEFAULT
	} else {
		adapter.error = reflect.TypeOf(err).String()
	}
	_self.resolverAdapters[adapter.error] = adapter
}

func (_self *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestId := r.Header.Get("Request-Id")
	if requestId == "" {
		requestId = uuid.New().String()
	}
	w.Header().Set("Request-Id", requestId)
	w.Header().Set("Server", _self.tag)
	w.Header().Set("Date", _self.formatTime(time.Now().UTC()))

	// parse match adapter
	interceptorAdapters := make([]*InterceptorAdapter, 0)
	for _, interceptor := range _self.interceptorAdapters {
		if _self.matchPattern(r, interceptor.pathPattern) {
			isIgnore := false
			for _, ignorePattern := range interceptor.ignorePatterns {
				if _self.matchPattern(r, ignorePattern) {
					isIgnore = true
					break
				}
			}
			if !isIgnore {
				interceptorAdapters = append(interceptorAdapters, interceptor)
			}
		}
	}

	// pre handle the request
	c := &Context{Data: make(map[string]interface{})}
	for _, adapter := range interceptorAdapters {
		if adapter.interceptor != nil {
			if !adapter.interceptor.PreHandleFunc(w, r, c) {
				return
			}
		}
	}

	isFound := false
	// serve the request
	for _, adapter := range _self.handlerAdapters {
		// match the adapter
		if adapter.method == ANY || adapter.method == r.Method {
			if _self.matchPattern(r, adapter.pathPattern) {
				_self.serveRequest(w, r, adapter, c)
				isFound = true
				break
			}
		}
	}
	// not found
	if !isFound {
		http.NotFound(w, r)
		return
	}

	// post handle the the request
	for _, interceptor := range interceptorAdapters {
		if interceptor.interceptor != nil {
			interceptor.interceptor.PostHandleFunc(w, r, c)
		}
	}
}

func (_self *Router) parsePattern(path string) (*PathPattern, error) {
	// split the url into sections
	parts := strings.Split(path, "/")

	// find params that start with ":"
	// replace with regular expressions
	j := 0
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			if n := strings.Index(part, "("); n != -1 {
				expr = part[n:]
				part = part[:n]
			}
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	// re create the url path, with parameters replaced
	path = strings.Join(parts, "/")

	// check the pathPattern
	regex, err := regexp.Compile(path)
	if err != nil {
		panic(err)
		return nil, err
	}

	return &PathPattern{path: path, regex: regex, params: params}, nil
}

func (_self *Router) matchPattern(r *http.Request, p *PathPattern) bool {
	// path match
	regex := p.regex
	requestPath := r.URL.Path
	if !regex.MatchString(requestPath) {
		return false
	}
	matches := regex.FindStringSubmatch(requestPath)
	if len(matches[0]) != len(requestPath) {
		return false
	}

	// match params
	params := p.params
	values := r.URL.Query()
	if len(params) > 0 {
		for i, match := range matches[1:] {
			values.Add(params[i], match)
		}
		// add to raw query
		r.URL.RawQuery = url.Values(values).Encode()
	}
	return true
}

func (_self *Router) serveRequest(w http.ResponseWriter, r *http.Request, a *HandlerAdapter, c *Context) {
	_self.logger.Println(r.Method, r.RequestURI)

	defer func() {
		if err := recover(); err != nil {
			_self.logger.Println(err)

			resolverAdapter := _self.resolverAdapters[reflect.TypeOf(err).String()]
			if resolverAdapter == nil {
				resolverAdapter = _self.resolverAdapters[DEFAULT]
				if resolverAdapter != nil {
					e, ok := err.(error)
					if ok {
						resolverAdapter.resolver(w, r, e)
					}
				}
			}
		}
	}()

	if a.bind != nil {
		c.Data[BIND] = a.bind
	}
	// handle the request
	a.handler(w, r, c)
}

func (_self *Router) serveFile(w http.ResponseWriter, r *http.Request, c *Context) {
	// check the bind data and type
	bind := c.Data[BIND]
	if bind == nil {
		http.NotFound(w, r)
		return
	}

	bindData, ok := bind.(map[string]interface{})
	if !ok {
		http.NotFound(w, r)
		return
	}

	root := bindData[ROOT]
	rootData, ok := root.(string)
	if !ok {
		http.NotFound(w, r)
		return
	}

	index := bindData[INDEX]
	indexData, ok := index.(string)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// deal the file path
	filePath := _self.joinPath(rootData, r.URL.Path)
	if _self.isDirExists(filePath) {
		filePath = _self.joinPath(filePath, indexData)
	}

	// check the file
	if !_self.isFileExists(filePath) {
		http.NotFound(w, r)
		return
	}

	// serve the file
	content, suffix, err := _self.readFile(filePath)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", mime.TypeByExtension(suffix))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	//w.Header().Set("Cache-Control", "private, immutable, max-age=60")
	w.Write(content)
}

func (_self *Router) formatTime(t time.Time) string {
	webTime := t.Format(time.RFC1123)
	if strings.HasSuffix(webTime, "UTC") {
		webTime = webTime[0:len(webTime)-3] + "GMT"
	}

	return webTime
}

func (_self *Router) joinPath(elem ...string) string {
	return filepath.Join(elem...)
}

func (_self *Router) readFile(filePath string) ([]byte, string, error) {
	// serve the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, "", err
	}
	ext := path.Ext(filePath)

	return data, ext, nil
}

func (_self *Router) isDirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func (_self *Router) isFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
