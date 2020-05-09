package warpstone

import (
	"errors"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type WSConn struct {
	c *websocket.Conn
}

func (W *WSConn) Close() error {
	return W.c.Close()
}

func (W *WSConn) Send(msg []byte) error {
	return W.c.WriteMessage(websocket.BinaryMessage, msg)
}

func (W *WSConn) Recv() ([]byte, error) {
	t, bytes, err := W.c.ReadMessage()
	if err != nil {
		return nil, err
	}
	if t != websocket.BinaryMessage {
		return nil, errors.New("invalid message type")
	}
	return bytes, nil
}

func Dial(crypto *ClientCrypto, url string) (Conn, error) {
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return crypto.Wrap(&WSConn{clientConn})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

func Listen(crypto *ServerCrypto, handler func(conn Conn)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.WithError(err).Warn("upgrading ws connection")
			return
		}
		conn, err := crypto.Wrap(&WSConn{c: c})
		if err != nil {
			log.WithError(err).Warn("cryptowrapping ws connection")
			return
		}
		handler(conn)
	}
}
