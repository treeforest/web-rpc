package server

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"unicode"
	"unicode/utf8"
)

// return data of rpc
type returnData struct {
	Body []byte
	Err  string
}

// Precompute the reflect type of error.
// Can't use error directly,because TypeOf takes an empty interface value.This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// Precompute the reflect type of context.
var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

func suitableMethods(typ reflect.Type) map[string]*methodType {
	methods := make(map[string]*methodType)

	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		log.Printf("suitable %s %v \n", mname, mtype)

		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		// Method needs three input parameters ins: receiver, context.Context, *args.
		if mtype.NumIn() != 3 {
			log.Printf("method %s has wong number of ins: %v \n", mname, mtype.NumIn())
			continue
		}

		// First arg must be context.Context
		ctxType := mtype.In(1)
		if !ctxType.Implements(typeOfContext) {
			log.Printf("method %s must usr context.Context as the first parameter.\n", mname)
			continue
		}

		// Second arg need not be a pointer.
		argType := mtype.In(2)
		if !isExportedOrBuiltinType(argType) {
			log.Printf("%s's parameter type not exported: %s \n", mname, argType.String())
			continue
		}

		// Method needs two output parameters ins: *reply, error
		if mtype.NumOut() != 2 {
			log.Printf("method %s has wrong number of output parameters: %v \n", mname, mtype.NumOut())
			continue
		}

		// Reply must be a pointer.
		replyType := mtype.Out(0)
		if replyType.Kind() != reflect.Ptr && replyType.Kind() != reflect.String && replyType.Kind() != reflect.Slice {
			log.Printf("method %s reply type not a pointer. \n", mname)
			continue
		}

		// Replay must be exported
		if !isExportedOrBuiltinType(replyType) {
			log.Printf("method %s reply type not exported: %s \n", mname, replyType.String())
			continue
		}

		// The return type of the method must be error.
		if returnType := mtype.Out(1); returnType != typeOfError {
			log.Printf("method %s returns %s not type of error. \n", mname, returnType.String())
		}

		methods[mname] = &methodType{method: method, argType: argType, replyType: replyType}
	}

	return methods
}

// encode
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// decode
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
