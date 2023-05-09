package engine

import (
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestBuildEngine(t *testing.T) {
	builder := NewBuilder()
	builder.RegisterHandler("mock-handler", &mockHandler{})
	builder.RegisterMiddleware("auth", func(ctx *Context) error {
		uid := ctx.Request().FormValue("uid")
		if uid != "" {
			ctx.Set("uid", uid)
			return ctx.Next()
		} else {
			http.Error(ctx.Response(), http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return nil
		}
	})
	builder.Use(func(ctx *Context) error {
		defer recover()
		return ctx.Next()
	})
	builder.Dialer(&GrpcDialer{Opts: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}})
	builder.NotFound(func(ctx *Context) error {
		ctx.Response().Write([]byte(`404`))
		return nil
	})
	_ = builder.Build()
}
