package warpstone

import (
	"encoding/json"
	"testing"
)

func TestE2ECrypto(t *testing.T) {
	s, err := NewServerCrypto()
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
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
	t.Log(string(cccb))
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadClientCrypto(cccb)
	if err != nil {
		t.Fatal(err)
	}
	sc, cc := newMemoryChannel()
	done := make(chan string)
	go func() {
		sc2, err := s.Wrap(sc)
		if err != nil {
			done <- "serverError: " + err.Error()
			panic(err)
		}
		err = sc2.Send([]byte("ServerTest"))
		if err != nil {
			done <- "serverError: " + err.Error()
			panic(err)
		}
		recv, err := sc2.Recv()
		if err != nil {
			done <- "serverError: " + err.Error()
			panic(err)
		}
		println(string(recv))
		done <- "serverDone"
	}()
	go func() {
		cc2, err := c.Wrap(cc)
		if err != nil {
			done <- "clientError: " + err.Error()
			panic(err)
		}
		recv, err := cc2.Recv()
		if err != nil {
			done <- "clientError: " + err.Error()
			panic(err)
		}
		println(string(recv))
		err = cc2.Send([]byte("ClientTest"))
		if err != nil {
			done <- "clientError: " + err.Error()
			panic(err)
		}
		done <- "clientDone"
	}()
	println(<-done)
	println(<-done)
}
