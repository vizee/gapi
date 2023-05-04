package jsonapi

import (
	"github.com/vizee/gapi/engine"
	"github.com/vizee/gapi/internal/ioutil"
	"github.com/vizee/gapi/metadata"
	"github.com/vizee/jsonpb"
	"github.com/vizee/jsonpb/jsonlit"
	"github.com/vizee/jsonpb/proto"
)

type Handler struct {
	SurroundOutput []string
}

func (h *Handler) ReadRequest(call *metadata.Call, ctx *engine.Context) ([]byte, error) {
	data, err := ioutil.ReadLimited(ctx.Request().Body, ctx.Request().ContentLength, 1*1024*1024)
	if err != nil {
		return nil, err
	}
	var enc proto.Encoder
	err = jsonpb.TranscodeToProto(&enc, jsonlit.NewIter(data), call.In)
	if err != nil {
		return nil, err
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
	if len(h.SurroundOutput) > 0 {
		j.AppendString(h.SurroundOutput[0])
	}
	err := jsonpb.TranscodeToJson(&j, proto.NewDecoder(data), call.Out)
	if err != nil {
		return err
	}
	if len(h.SurroundOutput) > 1 {
		j.AppendString(h.SurroundOutput[1])
	}

	resp := ctx.Response()
	resp.Header().Set("Content-Type", "application/json")
	_, err = resp.Write(data)
	return err
}
