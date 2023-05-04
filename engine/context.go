package engine

import (
	"net/http"
)

type HandleFunc func(ctx *Context) error

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) Get(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

type Context struct {
	req    *http.Request
	resp   http.ResponseWriter
	values map[string]string
	params Params
	chain  []HandleFunc
	next   int
}

func (c *Context) Request() *http.Request {
	return c.req
}

func (c *Context) Response() http.ResponseWriter {
	return c.resp
}

func (ctx *Context) Set(name string, value string) {
	if ctx.values == nil {
		ctx.values = map[string]string{}
	}
	ctx.values[name] = value
}

func (ctx *Context) Get(name string) (string, bool) {
	v, ok := ctx.values[name]
	return v, ok
}

func (c *Context) Params() Params {
	return c.params
}

func (ctx *Context) reset() {
	ctx.req = nil
	ctx.resp = nil
	ctx.values = nil
	ctx.params = nil
	ctx.chain = nil
	ctx.next = 0
}
