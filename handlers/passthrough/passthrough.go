package passthrough

import (
	"github.com/vizee/gapi/engine"
	"github.com/vizee/gapi/internal/ioutil"
	"github.com/vizee/gapi/metadata"
)

var _ engine.CallHandler = &Handler{}

type Handler struct {
}

func (*Handler) ReadRequest(_ *metadata.Call, ctx *engine.Context) ([]byte, error) {
	return ioutil.ReadLimited(ctx.Request().Body, ctx.Request().ContentLength, 1*1024*1024)
}

func (*Handler) WriteResponse(call *metadata.Call, ctx *engine.Context, data []byte) error {
	_, err := ctx.Response().Write(data)
	return err
}
