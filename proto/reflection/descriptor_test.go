package reflection

import (
	"testing"

	"github.com/vizee/gapi/testdata/pdtest"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func printFds(t *testing.T, fds *descriptorpb.FileDescriptorSet) {
	t.Logf("total: %d", len(fds.File))
	for i, fdp := range fds.File {
		t.Logf("> %d: %s", i, fdp)
	}
}

func TestCollectServerFiles(t *testing.T) {
	srv := grpc.NewServer()
	pdtest.RegisterTestServiceServer(srv, &pdtest.UnimplementedTestServiceServer{})
	fds, err := CollectServerFiles(srv)
	if err != nil {
		t.Fatal(err)
	}
	printFds(t, fds)
	data, _ := proto.Marshal(fds)
	t.Log(len(data))
}

func TestCollectGapiFiles(t *testing.T) {
	srv := grpc.NewServer()
	pdtest.RegisterTestServiceServer(srv, &pdtest.UnimplementedTestServiceServer{})
	fds, err := CollectGapiFiles(srv)
	if err != nil {
		t.Fatal(err)
	}
	printFds(t, fds)
	data, _ := proto.Marshal(fds)
	t.Log(len(data))
}
