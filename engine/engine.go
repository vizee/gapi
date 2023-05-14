package engine

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/julienschmidt/httprouter"
	"github.com/vizee/gapi/log"
	"github.com/vizee/gapi/metadata"
	"google.golang.org/grpc"
)

type HandleFunc func(ctx *Context) error

type Dialer interface {
	Dial(server string) (*grpc.ClientConn, error)
}

type CallHandler interface {
	ReadRequest(call *metadata.Call, ctx *Context) ([]byte, error)
	WriteResponse(call *metadata.Call, ctx *Context, data []byte) error
}

type Engine struct {
	middlewares map[string]HandleFunc
	handlers    map[string]CallHandler
	uses        []HandleFunc
	dialer      Dialer
	notFound    HandleFunc
	ctxpool     *sync.Pool

	router    atomic.Pointer[httprouter.Router]
	clients   map[string]*grpc.ClientConn
	routeLock sync.Mutex
}

func (e *Engine) generateMiddlewareChain(cache map[string][]HandleFunc, middlewares []string) ([]HandleFunc, error) {
	if len(middlewares) == 0 {
		return e.uses, nil
	}

	cacheKey := strings.Join(middlewares, ";")
	mws, ok := cache[cacheKey]
	if ok {
		return mws, nil
	}
	mws = append(make([]HandleFunc, 0, len(e.uses)+len(middlewares)), e.uses...)
	for _, name := range middlewares {
		mw := e.middlewares[name]
		if mw == nil {
			return nil, fmt.Errorf("no such middleware %s", name)
		}
		mws = append(mws, mw)
	}
	cache[cacheKey] = mws
	return mws, nil
}

func (e *Engine) ClearRouter() {
	e.routeLock.Lock()
	clients := e.clients
	e.clients = nil
	e.router.Store(nil)
	e.routeLock.Unlock()
	for _, cc := range clients {
		cc.Close()
	}
}

type routesSliceIter struct {
	rs []*metadata.Route
	i  int
}

func (it *routesSliceIter) NextRoute() *metadata.Route {
	if it.i < len(it.rs) {
		return it.rs[it.i]
	}
	return nil
}

func (e *Engine) RebuildRouter(routes []*metadata.Route, ignoreError bool) error {
	return RebuildEngineRouter(e, &routesSliceIter{rs: routes}, ignoreError)
}

func registerRoute(router *httprouter.Router, method string, path string, handle httprouter.Handle) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("router.Handle: %v", e)
		}
	}()
	router.Handle(method, path, handle)
	return
}

func (e *Engine) Execute(w http.ResponseWriter, req *http.Request, params Params, chain []HandleFunc, handle HandleFunc) {
	ctx := e.ctxpool.Get().(*Context)
	ctx.req = req
	ctx.resp = w
	ctx.params = params
	ctx.chain = chain
	ctx.handle = handle

	err := ctx.Next()
	if err != nil {
		log.Errorf("Execute %s: %v", req.URL.Path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	ctx.reset()
	e.ctxpool.Put(ctx)
}

func (e *Engine) NotFound(w http.ResponseWriter, req *http.Request) {
	e.Execute(w, req, nil, e.uses, e.notFound)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Debugf("Route %s %s", req.Method, req.URL.Path)

	router := e.router.Load()
	if router != nil {
		path := req.URL.Path
		handle, ps, tsr := router.Lookup(req.Method, path)
		if handle != nil {
			handle(w, req, ps)
			return
		} else if tsr && path != "/" {
			log.Debugf("Trailing slash redirect %s", req.URL.Path)
			req.URL.Path = path + "/"
			http.Redirect(w, req, req.URL.String(), http.StatusMovedPermanently)
			return
		}
	}

	e.NotFound(w, req)
}

type RouteIter interface {
	NextRoute() *metadata.Route
}

func RebuildEngineRouter[R RouteIter](e *Engine, routeIter R, ignoreError bool) error {
	e.routeLock.Lock()
	defer e.routeLock.Unlock()

	old := e.clients
	clients := make(map[string]*grpc.ClientConn)
	defer func() {
		for server, cc := range clients {
			if old[server] == nil {
				cc.Close()
			}
		}
	}()

	// 在同一次 router 构建中尽可能复用重复的 chain，在大量路由的情况下会带来一些内存节约
	chainCache := make(map[string][]HandleFunc)
	router := httprouter.New()
	for {
		route := routeIter.NextRoute()
		if route == nil {
			break
		}
		ch := e.handlers[route.Call.Handler]
		if ch == nil {
			if ignoreError {
				continue
			}
			return fmt.Errorf("route %s handler %s not found", route.Path, route.Call.Handler)
		}

		// 建立连接，尽可能复用旧连接
		client := clients[route.Call.Server]
		if client == nil {
			client = old[route.Call.Server]
			if client == nil {
				var err error
				client, err = e.dialer.Dial(route.Call.Server)
				if err != nil {
					if ignoreError {
						continue
					}
					return fmt.Errorf("dial %s: %w", route.Call.Server, err)
				}
			}
			clients[route.Call.Server] = client
		}

		middlewares, err := e.generateMiddlewareChain(chainCache, route.Use)
		if err != nil {
			if ignoreError {
				continue
			}
			return fmt.Errorf("middleware of %s: %v", route.Path, err)
		}

		gr := &grpcRoute{
			engine:      e,
			middlewares: middlewares,
			call:        route.Call,
			ch:          ch,
			client:      client,
		}
		err = registerRoute(router, route.Method, route.Path, gr.handleRoute)
		if err != nil {
			if ignoreError {
				log.Warnf("registerRoute(%s %s): %v", route.Method, route.Path, err)
				continue
			}
			return err
		}
	}

	e.router.Store(router)
	e.clients = clients
	for server, cc := range old {
		if clients[server] == nil {
			cc.Close()
		}
	}
	// 如果正常退出，确保新链接不会被关闭
	clients = nil

	return nil
}
