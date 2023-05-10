package engine

import (
	"net/http"
	"strings"
	"testing"

	"github.com/vizee/gapi/metadata"
)

func TestEngine_RebuildRouter(t *testing.T) {
	builder := NewBuilder()
	builder.RegisterHandler("mock-handler", &mockHandler{})
	builder.RegisterMiddleware("auth", func(ctx *Context) error {
		uid := ctx.Query().Get("uid")
		if uid != "" {
			ctx.Set("uid", uid)
			return ctx.Next()
		} else {
			http.Error(ctx.Response(), http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return nil
		}
	})
	want404 := false
	builder.NotFound(func(ctx *Context) error {
		if !want404 {
			t.Fatal("404")
		}
		return nil
	})
	e := builder.Build()
	e.RebuildRouter([]metadata.Route{
		{Method: "POST", Path: "/add", Use: []string{"auth"}, Call: mockAddCall()},
	}, true)
	req, err := http.NewRequest("POST", "http://localhost/add?uid=1", strings.NewReader(`{"a":1,"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	resp := &mockResponse{}
	e.ServeHTTP(resp, req)
	t.Logf("response %d: %s", resp.statusCode, resp.data)
}
