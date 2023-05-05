package metadata

import (
	"time"

	"github.com/vizee/jsonpb"
)

type BindSource uint32

const (
	BindDefault BindSource = iota
	BindQuery
	BindParams
	BindHeader
	BindContext
)

type FieldBinding struct {
	Name string
	Kind jsonpb.Kind
	Tag  uint32
	Bind BindSource
}

type Call struct {
	Server   string
	Handler  string
	Method   string
	In       *jsonpb.Message
	Out      *jsonpb.Message
	Bindings []FieldBinding // 仅支持从参数提取 Bindings
	Timeout  time.Duration
}

type Route struct {
	Method string
	Path   string
	Use    []string
	Call   *Call
}
