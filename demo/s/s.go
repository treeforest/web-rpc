package main

import (
	"context"
	"log"
	"github.com/treeforest/rpc/server"
)

type Data struct {
	A int
	B string
}

type Handler struct {}

func (p *Handler) Hello(ctx context.Context, in []byte)(out []byte, err error) {
	log.Println("Hello ", in)
	return in, nil
}

func (p *Handler) Hello2(ctx context.Context, in *Data)(out *Data, err error) {
	log.Println("Hello2 ", in.A, in.B)
	in.A = in.A + 100
	return in, nil
}

func main()  {
	s := server.NewServer(888, new(Handler))
	s.Start()
}