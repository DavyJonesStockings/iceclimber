package rpc

import (
	"fmt"
	"log"
	"net"
	"reflect"

	msgpackrpc "github.com/msgpack-rpc/msgpack-rpc-go/rpc"
)

type methodResolver struct {
	receiver reflect.Value
}

func (r *methodResolver) Resolve(name string, args []reflect.Value) (reflect.Value, error) {
	method := r.receiver.MethodByName(name)
	if !method.IsValid() {
		return reflect.Value{}, fmt.Errorf("unknown method: %s", name)
	}
	return method, nil
}

func Start(cb func(Event)) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	handler := &Server{cb: cb}
	resolver := &methodResolver{receiver: reflect.ValueOf(handler)}

	server := msgpackrpc.NewServer(resolver, true, nil)
	server.Listen(ln)

	log.Println("RPC listening on", ln.Addr())

	server.Run()
}
