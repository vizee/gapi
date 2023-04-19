package jsonpb

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/vizee/gapi/encoding/jsonlit"
	"github.com/vizee/gapi/encoding/proto"
	"github.com/vizee/gapi/metadata"
)

func skipJsonValueCase(j string) error {
	it := jsonlit.NewIter([]byte(j))
	tok, _ := it.Next()
	err := skipJsonValue(it, tok)
	if err != nil {
		return err
	}
	if !it.EOF() {
		return fmt.Errorf("incomplete")
	}
	return nil
}

func Test_skipJsonValue(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{name: "array", arg: `[1,"hello",false,{"k1":"v1","k2":"v2"}]`, wantErr: false},
		{name: "object", arg: `{"a":1,"b":"hello","c":[1,"hello",false,{"k1":"v1","k2":"v2"}],"d":{"k1":"v1","k2":"v2"}}`, wantErr: false},
		{name: "ignore_syntax", arg: `{"a" 1 "b" "hello" "c":[1 "hello" false {"k1":"v1","k2":"v2"}],"d":{"k1":"v1","k2":"v2"}}`, wantErr: false},
		{name: "bad_token", arg: `:`, wantErr: true},
		{name: "unterminated", arg: `{"k1":1,"k2":2`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := skipJsonValueCase(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("skipJsonValueCase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func transJsonBytesCase(tag uint32, omitEmpty bool, s []byte) (string, error) {
	var buf proto.Encoder
	err := transJsonBytes(&buf, tag, omitEmpty, s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func Test_transJsonBytes(t *testing.T) {
	type args struct {
		tag       uint32
		omitEmpty bool
		s         []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "empty", args: args{tag: 1, s: []byte(`""`)}, want: "0a00"},
		{name: "omit_empty", args: args{tag: 1, omitEmpty: true, s: []byte(`""`)}, want: ""},
		{name: "simple", args: args{tag: 1, s: []byte(`"aGVsbG8gd29ybGQ="`)}, want: "0a0b68656c6c6f20776f726c64"},
		{name: "illegal_base64", args: args{tag: 1, s: []byte(`"aGVsbG8gd29ybGQ"`)}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transJsonBytesCase(tt.args.tag, tt.args.omitEmpty, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("transJsonBytesCase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("transJsonBytesCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func transJsonStringCase(tag uint32, omitEmpty bool, s []byte) (string, error) {
	var buf proto.Encoder
	err := transJsonString(&buf, tag, omitEmpty, s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func Test_transJsonStringCase(t *testing.T) {
	type args struct {
		tag       uint32
		omitEmpty bool
		s         []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "empty", args: args{tag: 1, s: []byte(`""`)}, want: "0a00"},
		{name: "omit_empty", args: args{tag: 1, omitEmpty: true, s: []byte(`""`)}, want: ""},
		{name: "simple", args: args{tag: 1, s: []byte(`"hello world"`)}, want: "0a0b68656c6c6f20776f726c64"},
		{name: "escape", args: args{tag: 1, s: []byte(`"\u4f60\u597d"`)}, want: "0a06e4bda0e5a5bd"},
		{name: "illegal_escape", args: args{tag: 1, s: []byte(`"\z"`)}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transJsonStringCase(tt.args.tag, tt.args.omitEmpty, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("transJsonStringCase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("transJsonStringCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func transJsonNumericCase(tag uint32, kind metadata.Kind, s []byte) (string, error) {
	var buf proto.Encoder
	err := transJsonNumeric(&buf, tag, kind, s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func Test_transJsonNumeric(t *testing.T) {
	type args struct {
		tag  uint32
		kind metadata.Kind
		s    []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "omit_zero", args: args{tag: 1, kind: metadata.Int32Kind, s: []byte(`0`)}, want: ""},
		{name: "double", args: args{tag: 1, kind: metadata.DoubleKind, s: []byte(`1`)}, want: "09000000000000f03f"},
		{name: "float", args: args{tag: 2, kind: metadata.FloatKind, s: []byte(`2`)}, want: "1500000040"},
		{name: "int32", args: args{tag: 3, kind: metadata.Int32Kind, s: []byte(`3`)}, want: "1803"},
		{name: "int64", args: args{tag: 4, kind: metadata.Int64Kind, s: []byte(`4`)}, want: "2004"},
		{name: "uint32", args: args{tag: 5, kind: metadata.Uint32Kind, s: []byte(`5`)}, want: "2805"},
		{name: "uint64", args: args{tag: 6, kind: metadata.Uint64Kind, s: []byte(`6`)}, want: "3006"},
		{name: "sint32", args: args{tag: 7, kind: metadata.Sint32Kind, s: []byte(`7`)}, want: "380e"},
		{name: "sint64", args: args{tag: 8, kind: metadata.Sint64Kind, s: []byte(`8`)}, want: "4010"},
		{name: "fixed32", args: args{tag: 9, kind: metadata.Fixed32Kind, s: []byte(`9`)}, want: "4d09000000"},
		{name: "fixed64", args: args{tag: 10, kind: metadata.Fixed64Kind, s: []byte(`10`)}, want: "510a00000000000000"},
		{name: "sfixed32", args: args{tag: 11, kind: metadata.Sfixed32Kind, s: []byte(`11`)}, want: "5d0b000000"},
		{name: "sfixed64", args: args{tag: 12, kind: metadata.Sfixed64Kind, s: []byte(`12`)}, want: "610c00000000000000"},
		{name: "invalid_kind", args: args{tag: 1, kind: metadata.BoolKind, s: []byte(`1`)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transJsonNumericCase(tt.args.tag, tt.args.kind, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("transJsonNumericCase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("transJsonNumericCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
