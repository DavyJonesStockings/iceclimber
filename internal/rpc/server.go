package rpc

import (
	"log"
	"net"

	"github.com/msgpack-rpc/msgpack-rpc-go/rpc"
)

type Event struct {
	Widths []int
}

type Server struct {
	cb func(Event)
}

func (s *Server) SendWidths(widths []int) error {
	s.cb(Event{Widths: widths})
	return nil
}

func Start(cb func(Event)) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	server := rpc.NewServer()
	server.Register(&Server{cb: cb})

	log.Println("RPC listening on", ln.Addr())

	server.Accept(ln)
}
