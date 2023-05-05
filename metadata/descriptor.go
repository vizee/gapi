package metadata

import (
	"errors"
	"time"

	annotation "github.com/vizee/gapi-proto-go/gapi"
	"github.com/vizee/gapi/internal/slices"
	"github.com/vizee/gapi/proto/descriptor"
	"github.com/vizee/jsonpb"
	"google.golang.org/protobuf/types/descriptorpb"
)

type messageDesc struct {
	*jsonpb.Message
	bindings []FieldBinding
}

type MessageCache struct {
	cache map[string]*messageDesc
}

func (mc *MessageCache) Resolve(md *descriptor.MessageDesc) *messageDesc {
	if mc.cache == nil {
		mc.cache = make(map[string]*messageDesc)
	}

	msg := mc.cache[md.Name]
	if msg != nil {
		return msg
	}

	msg = &messageDesc{
		Message: &jsonpb.Message{
			Name:   md.Name,
			Fields: make([]jsonpb.Field, 0, len(md.Fields)),
		},
	}
	// 防止递归
	mc.cache[msg.Name] = msg

	for _, fd := range md.Fields {
		kind, ok := getTypeKind(fd.Type)
		if !ok {
			continue
		}
		name := fd.Name
		if fd.Alias != "" {
			name = fd.Alias
		}

		if fd.Bind == annotation.FIELD_BIND_FROM_DEFAULT {
			repeated := fd.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
			var msgRef *jsonpb.Message
			if kind == jsonpb.MessageKind {
				msgRef = mc.Resolve(fd.Ref).Message
				if fd.Ref.MapEntry {
					kind = jsonpb.MapKind
					repeated = false
				}
			}
			omit := jsonpb.OmitProtoEmpty
			if fd.OmitEmpty {
				omit = jsonpb.OmitEmpty
			}
			msg.Fields = append(msg.Fields, jsonpb.Field{
				Name:     name,
				Kind:     kind,
				Ref:      msgRef,
				Tag:      uint32(fd.Tag),
				Repeated: repeated,
				Omit:     omit,
			})
		} else {
			var bind BindSource
			switch fd.Bind {
			case annotation.FIELD_BIND_FROM_QUERY:
				bind = BindQuery
			case annotation.FIELD_BIND_FROM_PARAMS:
				bind = BindParams
			case annotation.FIELD_BIND_FROM_HEADER:
				bind = BindHeader
			case annotation.FIELD_BIND_FROM_CONTEXT:
				bind = BindContext
			}
			msg.bindings = append(msg.bindings, FieldBinding{
				Name: name,
				Kind: kind,
				Tag:  uint32(fd.Tag),
				Bind: bind,
			})
		}
	}

	msg.BakeTagIndex()
	msg.BakeNameIndex()

	if len(msg.bindings) > 0 {
		msg.bindings = slices.Shrink(msg.bindings)
	}

	return msg
}

func ResolveRoutes(mc *MessageCache, sds []*descriptor.ServiceDesc, ignoreError bool) ([]Route, error) {
	var routes []Route

walksd:
	for _, sd := range sds {
		server := sd.Opts.Server
		if server == "" {
			if ignoreError {
				continue
			}
			return nil, errors.New("invalid service '" + sd.Name + "'")
		}
		for _, use := range sd.Opts.Use {
			if !checkMiddlewareName(use) {
				if ignoreError {
					continue walksd
				}
				return nil, errors.New("invalid middleware name '" + use + "'")
			}
		}

	walkmd:
		for _, md := range sd.Methods {
			for _, use := range md.Opts.Use {
				if !checkMiddlewareName(use) {
					if ignoreError {
						continue walkmd
					}
					return nil, errors.New("invalid middleware name '" + use + "'")
				}
			}

			handler := md.Opts.Handler
			if handler == "" {
				handler = sd.Opts.DefaultHandler
			}
			if handler == "" || md.Opts.Method == "" || md.Opts.Path == "" || md.In == nil || md.In.Incomplete || md.Out == nil || md.Out.Incomplete {
				if ignoreError {
					continue
				}
				return nil, errors.New("invalid method '" + md.Name + "'")
			}

			timeout := md.Opts.Timeout
			if timeout == 0 {
				timeout = sd.Opts.DefaultTimeout
			}

			inMsg := mc.Resolve(md.In)
			routes = append(routes, Route{
				Method: md.Opts.Method,
				Path:   md.Opts.Path,
				Use:    slices.Merge(sd.Opts.Use, md.Opts.Use),
				Call: &Call{
					Server:   server,
					Handler:  handler,
					Method:   md.Name,
					In:       inMsg.Message,
					Out:      mc.Resolve(md.Out).Message,
					Bindings: inMsg.bindings,
					Timeout:  time.Duration(timeout) * time.Millisecond,
				},
			})
		}
	}

	return routes, nil
}

func checkMiddlewareName(name string) bool {
	if name == "" {
		return false
	}

	for i := 0; i < len(name); i++ {
		c := name[i]
		if 'a' <= c && c <= 'z' ||
			'A' <= c && c <= 'Z' ||
			'0' <= c && c <= '9' ||
			c == '_' || c == '-' {
			continue
		}
		return false
	}
	return true
}

var typeKinds = [...]jsonpb.Kind{
	descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:   jsonpb.DoubleKind,
	descriptorpb.FieldDescriptorProto_TYPE_FLOAT:    jsonpb.FloatKind,
	descriptorpb.FieldDescriptorProto_TYPE_INT64:    jsonpb.Int64Kind,
	descriptorpb.FieldDescriptorProto_TYPE_UINT64:   jsonpb.Uint64Kind,
	descriptorpb.FieldDescriptorProto_TYPE_INT32:    jsonpb.Int32Kind,
	descriptorpb.FieldDescriptorProto_TYPE_FIXED64:  jsonpb.Fixed64Kind,
	descriptorpb.FieldDescriptorProto_TYPE_FIXED32:  jsonpb.Fixed32Kind,
	descriptorpb.FieldDescriptorProto_TYPE_BOOL:     jsonpb.BoolKind,
	descriptorpb.FieldDescriptorProto_TYPE_STRING:   jsonpb.StringKind,
	descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:  jsonpb.MessageKind,
	descriptorpb.FieldDescriptorProto_TYPE_BYTES:    jsonpb.BytesKind,
	descriptorpb.FieldDescriptorProto_TYPE_UINT32:   jsonpb.Uint32Kind,
	descriptorpb.FieldDescriptorProto_TYPE_ENUM:     jsonpb.Int32Kind,
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: jsonpb.Sfixed32Kind,
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: jsonpb.Sfixed64Kind,
	descriptorpb.FieldDescriptorProto_TYPE_SINT32:   jsonpb.Sint32Kind,
	descriptorpb.FieldDescriptorProto_TYPE_SINT64:   jsonpb.Sint64Kind,
}

func getTypeKind(ty descriptorpb.FieldDescriptorProto_Type) (jsonpb.Kind, bool) {
	if int(ty) < len(typeKinds) {
		return typeKinds[ty], true
	}
	return 0, false
}