package warpstone

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/cloudflare/circl/dh/sidh"
)

type clientCryptoConfig struct {
	PublicKey string
	PSK       string
}

func (c *ClientCrypto) Save() ([]byte, error) {
	cc := clientCryptoConfig{}
	cc.PSK = hex.EncodeToString(c.PSK[:])
	buf := make([]byte, c.Pub.Size())
	c.Pub.Export(buf)
	cc.PublicKey = hex.EncodeToString(buf)
	return json.MarshalIndent(cc, "", " ")
}

func LoadClientCrypto(data []byte) (*ClientCrypto, error) {
	cc := clientCryptoConfig{}
	key := sidh.NewPublicKey(sidh.Fp751, sidh.KeyVariantSike)
	res := ClientCrypto{
		PSK: [32]byte{},
		Pub: *key,
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

func (c *ClientCrypto) Wrap(conn Conn) (Conn, error) {
	sidhSecret := make([]byte, c.Kem.SharedSecretSize())
	sidhCipher := make([]byte, c.Kem.CiphertextSize())
	err := c.Kem.Encapsulate(sidhCipher, sidhSecret, &c.Pub)
	if err != nil {
		return nil, err
	}
	err = conn.Send(sidhCipher)
	if err != nil {
		return nil, err
	}
	h := sha512.New()
	h.Write(sidhSecret)
	h.Write(c.PSK[:])
	skb := h.Sum(nil)
	sk := [32]byte{}
	copy(sk[:], skb)
	return &cryptoConn{
		conn: conn,
		sk:   sk,
	}, nil
}
