package server

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
)

type methodType struct {
	method    reflect.Method
	argType   reflect.Type
	replyType reflect.Type
}

type service struct {
	sync.Mutex
	rcvr   reflect.Value          // receiver.go of methods for the service
	typ    reflect.Type           // type of the receiver.go
	method map[string]*methodType // registered methods
	obj    interface{}
}

func (s *service) call(ctx context.Context, method string, argv []byte) (rep []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[service internal error]: %v, method: %s, argv: %+v", r, method, argv)
			log.Printf("call failed: %+v \n", err)
		}
	}()

	md := s.method[method]
	if md == nil {
		return nil, fmt.Errorf("method[%s] not exist.", method)
	}

	function := md.method.Func

	//log.Printf("call method:%s argType:%v\n", method, md.argType.Kind())
	var retValues []reflect.Value
	kType := md.argType.Kind()
	if kType == reflect.Struct || kType == reflect.Ptr {
		argObj := reflect.New(md.argType.Elem()).Interface()
		Unmarshal(argv, &argObj)
		retValues = function.Call([]reflect.Value{s.rcvr, reflect.ValueOf(ctx), reflect.ValueOf(argObj)})
	} else {
		if kType == reflect.String {
			retValues = function.Call([]reflect.Value{s.rcvr, reflect.ValueOf(ctx), reflect.ValueOf((string)(argv))})
		} else {
			retValues = function.Call([]reflect.Value{s.rcvr, reflect.ValueOf(ctx), reflect.ValueOf(argv)})
		}
	}

	//log.Println("retValues: ", retValues)
	// The return value of the method is a string or []byte or *struct
	reply := retValues[0].Interface()
	buf, ok := reply.([]byte)
	if ok {
		rep = buf
	} else {
		str, ok := reply.(string)
		if ok {
			rep = ([]byte)(str)
		} else {
			rep, _ = Marshal(reply)
		}
	}

	// The return value of the method is an error.
	errRet := retValues[1].Interface()
	if errRet != nil {
		return rep, errRet.(error)
	}

	return rep, nil
}
