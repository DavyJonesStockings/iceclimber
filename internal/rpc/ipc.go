package rpc

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
