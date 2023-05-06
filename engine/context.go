package engine

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Params httprouter.Params

func (ps Params) Get(name string) (string, bool) {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value, true
		}
	}
	return "", false
}

type Context struct {
	req    *http.Request
	resp   http.ResponseWriter
	values map[string]string
	params Params
	chain  []HandleFunc
	handle HandleFunc
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
		ctx.values = make(map[string]string)
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

func (c *Context) Next() error {
	// 分开 chain 和 handle 为了让 chain 复用
	if c.next == len(c.chain) {
		c.next++
		return c.handle(c)
	} else if c.next < len(c.chain) {
		h := c.chain[c.next]
		c.next++
		return h(c)
	}
	return nil
}

func (ctx *Context) reset() {
	ctx.req = nil
	ctx.resp = nil
	ctx.values = nil
	ctx.params = nil
	ctx.chain = nil
	ctx.handle = nil
	ctx.next = 0
}
