package jsonapi

import (
	"strconv"

	"github.com/vizee/gapi/engine"
	"github.com/vizee/gapi/metadata"
	"github.com/vizee/jsonpb"
	"github.com/vizee/jsonpb/proto"
)

func appendBindings(enc *proto.Encoder, ctx *engine.Context, bindings []metadata.FieldBinding) error {
	for _, b := range bindings {
		var val string
		switch b.Bind {
		case metadata.BindQuery:
			val = ctx.Query().Get(b.Name)
		case metadata.BindParams:
			var ok bool
			val, ok = ctx.Params().Get(b.Name)
			if !ok {
				continue
			}
		case metadata.BindHeader:
			val = ctx.Request().Header.Get(b.Name)
		case metadata.BindContext:
			var ok bool
			val, ok = ctx.Get(b.Name)
			if !ok {
				continue
			}
		default:
			continue
		}
		switch b.Kind {
		case jsonpb.Int64Kind:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			enc.EmitVarint(b.Tag, uint64(n))
		case jsonpb.Int32Kind:
			n, err := strconv.ParseInt(val, 10, 32)
			if err != nil {
				return err
			}
			enc.EmitVarint(b.Tag, uint64(n))
		case jsonpb.StringKind:
			enc.EmitString(b.Tag, val)
		case jsonpb.BoolKind:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			if n != 0 {
				n = 1
			}
			enc.EmitVarint(b.Tag, uint64(n))
		default:
			return jsonpb.ErrTypeMismatch
		}
	}
	return nil
}
