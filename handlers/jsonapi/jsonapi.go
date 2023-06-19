package jsonapi

import (
	"github.com/vizee/gapi/engine"
	"github.com/vizee/gapi/metadata"
	"github.com/vizee/jsonpb"
	"github.com/vizee/jsonpb/jsonlit"
	"github.com/vizee/jsonpb/proto"
)

var _ engine.CallHandler = &Handler{}

type Handler struct {
	SurroundOutput [2]string
	EmptyRequest   bool
}

func (h *Handler) ReadRequest(call *metadata.Call, ctx *engine.Context) ([]byte, error) {
	data, err := ctx.ReadBody()
	if err != nil {
		return nil, err
	}
	var enc proto.Encoder
	if len(data) > 0 || !h.EmptyRequest {
		err := jsonpb.TranscodeToProto(&enc, jsonlit.NewIter(data), call.In)
		if err != nil {
			return nil, err
		}
	}
	if len(call.Bindings) > 0 {
		err = appendBindings(&enc, ctx, call.Bindings)
		if err != nil {
			return nil, err
		}
	}
	return enc.Bytes(), nil
}

func (h *Handler) WriteResponse(call *metadata.Call, ctx *engine.Context, data []byte) error {
	var j jsonpb.JsonBuilder
	j.AppendString(h.SurroundOutput[0])
	err := jsonpb.TranscodeToJson(&j, proto.NewDecoder(data), call.Out)
	if err != nil {
		return err
	}
	j.AppendString(h.SurroundOutput[1])

	resp := ctx.Response()
	resp.Header().Set("Content-Type", "application/json")
	_, err = resp.Write(j.IntoBytes())
	return err
}
