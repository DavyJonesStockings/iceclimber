package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Server struct {
	cb   func(Event)
	conn net.Conn
}

func Start(cb func(Event)) *Server {
	ln, err := net.Listen("tcp", "127.0.0.1:4545")
	if err != nil {
		log.Fatal(err)
	}
	// using fmt instead of log so that this goes to stdout not stderr
	fmt.Println("ICECLIMBER_READY")

	conn, err := ln.Accept()
	if err != nil {
		log.Println("accept error:", err)
	}
	log.Println("client connected")

	s := &Server{cb: cb, conn: conn}
	go s.readLoop()
	return s

}

func (s *Server) readLoop() {
	scanner := bufio.NewScanner(s.conn)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			log.Println("decode error: ", err)
			continue
		}
		s.cb(event)
		log.Printf("recv event: type=%s top=%d bot=%d lines=%d cursor=%v",
			event.Type, event.Top, event.Bot, len(event.Lines), event.Cursor)
	}
	if err := scanner.Err(); err != nil {
		log.Println("scan error: ", err)
	}
	log.Println("client disconnected")
}

func (s *Server) SendCommand(cmd Command) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = s.conn.Write(data)
	return err
}
