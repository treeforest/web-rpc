package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"github.com/treeforest/rpc"
)

type server struct {
	once    sync.Once
	port    int
	service *service
}

func NewServer(port int, obj interface{}) rpc.Server {
	s := new(server)
	s.port = port
	s.init(obj)
	return s
}

func (s *server) init(obj interface{}) {
	s.register(obj)
}

func (s *server) register(obj interface{}) error {
	if s.service != nil {
		return errors.New("can't register again")
	}

	service := new(service)
	service.obj = obj
	service.typ = reflect.TypeOf(obj)
	service.rcvr = reflect.ValueOf(obj)

	// install the methods
	service.method = suitableMethods(service.typ)

	if len(service.method) == 0 {
		var errstr string

		method := suitableMethods(reflect.PtrTo(service.typ))
		if len(method) != 0 {
			errstr = "rpc.register: type has no exported methods of suitable type. (hint: pass a pointer to value of the type)"
		} else {
			errstr = "rpc.register: type has no exported methods of suitable type."
		}
		log.Println(errstr)
		return errors.New(errstr)
	}

	s.service = service
	return nil
}

func (s *server) Start() {
	s.once.Do(func() {
		http.HandleFunc("/", s.handle)
		log.Printf("Start service: %d \n\n", s.port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
		if err != nil {
			log.Println("ListenAndServe error: ", err)
		}
	})
}

func (s *server) handle(w http.ResponseWriter, req *http.Request) {
	u, _ := url.Parse(req.RequestURI)
	mp, _ := url.ParseQuery(u.RawQuery)
	log.Printf("request method:%v url:%s: \n", mp["mn"], u.String())

	rb := &returnData{}
	typ := ""
	encodeType, ok := mp["ec"]
	if ok {
		typ = encodeType[0]
	}

	method, ok := mp["mn"]
	if !ok {
		rb.Err = fmt.Sprintf("Error Request!\n")
		goto end
	}

	switch typ {
	case "g":
		{
			delete(mp, "sn")
			delete(mp, "ec")
			delete(mp, "mn")

			body, err := Marshal(mp)
			if err != nil {
				rb.Err = fmt.Sprintf("Marshal error: %s", err.Error())
			} else {
				reBody, err := s.service.call(context.Background(), method[0], body)
				log.Println("Call return: ", reBody, err)
				rb.Body = reBody
				if err != nil {
					rb.Err = err.Error()
				}
			}
		}
	case "c":
		{

		}
	default:// json codec
		{
			req.ParseForm()
			values := req.PostForm
			log.Println("post form: ", values)
			pb := values["body"]
			if len(pb) < 1 {
				rb.Err = fmt.Sprintf("Error Request!\n")
				goto end
			}

			reBody, err := s.service.call(context.Background(), method[0], ([]byte)(pb[0]))
			if err != nil {
				rb.Err = err.Error()
			} else {
				rb.Body = reBody
			}
		}
	}

end:
	io.WriteString(w, rb.Err+"\n\r\n")
	w.Write(rb.Body)

	log.Printf("response body:%v err:%v \n\n", rb.Body, rb.Err)
}
