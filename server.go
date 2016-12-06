package netter

import (
	"net"
)

type Server struct {
	listener net.Listener

	channelSize int

	handler Handler

	protocol Protocol
}

// build new server instance
func BuildServer(listener net.Listener, protocol Protocol, handler Handler, channelSize int) *Server {
	return &Server{
		listener: listener,
		protocol: protocol,
		handler:  handler,
		protocol: protocol,
	}
}

func Listen(network, address string, protocol Protocol, handler Handler, channelSize int) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return BuildServer(listener, protocol, handler, channelSize), nil
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Accept(callback func(*Session)) error {
	defer s.Close()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return err
			}
		}
		session := BuildSession(conn, s.protocol, s.handler, s.channelSize)
		callback(session)
	}
}
