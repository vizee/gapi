package engine

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/vizee/gapi/metadata"
	"google.golang.org/grpc"
)

type passthroughCodec struct {
}

func (*passthroughCodec) Marshal(v interface{}) ([]byte, error) {
	return v.([]byte), nil
}

func (*passthroughCodec) Unmarshal(data []byte, v interface{}) error {
	// 这里直接保留 data 的前提是 data 不会在其他地方改动或者复用，在 GRPC 1.54 版本中看起来是安全的。
	// 更合理的方式是复制 data 或者在 Unmarshal 里完成 WriteResponse
	*(v.(*[]byte)) = data
	return nil
}

func (*passthroughCodec) Name() string {
	return "passthrough"
}

type grpcRoute struct {
	engine      *Engine
	middlewares []HandleFunc
	call        *metadata.Call
	ch          CallHandler
	client      *grpc.ClientConn
}

func (r *grpcRoute) handle(ctx *Context) error {
	call := r.call

	reqData, err := r.ch.ReadRequest(call, ctx)
	if err != nil {
		return err
	}

	callctx := ctx.req.Context()
	var cancel func()
	if call.Timeout > 0 {
		callctx, cancel = context.WithTimeout(callctx, call.Timeout)
	}
	var respData []byte
	err = r.client.Invoke(callctx, call.Method, reqData, &respData, grpc.ForceCodec(&passthroughCodec{}))
	if cancel != nil {
		cancel()
	}
	if err != nil {
		return err
	}

	return r.ch.WriteResponse(call, ctx, respData)
}

func (r *grpcRoute) handleRoute(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	// 封装闭包可能带来一点点内存开销
	r.engine.Execute(w, req, Params(params), r.middlewares, r.handle)
}
