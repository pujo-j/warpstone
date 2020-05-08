package warpstone

type memoryChannel struct {
	c1 chan []byte
	c2 chan []byte
}

func newMemoryChannel() (Conn, Conn) {
	c := memoryChannel{
		c1: make(chan []byte, 1),
		c2: make(chan []byte, 1),
	}
	return &c, c.Reverse()
}

func (m *memoryChannel) Send(msg []byte) error {
	m.c1 <- msg
	return nil
}

func (m *memoryChannel) Recv() ([]byte, error) {
	b := <-m.c2
	return b, nil
}

func (m *memoryChannel) Reverse() Conn {
	return &memoryChannel{
		c1: m.c2,
		c2: m.c1,
	}
}
