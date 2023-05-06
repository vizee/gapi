package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/vizee/gapi/engine"
	"github.com/vizee/gapi/handlers/httpview"
	"github.com/vizee/gapi/handlers/jsonapi"
	"github.com/vizee/gapi/metadata"
	"github.com/vizee/gapi/proto/descriptor"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func writeJsonResponse(w http.ResponseWriter, o any) {
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(o)
	_, _ = w.Write(data)
}

func newEngine() *engine.Engine {
	builder := engine.NewBuilder()

	builder.RegisterHandler("httpview", &httpview.Handler{
		PassPath:      true,
		PassQuery:     true,
		PassParams:    true,
		CopyHeaders:   true,
		FilterHeaders: []string{"Content-Type", "User-Agent"},
		MaxBodySize:   1 * 1024 * 1024,
	})
	builder.RegisterHandler("jsonapi", &jsonapi.Handler{})

	builder.RegisterMiddleware("sign", func(ctx *engine.Context) error {
		sign := ctx.Request().URL.Query().Get("sign")
		if sign == "good" {
			return ctx.Next()
		} else {
			http.Error(ctx.Response(), http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}
	})

	builder.RegisterMiddleware("auth", func(ctx *engine.Context) error {
		uid := ctx.Request().URL.Query().Get("uid")
		if uid != "" {
			ctx.Set("uid", uid)
			return ctx.Next()
		} else {
			http.Error(ctx.Response(), http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return nil
		}
	})

	builder.Use(func(ctx *engine.Context) error {
		log.Printf("<> %s %s", ctx.Request().Method, ctx.Request().URL.String())
		return ctx.Next()
	})
	builder.Use(func(ctx *engine.Context) error {
		err := ctx.Next()
		if err != nil {
			if s, ok := status.FromError(err); ok {
				if s.Code() >= 1000 {
					writeJsonResponse(ctx.Response(), map[string]any{
						"code":    s.Code(),
						"message": s.Message(),
					})
					return nil
				}
			}
			log.Printf("ERROR %s %v", ctx.Request().RequestURI, err)
			http.Error(ctx.Response(), http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return nil
	})

	builder.NotFound(func(ctx *engine.Context) error {
		http.NotFound(ctx.Response(), ctx.Request())
		return nil
	})

	return builder.Build()
}

func loadRoutes(fname string) ([]metadata.Route, error) {
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	var fds descriptorpb.FileDescriptorSet
	err = proto.Unmarshal(data, &fds)
	if err != nil {
		return nil, err
	}
	p := descriptor.NewParser()
	for _, fd := range fds.File {
		err := p.AddFile(fd)
		if err != nil {
			return nil, err
		}
	}

	return metadata.ResolveRoutes(&metadata.MessageCache{}, p.Services(), false)
}

func main() {
	var (
		pdFile     string
		listenAddr string
	)
	flag.StringVar(&pdFile, "pd", "file.pd", "pd file")
	flag.StringVar(&listenAddr, "l", ":8080", "listen address")
	flag.Parse()

	engine := newEngine()
	routes, err := loadRoutes(pdFile)
	if err != nil {
		log.Fatalf("loadRoutes: %v", err)
	}
	err = engine.RebuildRouter(routes, false)
	if err != nil {
		log.Fatalf("RebuildRouter: %v", err)
	}

	log.Printf("listening: %s", listenAddr)

	err = http.ListenAndServe(listenAddr, engine)
	if err != nil {
		log.Fatalf("srv.ListenAndServe: %v", err)
	}
}
