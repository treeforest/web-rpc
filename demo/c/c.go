package main

import (
	"context"
	"github.com/treeforest/rpc/client"
	"github.com/treeforest/rpc/codec/jsoncodec"
	"log"
)

type Data struct {
	A int
	B string
}

const sname string = "127.0.0.1:888"

func main() {
	c := client.NewClient(jsoncodec.NewCodec())
	out1, err := c.Call(context.Background(), sname, "Hello", []byte("qwer"))
	log.Println("call 1", (string)(out1), err)

	out2 := new(Data)
	err = c.CallByObj(context.Background(), sname, "Hello2", &Data{99, "Tony"}, out2)
	log.Println("call 2", out2, err)
}
