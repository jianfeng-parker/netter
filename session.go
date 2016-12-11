package netter

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var baseSessionId uint64
var ClosedErr = errors.New("session closed")
var BlockingErr = errors.New("send channel blocking")

type Session struct {
	id       uint64
	conn     net.Conn
	protocol Protocol
	closed   int32 // 1->closed

	sendChannel  chan interface{}
	sendMutex    sync.RWMutex
	receiveMutex sync.Mutex

	closeChannel chan int
	closeMutext  sync.Mutex
}

func NewSession(conn net.Conn, protocol Protocol, sendChannelSize int) *Session {
	session := &Session{
		id:           atomic.AddUint64(&baseSessionId, 1),
		conn:         conn,
		protocol:     protocol,
		closeChannel: make(chan int),
	}
	if sendChannelSize > 1 {
		session.sendChannel = make(chan interface{}, sendChannelSize)
	}
	return session
}

func (s *Session) Close() error {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.closeChannel)
		if s.sendChannel != nil {
			s.sendMutex.Lock()
			close(s.sendChannel)
			s.sendMutex.Unlock()
		}
		err := s.conn.Close()
		// TODO 此处可以执行关闭session时的回调
		return err
	}
	return ClosedErr
}

func (s *Session) Receive() (interface{}, error) {
	s.receiveMutex.Lock()
	defer s.receiveMutex.Unlock()
	msg, err := s.protocol.Read(s.conn)
	if err != nil {
		s.Close()
	}
	return msg, err
}

func (s *Session) Send(msg interface{}) error {
	if s.sendChannel == nil {
		if s.IsClosed() {
			return ClosedErr
		}
		s.sendMutex.Lock()
		defer s.sendMutex.Unlock()
		err := s.protocol.Write(s.conn, msg)
		if err != nil {
			s.Close()
		}
		return err
	}
	s.sendMutex.RLock()
	if s.IsClosed() {
		return ClosedErr
	}
	select {
	case s.sendChannel <- msg:
		s.sendMutex.RUnlock()
		return nil
	default:
		s.sendMutex.RUnlock()
		s.Close()
		return BlockingErr
	}

}

func (s *Session) ID() uint64 {
	return s.id
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) == 1
}

func (s *Session) Protocol() Protocol {
	return s.protocol
}

// func (s *Session) ReceiveLoop() {
// 	defer s.Close()
// 	var buff []byte
// 	var msg interface{}
// 	var err error
// 	for {
// 		msg, err = s.protocol.Read(s.con)
// 		if err != nil {
// 			break
// 		}
// 		s.handler(s, packet)
// 	}
// }

func (s *Session) SendLoop() {
	defer s.Close()
	for {
		select {
		case msg, ok := <-s.sendChannel:
			if !ok || s.protocol.Write(s.conn, msg) != nil {
				return
			}
		case <-s.closeChannel:
			return
		}
	}
}
