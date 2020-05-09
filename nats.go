package warpstone

import (
	nats "github.com/nats-io/nats.go"
	"net"
)

type natsDialer struct {
	c   *ClientCrypto
	url string
}

func (n *natsDialer) Dial(_, _ string) (net.Conn, error) {
	conn, err := Dial(n.c, n.url)
	if err != nil {
		return nil, err
	}
	return ConnToStream(conn)
}

func withWarp(c *ClientCrypto, url string) nats.Option {
	return func(options *nats.Options) error {
		options.CustomDialer = &natsDialer{
			c:   c,
			url: url,
		}
		return nil
	}
}

func (c *ClientCrypto) DialNats(url string, options ...nats.Option) (*nats.Conn, error) {
	options = append(options, withWarp(c, url), nats.DontRandomize())
	return nats.Connect("default", options...)
}
