package reflection

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func collectFiles(fds []*descriptorpb.FileDescriptorProto, visit map[string]bool, fd protoreflect.FileDescriptor) []*descriptorpb.FileDescriptorProto {
	visit[fd.Path()] = true
	fds = append(fds, protodesc.ToFileDescriptorProto(fd))
	imports := fd.Imports()
	for i := 0; i < imports.Len(); i++ {
		imported := imports.Get(i)
		if visit[imported.Path()] {
			continue
		}
		fds = collectFiles(fds, visit, imported.FileDescriptor)
	}
	return fds
}

func CollecServerFiles(srv *grpc.Server) ([]*descriptorpb.FileDescriptorProto, error) {
	serviceInfo := srv.GetServiceInfo()
	var fds []*descriptorpb.FileDescriptorProto
	visit := make(map[string]bool)
	for name := range serviceInfo {
		sd, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(name))
		if err != nil {
			return nil, err
		}
		fd := sd.ParentFile()
		if visit[fd.Path()] {
			continue
		}

		fds = collectFiles(fds, visit, fd)
	}
	return fds, nil
}
