package jsonpb

import (
	"bytes"
	"testing"
)

func Test_asBytes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "empty", args: args{s: ""}, want: nil},
		{name: "abc", args: args{s: "abc"}, want: []byte("abc")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asBytes(tt.args.s); !bytes.Equal(got, tt.want) {
				t.Errorf("asBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bytesView(t *testing.T) {
	type args struct {
		s []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{s: []byte("")}, want: ""},
		{name: "abc", args: args{s: []byte("abc")}, want: "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesView(tt.args.s); got != tt.want {
				t.Errorf("bytesView() = %v, want %v", got, tt.want)
			}
		})
	}
}
