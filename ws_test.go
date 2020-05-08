package warpstone

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"testing"
)

func TestWSE2E(t *testing.T) {
	s, err := NewServerCrypto()
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	scc := serverCryptoConfig{}
	err = json.Unmarshal(bytes, &scc)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := LoadServerCrypto(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if c2.PSK != s.PSK {
		t.Error("PSK invalid")
	}
	ccc := clientCryptoConfig{
		PublicKey: scc.PublicKey,
		PSK:       scc.PSK,
	}
	cccb, err := json.MarshalIndent(ccc, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadClientCrypto(cccb)
	if err != nil {
		t.Fatal(err)
	}
	ch := func(c Conn) {
		recv, err := c.Recv()
		if err != nil {
			log.WithError(err).Error("Receiving from client")
			return
		}
		err = c.Send([]byte("Hello " + string(recv)))
		if err != nil {
			log.WithError(err).Error("Responding to client")
			return
		}
	}
	http.HandleFunc("/ws", Listen(s, ch))
	go func() {
		log.Info("listening on http")
		err := http.ListenAndServe("127.0.0.1:8579", nil)
		if err != nil {
			log.WithError(err).Error("listening on http")
		}
	}()
	dial, err := Dial(c, "ws://127.0.0.1:8579/ws")
	if err != nil {
		t.Fatal(err)
	}
	err = dial.Send([]byte("World"))
	if err != nil {
		log.WithError(err).Error("Sending to server")
		t.Fatal(err)
	}
	recv, err := dial.Recv()
	if err != nil {
		log.WithError(err).Error("Receiving from server")
		t.Fatal(err)
	}
	log.WithField("message", string(recv)).Info("Received answer from server")
}
