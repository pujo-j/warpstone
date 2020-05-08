package warpstone

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/cloudflare/circl/dh/sidh"
)

type serverCryptoConfig struct {
	PrivateKey string
	PublicKey  string
	PSK        string
}

func (c *ServerCrypto) Save() ([]byte, error) {
	cc := serverCryptoConfig{}
	cc.PSK = hex.EncodeToString(c.PSK[:])
	buf := make([]byte, c.Key.Size())
	c.Key.Export(buf)
	cc.PrivateKey = hex.EncodeToString(buf)
	buf = make([]byte, c.Pub.Size())
	c.Pub.Export(buf)
	cc.PublicKey = hex.EncodeToString(buf)
	return json.MarshalIndent(cc, "", " ")
}

func NewServerCrypto() (*ServerCrypto, error) {
	key := sidh.NewPrivateKey(sidh.Fp751, sidh.KeyVariantSike)
	err := key.Generate(rand.Reader)
	pub := sidh.NewPublicKey(sidh.Fp751, sidh.KeyVariantSike)
	key.GeneratePublicKey(pub)
	if err != nil {
		return nil, err
	}
	psk := [32]byte{}
	_, err = rand.Reader.Read(psk[:])
	if err != nil {
		return nil, err
	}
	return &ServerCrypto{
		PSK: psk,
		Key: *key,
		Pub: *pub,
		Kem: sidh.NewSike751(rand.Reader),
	}, nil
}

func LoadServerCrypto(data []byte) (*ServerCrypto, error) {
	cc := serverCryptoConfig{}
	key := sidh.NewPrivateKey(sidh.Fp751, sidh.KeyVariantSike)
	pub := sidh.NewPublicKey(sidh.Fp751, sidh.KeyVariantSike)
	res := ServerCrypto{
		PSK: [32]byte{},
		Key: *key,
		Pub: *pub,
		Kem: sidh.NewSike751(rand.Reader),
	}
	err := json.Unmarshal(data, &cc)
	if err != nil {
		return nil, err
	}
	bytes, err := hex.DecodeString(cc.PSK)
	if err != nil {
		return nil, err
	}
	if len(bytes) != 32 {
		return nil, ErrInvalidConfig
	}
	copy(res.PSK[:], bytes)
	bytes, err = hex.DecodeString(cc.PrivateKey)
	if err != nil {
		return nil, err
	}
	err = res.Key.Import(bytes)
	if err != nil {
		return nil, err
	}
	bytes, err = hex.DecodeString(cc.PublicKey)
	if err != nil {
		return nil, err
	}
	err = res.Pub.Import(bytes)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *ServerCrypto) Wrap(conn Conn) (Conn, error) {
	recv, err := conn.Recv()
	if err != nil {
		return nil, err
	}
	if len(recv) != s.Kem.CiphertextSize() {
		return nil, ErrCryptoNegociation
	}
	sidhSecret := make([]byte, s.Kem.SharedSecretSize())
	err = s.Kem.Decapsulate(sidhSecret, &s.Key, &s.Pub, recv)
	if len(recv) != s.Kem.CiphertextSize() {
		return nil, ErrCryptoNegociation
	}
	h := sha512.New()
	h.Write(sidhSecret)
	h.Write(s.PSK[:])
	skb := h.Sum(nil)
	sk := [32]byte{}
	copy(sk[:], skb)
	return &cryptoConn{
		conn: conn,
		sk:   sk,
	}, nil
}
