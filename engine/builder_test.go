package engine

import (
	"net/http"
	"testing"

	"github.com/vizee/gapi/internal/ioutil"
	"github.com/vizee/gapi/metadata"
	"google.golang.org/grpc"
)

var _ CallHandler = &TestHandler{}

type TestHandler struct{}

func (*TestHandler) ReadRequest(_ *metadata.Call, ctx *Context) ([]byte, error) {
	return ioutil.ReadLimited(ctx.Request().Body, ctx.Request().ContentLength, 1*1024*1024)
}

func (*TestHandler) WriteResponse(call *metadata.Call, ctx *Context, data []byte) error {
	_, err := ctx.Response().Write(data)
	return err
}

func TestBuildEngine(t *testing.T) {
	builder := NewBuilder()
	builder.RegisterHandler("test", &TestHandler{})
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
	builder.Dialer(&GrpcDialer{Opts: []grpc.DialOption{grpc.WithInsecure()}})
	builder.NotFound(func(ctx *Context) error {
		ctx.Response().Write([]byte(`404`))
		return nil
	})
	_ = builder.Build()
}
