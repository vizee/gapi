package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/vizee/gapi/proto/descriptor"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestResolveRoutes(t *testing.T) {
	data, err := os.ReadFile("../examples/proto/pdtest.pd")
	if err != nil {
		t.Fatal(err)
	}

	var fds descriptorpb.FileDescriptorSet
	err = proto.Unmarshal(data, &fds)
	if err != nil {
		t.Fatal(err)
	}
	p := descriptor.NewParser()
	for _, fd := range fds.File {
		err := p.AddFile(fd)
		if err != nil {
			t.Fatal(err)
		}
	}
	routes, err := ResolveRoutes(&MessageCache{}, p.Services(), false)
	if err != nil {
		t.Fatal(err)
	}
	j, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(j))
}
