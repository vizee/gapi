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

func ResolveDescToRoutes(mc *MessageCache, sds []*descriptor.ServiceDesc, skipInvalid bool) ([]Route, error) {
	var routes []Route

	for _, sd := range sds {
		server := sd.Opts.Server
		if server == "" {
			if skipInvalid {
				continue
			}
			return nil, errors.New("invalid service '" + sd.Name + "'")
		}

		for _, md := range sd.Methods {
			handler := md.Opts.Handler
			if handler == "" {
				handler = sd.Opts.DefaultHandler
			}
			if handler == "" || md.Opts.Method == "" || md.Opts.Path == "" || md.In == nil || md.In.Incomplete || md.Out == nil || md.Out.Incomplete {
				if skipInvalid {
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
					Name:     md.Name,
					In:       inMsg.Message,
					Bindings: inMsg.bindings,
					Out:      mc.Resolve(md.Out).Message,
					Timeout:  time.Duration(timeout) * time.Millisecond,
				},
			})
		}
	}

	return routes, nil
}

func getTypeKind(ty descriptorpb.FieldDescriptorProto_Type) (jsonpb.Kind, bool) {
	switch ty {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return jsonpb.DoubleKind, true
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return jsonpb.FloatKind, true
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return jsonpb.Int64Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return jsonpb.Uint64Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return jsonpb.Int32Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return jsonpb.Fixed64Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return jsonpb.Fixed32Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return jsonpb.BoolKind, true
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return jsonpb.StringKind, true
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return jsonpb.MessageKind, true
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return jsonpb.BytesKind, true
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return jsonpb.Uint32Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return jsonpb.Sfixed32Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return jsonpb.Sfixed64Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return jsonpb.Sint32Kind, true
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return jsonpb.Sint64Kind, true
	}
	return 0, false
}
