package httppb

import (
	"github.com/vizee/gapi-proto-go/gapi/httpview"
	"github.com/vizee/gapi/engine"
	"github.com/vizee/gapi/internal/ioutil"
	"github.com/vizee/gapi/metadata"
	"google.golang.org/protobuf/proto"
)

var _ engine.CallHandler = &Handler{}

type Handler struct {
	PassPath      bool
	PassQuery     bool
	PassParams    bool
	CopyHeaders   bool
	FilterHeaders []string
	MaxBodySize   int64
}

func (h *Handler) ReadRequest(_ *metadata.Call, ctx *engine.Context) ([]byte, error) {
	req := ctx.Request()
	var (
		path   string
		query  string
		params []string
	)
	if h.PassPath {
		path = req.URL.Path
	}
	if h.PassQuery {
		query = req.URL.RawQuery
	}
	if h.PassParams {
		ps := ctx.Params()
		params = make([]string, 0, len(ps)*2)
		for _, p := range ps {
			params = append(params, p.Key, p.Value)
		}
	}
	var headers map[string]string
	if h.CopyHeaders {
		if len(h.FilterHeaders) == 0 {
			headers = make(map[string]string, len(req.Header))
			for name, val := range req.Header {
				if len(val) > 0 {
					headers[name] = val[0]
				} else {
					headers[name] = ""
				}
			}
		} else {
			headers = make(map[string]string, len(h.FilterHeaders))
			for _, name := range h.FilterHeaders {
				val, ok := req.Header[name]
				if !ok {
					continue
				}
				if len(val) > 0 {
					headers[name] = val[0]
				} else {
					headers[name] = ""
				}
			}
		}
	}

	body, err := ioutil.ReadLimited(req.Body, req.ContentLength, h.MaxBodySize)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&httpview.HttpRequest{
		Path:    path,
		Query:   query,
		Headers: headers,
		Params:  params,
		Body:    body,
	})
}

func (*Handler) WriteResponse(_ *metadata.Call, ctx *engine.Context, data []byte) error {
	var r httpview.HttpResponse
	err := proto.Unmarshal(data, &r)
	if err != nil {
		return err
	}
	resp := ctx.Response()
	if r.Status > 0 {
		resp.WriteHeader(int(r.Status))
	}
	for k, v := range r.Headers {
		resp.Header().Set(k, v)
	}
	if len(r.Body) > 0 {
		_, err = resp.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}
