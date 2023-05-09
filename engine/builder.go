package engine

import (
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/vizee/gapi/internal/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcDialer struct {
	Opts []grpc.DialOption
}

func (d *GrpcDialer) Dial(server string) (*grpc.ClientConn, error) {
	return grpc.Dial(server, d.Opts...)
}

type Builder struct {
	engine *Engine
}

func NewBuilder() *Builder {
	b := &Builder{
		engine: &Engine{
			ctxpool: &sync.Pool{
				New: func() any {
					return &Context{}
				},
			},
			middlewares: make(map[string]HandleFunc),
			handlers:    make(map[string]CallHandler),
			dialer: &GrpcDialer{
				Opts: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			},
			notFound: func(ctx *Context) error {
				http.NotFound(ctx.resp, ctx.req)
				return nil
			},
		},
	}
	b.engine.router.Store(httprouter.New())
	return b
}

func (b *Builder) RegisterHandler(name string, handler CallHandler) {
	b.engine.handlers[name] = handler
}

func (b *Builder) RegisterMiddleware(name string, handle HandleFunc) {
	b.engine.middlewares[name] = handle
}

func (b *Builder) Use(use HandleFunc) {
	b.engine.uses = append(b.engine.uses, use)
}

func (b *Builder) Dialer(dialer Dialer) {
	b.engine.dialer = dialer
}

func (b *Builder) NotFound(notFound HandleFunc) {
	b.engine.notFound = notFound
}

func (b *Builder) Build() *Engine {
	b.engine.uses = slices.Shrink(b.engine.uses)
	return b.engine
}
