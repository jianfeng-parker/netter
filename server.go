package netter

import (
	"io"
	"net"
	"strings"
	"time"
)

type Server struct {
	listener net.Listener

	sendChannelSize int

	handler Handler

	protocol Protocol
}

// build new server instance
func NewServer(listener net.Listener, protocol Protocol, handler Handler, sendChannelSize int) *Server {
	return &Server{
		listener:        listener,
		protocol:        protocol,
		handler:         handler,
		sendChannelSize: sendChannelSize,
	}
}

func (s *Server) Stop() error {
	return s.listener.Close()
}

func (s *Server) Start() error {
	for {
		conn, err := s.accept()
		if err != nil {
			return err
		}
		go func() {
			// 每Accept一个连接后就生成对应的Session对象
			session := NewSession(conn, s.protocol, s.sendChannelSize)
			// TODO 将session管控起来
			// ...
			// TODO 使用handler处理连接
			s.handler.Handle()
		}()
	}
}

// 轮询监听获取新的连接
func (s *Server) accept() (net.Conn, error) {
	var tempDelay time.Duration
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil, io.EOF
			}
			return nil, err
		}
		return conn, nil
	}
}
