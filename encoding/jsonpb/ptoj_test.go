package jsonpb

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/vizee/gapi/encoding/proto"
	"github.com/vizee/gapi/metadata"
	"google.golang.org/protobuf/encoding/protowire"
)

func decodeBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func readProtoValueCase(s string, wire protowire.Type) (protoValue, int) {
	return readProtoValue(proto.NewDecoder(decodeBytes(s)), wire)
}

func Test_readProtoValueCase(t *testing.T) {
	type args struct {
		s    string
		wire protowire.Type
	}
	tests := []struct {
		name  string
		args  args
		want  protoValue
		want1 int
	}{
		{name: "varint", args: args{s: "7b", wire: protowire.VarintType}, want: protoValue{x: 123}},
		{name: "fixed32", args: args{s: "7b000000", wire: protowire.Fixed32Type}, want: protoValue{x: 123}},
		{name: "fixed64", args: args{s: "7b00000000000000", wire: protowire.Fixed64Type}, want: protoValue{x: 123}},
		{name: "bytes", args: args{s: "036f6b6b", wire: protowire.BytesType}, want: protoValue{s: []byte("okk")}},
		{name: "bad_wire", args: args{s: "", wire: protowire.StartGroupType}, want: protoValue{}, want1: -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := readProtoValueCase(tt.args.s, tt.args.wire)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readProtoValueCase() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("readProtoValueCase() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func transProtoBytesCase(s string) string {
	var j JsonBuilder
	transProtoBytes(&j, decodeBytes(s))
	return j.String()
}

func Test_transProtoBytesCase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{s: ""}, want: `""`},
		{name: "hello", args: args{s: "68656c6c6f"}, want: `"aGVsbG8="`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transProtoBytesCase(tt.args.s)
			if got != tt.want {
				t.Errorf("transProtoBytesCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func transProtoStringCase(s string) string {
	var j JsonBuilder
	transProtoString(&j, decodeBytes(s))
	return j.String()
}

func Test_transProtoStringCase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{s: ""}, want: `""`},
		{name: "hello", args: args{s: "68656c6c6f"}, want: `"hello"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transProtoStringCase(tt.args.s)
			if got != tt.want {
				t.Errorf("transProtoStringCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func transProtoSimpleValueCase(kind metadata.Kind, s string) string {
	wire := protowire.BytesType
	if int(kind) < len(wireTypeOfKind) {
		wire = wireTypeOfKind[kind]
	}
	pv, _ := readProtoValue(proto.NewDecoder(decodeBytes(s)), wire)
	var j JsonBuilder
	transProtoSimpleValue(&j, kind, pv.x)
	return j.String()
}

func Test_transProtoSimpleValueCase(t *testing.T) {
	type args struct {
		kind metadata.Kind
		s    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "double", args: args{kind: metadata.DoubleKind, s: "ae47e17a14aef33f"}, want: "1.23"},
		{name: "float", args: args{kind: metadata.FloatKind, s: "a4709d3f"}, want: "1.23"},
		{name: "int32", args: args{kind: metadata.Int32Kind, s: "7b"}, want: "123"},
		{name: "int64", args: args{kind: metadata.Int64Kind, s: "7b"}, want: "123"},
		{name: "uint32", args: args{kind: metadata.Uint32Kind, s: "7b"}, want: "123"},
		{name: "uint64", args: args{kind: metadata.Uint64Kind, s: "7b"}, want: "123"},
		{name: "sint32", args: args{kind: metadata.Sint32Kind, s: "f501"}, want: "-123"},
		{name: "sint64", args: args{kind: metadata.Sint64Kind, s: "f501"}, want: "-123"},
		{name: "fixed32", args: args{kind: metadata.Fixed32Kind, s: "7b000000"}, want: "123"},
		{name: "fixed64", args: args{kind: metadata.Fixed64Kind, s: "7b00000000000000"}, want: "123"},
		{name: "sfixed32", args: args{kind: metadata.Sfixed32Kind, s: "85ffffff"}, want: "-123"},
		{name: "sfixed64", args: args{kind: metadata.Sfixed64Kind, s: "85ffffffffffffff"}, want: "-123"},
		{name: "bool_true", args: args{kind: metadata.BoolKind, s: "01"}, want: "true"},
		{name: "bool_false", args: args{kind: metadata.BoolKind, s: "00"}, want: "false"},
		{name: "unexpected_kind", args: args{kind: metadata.StringKind, s: "00"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transProtoSimpleValueCase(tt.args.kind, tt.args.s); got != tt.want {
				t.Errorf("transProtoSimpleValueCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func transProtoRepeatedBytesCase(p string, field *metadata.Field, s string) (string, error) {
	var j JsonBuilder
	err := transProtoRepeatedBytes(&j, proto.NewDecoder(decodeBytes(p)), field, decodeBytes(s))
	if err != nil {
		return "", err
	}
	return j.String(), nil
}

func Test_transProtoRepeatedBytesCase(t *testing.T) {
	type args struct {
		p     string
		field *metadata.Field
		s     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "one_string", args: args{p: "", field: &metadata.Field{Tag: 1, Kind: metadata.StringKind}, s: "616263"}, want: `["abc"]`},
		{name: "more_strings", args: args{p: "0a0568656c6c6f0a05776f726c64", field: &metadata.Field{Tag: 1, Kind: metadata.StringKind}, s: "616263"}, want: `["abc","hello","world"]`},
		{name: "more_bytes", args: args{p: "0a0568656c6c6f0a05776f726c64", field: &metadata.Field{Tag: 1, Kind: metadata.BytesKind}, s: "616263"}, want: `["YWJj","aGVsbG8=","d29ybGQ="]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transProtoRepeatedBytesCase(tt.args.p, tt.args.field, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("transProtoRepeatedBytesCase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("transProtoRepeatedBytesCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
