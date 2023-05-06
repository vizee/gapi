package engine

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestContext_Get(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		values map[string]string
		args   args
		want   string
		want1  bool
	}{
		{name: "empty", values: map[string]string{}, args: args{name: "key"}, want1: false},
		{name: "simple", values: map[string]string{"key": "value"}, args: args{name: "key"}, want: "value", want1: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{}
			for k, v := range tt.values {
				ctx.Set(k, v)
			}
			got, got1 := ctx.Get(tt.args.name)
			if got != tt.want {
				t.Errorf("Context.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Context.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestContext_reset(t *testing.T) {
	ctx := &Context{
		req:    &http.Request{},
		resp:   nil,
		values: map[string]string{"key": "value"},
		params: Params{},
		chain:  []HandleFunc{func(ctx *Context) error { return nil }},
		handle: func(ctx *Context) error {
			return nil
		},
		next: 1,
	}
	ctx.reset()
	if !reflect.DeepEqual(ctx, &Context{}) {
		t.Fatal(ctx)
	}
}

func TestContext_Next(t *testing.T) {
	type fields struct {
		chain  []HandleFunc
		handle HandleFunc
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "handle", fields: fields{
			handle: func(ctx *Context) error {
				fmt.Println("handle")
				return nil
			},
		}},
		{name: "chain", fields: fields{
			chain: []HandleFunc{
				func(ctx *Context) error {
					fmt.Println("chain 0")
					return ctx.Next()
				},
				func(ctx *Context) error {
					fmt.Println("chain 1")
					return ctx.Next()
				},
				func(ctx *Context) error {
					fmt.Println("chain 2")
					return ctx.Next()
				},
			},
			handle: func(ctx *Context) error {
				fmt.Println("handle")
				return ctx.Next()
			},
		}},
		{name: "break", fields: fields{
			chain: []HandleFunc{
				func(ctx *Context) error {
					fmt.Println("chain 0")
					return ctx.Next()
				},
				func(ctx *Context) error {
					fmt.Println("chain 1")
					return nil
				},
				func(ctx *Context) error {
					fmt.Println("chain 2")
					return ctx.Next()
				},
			},
			handle: func(ctx *Context) error {
				fmt.Println("handle")
				return ctx.Next()
			},
		}},
		{name: "error", fields: fields{
			chain: []HandleFunc{
				func(ctx *Context) error {
					fmt.Println("chain 0")
					return ctx.Next()
				},
				func(ctx *Context) error {
					fmt.Println("chain 1")
					return errors.New("error")
				},
				func(ctx *Context) error {
					fmt.Println("chain 2")
					return ctx.Next()
				},
			},
			handle: func(ctx *Context) error {
				fmt.Println("handle")
				return ctx.Next()
			},
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				chain:  tt.fields.chain,
				handle: tt.fields.handle,
			}
			if err := c.Next(); (err != nil) != tt.wantErr {
				t.Errorf("Context.Next() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
