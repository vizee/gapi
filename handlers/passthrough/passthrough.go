package passthrough

import (
	"github.com/vizee/gapi/engine"
	"github.com/vizee/gapi/metadata"
)

var _ engine.CallHandler = &Handler{}

type Handler struct {
}

func (*Handler) ReadRequest(_ *metadata.Call, ctx *engine.Context) ([]byte, error) {
	return ctx.ReadBody()
}

func (*Handler) WriteResponse(call *metadata.Call, ctx *engine.Context, data []byte) error {
	_, err := ctx.Response().Write(data)
	return err
}
