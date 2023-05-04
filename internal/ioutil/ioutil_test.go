package ioutil

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReadLimited(t *testing.T) {
	type args struct {
		r     io.Reader
		n     int64
		limit int64
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "empty", args: args{}, want: []byte(``)},
		{name: "limited", args: args{r: strings.NewReader(`abc`), limit: 1}, want: []byte(`a`)},
		{name: "no_alloc", args: args{r: strings.NewReader(`abc`), n: 3, limit: 3}, want: []byte(`abc`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadLimited(tt.args.r, tt.args.n, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadLimited() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("ReadLimited() = %v, want %v", got, tt.want)
			}
		})
	}
}
