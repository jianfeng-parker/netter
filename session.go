package netter

import (
	"errors"
	"net"
	"sync/atomic"
)

const (
	ErrStoped             = errors.New("netter:session stoped")
	ErrSendChannelBlocing = errors.New("netter: send channel blocking")
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
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		s.conn.Close()
		close(s.stopedChannel)
		if s.closeCallback != nil {
			s.closeCallback(s)
		}
	}
	return nil
}

func (s *Session) SetCloseCallback(callback func(*Session)) {
	s.closeCallback = callback
}

func (s *Session) SetSendCallback(callback func(*Session)) {
	s.sendCallback = callback
}

func (s *Session) SetHandler(handler Handler) {
	s.handler = handler
}

func (s *Session) SetProtocol(protocol Protocol) {
	s.protocol = protocol
}

func (s *Session) SetSendChannel(size int) {
	s.sendChannel = make(chan interface{}, size)
}

func (s *Session) GetSendChannelSize() int {
	return cap(s.sendChannel)
}

func (s *Session) Start() {
	if atomic.CompareAndSwapInt32(&s.closed, -1, 0) {
		go s.receiveLoop()
		go s.sendLoop()
	}
}

func (s *Session) AyncSend(packet interface{}) error {
	select {
	case s.sendChannel <- packet:
	case <-s.stopedChannel:
		return ErrStoped
	default:
		return ErrSendChannelBlocing
	}
	return nil
}

func (s *Session) receiveLoop() {
	defer s.Close()
	var buff []byte
	var packet interface{}
	var err error
	for {
		packet, buff, err = s.protocol.Read(s.con, buff)
		if err != nil {
			break
		}
		s.handler(s, packet)
	}
}

func (s *Session) sendLoop() {
	defer s.Close()
	var buff []byte
	var err error
	for {
		select {
		case packet, ok := <-s.sendChannel:
			{
				if !ok {
					return
				}
				if buff, err := s.protocol.BuildPacket(packet, buff); err == nil {
					err = s.protocol.Write(s.conn, buff)
				}
				if err != nil {
					return
				}
				if s.sendCallback != nil {
					s.sendCallback(s, packet)
				}
			}
		case <-s.stopedChannel:
			{
				return
			}
		}
	}
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
