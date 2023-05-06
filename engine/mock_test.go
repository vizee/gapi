package engine

import (
	"net/http"

	"github.com/vizee/gapi/internal/ioutil"
	"github.com/vizee/gapi/metadata"
	"github.com/vizee/jsonpb"
)

type mockHandler struct{}

func (*mockHandler) ReadRequest(_ *metadata.Call, ctx *Context) ([]byte, error) {
	return ioutil.ReadLimited(ctx.Request().Body, ctx.Request().ContentLength, 1*1024*1024)
}

func (*mockHandler) WriteResponse(call *metadata.Call, ctx *Context, data []byte) error {
	_, err := ctx.Response().Write(data)
	return err
}

type mockResponse struct {
	statusCode int
	header     http.Header
	data       []byte
}

// Header implements http.ResponseWriter
func (r *mockResponse) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

// Write implements http.ResponseWriter
func (r *mockResponse) Write(data []byte) (int, error) {
	r.data = append(r.data, data...)
	return len(data), nil
}

// WriteHeader implements http.ResponseWriter
func (r *mockResponse) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func mockAddCall() *metadata.Call {
	return &metadata.Call{
		Server:  "localhost:50051",
		Handler: "mock-handler",
		Method:  "Add",
		In: &jsonpb.Message{
			Name: ".gapi.testdata.pdtest.AddRequest",
			Fields: []jsonpb.Field{
				{
					Name:     "a",
					Kind:     2,
					Ref:      nil,
					Tag:      1,
					Repeated: false,
					Omit:     0,
				},
				{
					Name:     "b",
					Kind:     2,
					Ref:      nil,
					Tag:      2,
					Repeated: false,
					Omit:     0,
				},
			},
		},
		Out: &jsonpb.Message{
			Name: ".gapi.testdata.pdtest.AddResponse",
			Fields: []jsonpb.Field{
				{
					Name:     "sum",
					Kind:     2,
					Ref:      nil,
					Tag:      1,
					Repeated: false,
					Omit:     0,
				},
			},
		},
		Timeout: 5000000000,
	}
}
