package ioutil

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReadToEnd(t *testing.T) {
	type args struct {
		r        io.Reader
		expected int64
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "empty", args: args{}, want: []byte(``)},
		{name: "limited", args: args{r: strings.NewReader(`abc`), expected: 1}, want: []byte(`a`)},
		{name: "no_alloc", args: args{r: strings.NewReader(`abc`), expected: 3}, want: []byte(`abc`)},
		{name: "oversize", args: args{r: strings.NewReader(`abc`), expected: 3}, want: []byte(`abc`)},
		{name: "unlimited", args: args{r: strings.NewReader(strings.Repeat("a", 1024)), expected: -1}, want: []byte(bytes.Repeat([]byte("a"), 1024))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadToEnd(tt.args.r, tt.args.expected)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadToEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("ReadToEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}
