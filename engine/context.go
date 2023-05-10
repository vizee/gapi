package engine

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/vizee/gapi/internal/ioutil"
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
	params Params
	query  url.Values
	values map[string]string
	body   []byte
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

func (c *Context) SetResponse(w http.ResponseWriter) {
	c.resp = w
}

func (c *Context) Params() Params {
	return c.params
}

func (c *Context) Query() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query
}

func (c *Context) Get(name string) (string, bool) {
	v, ok := c.values[name]
	return v, ok
}

func (c *Context) Set(name string, value string) {
	if c.values == nil {
		c.values = make(map[string]string)
	}
	c.values[name] = value
}

func (c *Context) GetBody() []byte {
	return c.body
}

func (c *Context) CacheBody(data []byte) {
	c.body = data
}

func (c *Context) ReadBody() ([]byte, error) {
	if c.body == nil {
		var err error
		c.body, err = ioutil.ReadToEnd(c.req.Body, c.req.ContentLength)
		if err != nil {
			return nil, err
		}
	}
	return c.body, nil
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

func (c *Context) reset() {
	c.req = nil
	c.resp = nil
	c.params = nil
	c.query = nil
	c.values = nil
	c.body = nil
	c.chain = nil
	c.handle = nil
	c.next = 0
}
