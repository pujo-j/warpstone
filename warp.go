package warpstone

import "github.com/cloudflare/circl/dh/sidh"

type Conn interface {
	Send(msg []byte) error
	Recv() ([]byte, error)
}

type ServerCrypto struct {
	PSK [32]byte
	Key sidh.PrivateKey
	Pub sidh.PublicKey
	Kem *sidh.KEM
}

type ClientCrypto struct {
	PSK [32]byte
	Pub sidh.PublicKey
	Kem *sidh.KEM
}

// errorString is a trivial implementation of error.
type errorString string

func (e errorString) Error() string {
	return string(e)
}

const ErrInvalidConfig = errorString("invalid crypto config")
const ErrCryptoNegociation = errorString("error in negociating crypto stream")
