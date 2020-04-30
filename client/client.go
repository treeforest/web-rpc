package client

import (
	"context"
	"github.com/treeforest/rpc/codec"
	"github.com/treeforest/rpc"
	"log"
	"net/http"
	"strings"
	"io/ioutil"
	"fmt"
)

type client struct {
	cc codec.Codec
}

func NewClient(cc codec.Codec) rpc.Client{
	c := new(client)
	c.cc = cc
	return c
}

func (c *client) CallByObj(ctx context.Context, sname, name string, in, out interface{}) error{
	buf, err := c.Call(ctx, sname, name, in)

	pout, ok := out.(*[]byte)
	if ok {
		pout = &buf
		if pout == nil {
			log.Println("not out data.")
		}
	} else {
		if err = c.cc.Unmarshal(buf, out); err != nil {
			log.Println("Unmarshal error: ", err)
		}
	}

	return err
}

func (c *client) Call(ctx context.Context, sname, name string, in interface{})(rbuf []byte, err error) {
	buf, ok := in.([]byte)
	if !ok {
		str, ok := in.(string)
		if ok {
			buf = ([]byte)(str)
		} else {
			buf, err = c.cc.Marshal(in)
			if err != nil {
				log.Println("in error: ", err)
				return rbuf, err
			}
		}
	}

	req, err := http.Post("http://"+sname+"/?mn="+name,
		"application/x-www-form-urlencoded", strings.NewReader("body="+(string)(buf)))
	if err != nil {
		return rbuf, err
	}

	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return rbuf, err
	}

	rb := strings.Split((string)(body), "\n\r\n")
	if rb[0] != "" {
		return rbuf, fmt.Errorf(rb[0])
	}

	return ([]byte)(rb[1]), nil
}
