package netter

/**
 * 由使用方调用
 */

import (
	"io"
	"net"
)

// 服务端调用获取一个Server对象
func Listen(network, address string, protocol Protocol, handler Handler, channelSize int) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, protocol, handler, channelSize), nil
}

// 客户端调用获取一个Session对象
func Dial(network, address string, protocol Protocol, sendChannelSize int) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, protocol, sendChannelSize), nil
}

// 由调用方实现
type Protocol interface {
	Reader
	Writer
}

type Reader interface {
	// 从Conn中读取数据
	Read(conn net.Conn) (interface{}, error)
}

type Writer interface {
	// 向连接中写入数据
	Write(conn net.Conn, msg interface{}) error
}

type Handler interface {
	Handle(*Session)
}

type DefaultHandler func(*Session)

func (h DefaultHandler) Handle(s *Session) {
	h.Handle(s)
}
