package engine

import (
	"net/http"
	"strings"
	"testing"

	"google.golang.org/grpc"
)

func Test_grpcRoute_handleRoute(t *testing.T) {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	gr := &grpcRoute{
		engine:      NewBuilder().Build(),
		middlewares: []HandleFunc{},
		call:        mockAddCall(),
		ch:          &mockHandler{},
		client:      cc,
	}
	req, err := http.NewRequest("POST", "http://localhost/add", strings.NewReader(`{"a":1,"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	gr.handleRoute(&mockResponse{}, req, nil)
}
