package warpstone

import (
	"net"
	"sync"
	"time"
)

type NetConn struct {
	delegate   Conn
	recvBuffer []byte
	lock       sync.Mutex
}

const maxMessageSize = 1024

func ConnToStream(c Conn) (net.Conn, error) {
	res := &NetConn{
		delegate:   c,
		recvBuffer: nil,
		lock:       sync.Mutex{},
	}
	return res, nil
}

func (m *NetConn) Read(p []byte) (n int, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.recvBuffer == nil || len(m.recvBuffer) == 0 {
		m.recvBuffer, err = m.delegate.Recv()
		if err != nil {
			return 0, err
		}
	}
	if len(m.recvBuffer) > len(p) {
		n := copy(p, m.recvBuffer[:len(p)])
		m.recvBuffer = m.recvBuffer[n:]
		return n, nil
	} else {
		n := copy(p[:len(m.recvBuffer)], m.recvBuffer)
		m.recvBuffer = nil
		return n, nil
	}
}

func (m *NetConn) Write(p []byte) (n int, err error) {
	toSend := p
	sent := 0
	for len(toSend) > 0 {
		if len(toSend) < maxMessageSize {
			err = m.delegate.Send(toSend)
			if err != nil {
				return sent, err
			}
			sent += len(toSend)
			toSend = toSend[:0]
		} else {
			err := m.delegate.Send(toSend[:maxMessageSize])
			if err != nil {
				return sent, err
			}
			sent += maxMessageSize
			toSend = toSend[maxMessageSize:]
		}
	}
	return sent, nil
}

func (m *NetConn) Close() error {
	return m.delegate.Close()
}

type warpAddr struct {
	s string
}

func (w *warpAddr) Network() string {
	return "tcp"
}

func (w *warpAddr) String() string {
	return w.s
}

func (m *NetConn) LocalAddr() net.Addr {
	return &warpAddr{"localhost"}
}

func (m *NetConn) RemoteAddr() net.Addr {
	return &warpAddr{"gateway"}
}

func (m *NetConn) SetDeadline(_ time.Time) error {
	return nil
}

func (m *NetConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (m *NetConn) SetWriteDeadline(_ time.Time) error {
	return nil
}
