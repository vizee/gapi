package jsonpb

import "github.com/vizee/gapi/encoding/jsonlit"

type JsonBuilder struct {
	buf []byte
}

func (b *JsonBuilder) Len() int {
	return len(b.buf)
}

func (b *JsonBuilder) Clear() {
	b.buf = b.buf[:0]
}

func (b *JsonBuilder) Reserve(n int) {
	if cap(b.buf)-len(b.buf) < n {
		newbuf := make([]byte, len(b.buf), cap(b.buf)+n)
		copy(newbuf, b.buf)
		b.buf = newbuf
	}
}

func (b *JsonBuilder) String() string {
	return asString(b.buf)
}

func (b *JsonBuilder) IntoBytes() []byte {
	buf := b.buf
	b.buf = nil
	return buf
}

func (b *JsonBuilder) AppendBytes(s []byte) {
	b.buf = append(b.buf, s...)
}

func (b *JsonBuilder) AppendString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *JsonBuilder) AppendByte(c byte) {
	b.buf = append(b.buf, c)
}

func (b *JsonBuilder) AppendEscapedString(s string) {
	b.buf = jsonlit.EscapeString(b.buf, s)
}
