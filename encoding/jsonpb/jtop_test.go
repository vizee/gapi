package jsonpb

import (
	"fmt"
	"testing"

	"github.com/vizee/gapi/encoding/jsonlit"
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
