package main

import (
	nats "github.com/nats-io/nats.go"
	"github.com/pujo-j/warpstone"
	"io/ioutil"
	"testing"
	"time"
)

func TestGateway(t *testing.T) {
	file, err := ioutil.ReadFile("warpstone.json")
	if err != nil {
		t.Fatal(err)
	}
	crypto, err := warpstone.LoadClientCrypto(file)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := crypto.DialNats("ws://localhost:5678/")
	if err != nil {
		t.Fatal(err)
	}
	conn.SetReconnectHandler(func(conn *nats.Conn) {
		println("Reconnected")
	})
	conn.SetDisconnectErrHandler(func(conn *nats.Conn, err error) {
		println("Disconnected:" + err.Error())
	})
	conn.SetClosedHandler(func(conn *nats.Conn) {
		println("Closed")
	})
	for {
		println("Sending message")
		err = conn.Publish("TEST", []byte("test"))
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Second)
	}
}
