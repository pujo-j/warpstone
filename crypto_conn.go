package warpstone

import "golang.org/x/crypto/nacl/secretbox"
import "crypto/rand"

type cryptoConn struct {
	conn Conn
	sk   [32]byte
}

func (c *cryptoConn) Send(msg []byte) error {
	nonce := [24]byte{}
	_, err := rand.Read(nonce[:])
	if err != nil {
		return err
	}
	out := make([]byte, 24, len(msg)+secretbox.Overhead+24)
	copy(out, nonce[:])
	out = secretbox.Seal(out, msg, &nonce, &c.sk)
	return c.conn.Send(out)
}

func (c *cryptoConn) Recv() ([]byte, error) {
	recv, err := c.conn.Recv()
	if err != nil {
		return nil, err
	}
	if len(recv) < 1+secretbox.Overhead+24 {
		return nil, ErrCryptoNegociation
	}
	nonce := [24]byte{}
	copy(nonce[:], recv[:24])
	out := make([]byte, 0, len(recv)-(24+secretbox.Overhead))
	open, ok := secretbox.Open(out, recv[24:], &nonce, &c.sk)
	if !ok {
		return nil, ErrCryptoNegociation
	}
	return open, nil
}
