package rpc

import "context"

type Server interface {
	Start()
}

type Client interface {
	CallByObj(ctx context.Context, sname, name string, in, out interface{}) error
	Call(ctx context.Context, sname, name string, in interface{}) (rbuf []byte, err error)
}
