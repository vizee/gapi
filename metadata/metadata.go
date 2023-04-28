package metadata

import (
	"time"

	"github.com/vizee/jsonpb"
)

type BindType uint32

const (
	BindDefault BindType = iota
	BindQuery
	BindParams
	BindHeader
	BindContext
)

type FieldBinding struct {
	Name string
	Kind jsonpb.Kind
	Tag  uint32
	Bind BindType
}

type Message struct {
	*jsonpb.Message
	Bindings []FieldBinding // 只有顶级消息支持 Bindings
}

type Call struct {
	Server  string
	Handler string
	Name    string
	In      *Message
	Out     *Message
	Timeout time.Duration
}

type Route struct {
	Method string
	Path   string
	Use    []string
	Call   *Call
}
