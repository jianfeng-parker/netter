package netter

import (
	"errors"
	"net"
	"sync/atomic"
)

type Session struct {
	closed        int32
	conn          net.Conn
	sendChannel   chan interface{}
	stopedChannel chan struct{}
	closeCallback func(*Session)
	sendCallback  func(*Session, interface{})
	handler       Handler
	protocol      Protocol
}

func (s *Session) RawConn() net.Conn {
	return s.conn
}

func (s *Session) Close() error {

	return nil
}

func BuildSession(conn net.Conn, protocol Protocol, handler Handler, sendChannelSize int) *Session {
	return &Session{
		closed:        -1,
		conn:          conn,
		sendChannel:   make(chan interface{}, sendChannelSize),
		stopedChannel: make(chan struct{}),
		handler:       handler,
		protocol:      protocol,
	}
}

func Dial(network, address string, handler Handler, protocol Protocol, sendChannelSize int) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return BuildSession(conn, protocol, handler, sendChannelSize), nil
}

type Handler func(s *Session, packet interface{})

// 声明 reader接口，具体由用户自己实现
type Reader interface {
	// 从Conn中读取数据，并创建一个完整的Packet
	Read(conn net.Conn, buff []byte) (interface{}, []byte, error)
}

type Writer interface {
	BuildPacket(packet interface{}, buff []byte) ([]byte, error)

	Write(conn net.Conn, buff []byte) error
}

type Protocol interface {
	Reader
	Writer
}
