package main

import (
	"github.com/pujo-j/warpstone"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
)
import log "github.com/sirupsen/logrus"

var rootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
		var err error
		var crypto *warpstone.ServerCrypto
		file, err := ioutil.ReadFile("warpstone.json")
		if err != nil {
			crypto, err = warpstone.NewServerCrypto()
			if err != nil {
				log.WithError(err).Panic("generating crypto parameters")
			}
			file, err = crypto.Save()
			if err != nil {
				log.WithError(err).Panic("saving crypto parameters")
			}
			err = ioutil.WriteFile("warpstone.json", file, 0600)
			if err != nil {
				log.WithError(err).Panic("saving crypto parameters")
			}
		} else {
			crypto, err = warpstone.LoadServerCrypto(file)
			if err != nil {
				log.WithError(err).Panic("loading crypto parameters from warpstone.json")
			}
		}
		chanHandler := func(conn warpstone.Conn) {
			remoteConn, err := warpstone.ConnToStream(conn)
			if err != nil {
				log.WithError(err).Error("converting remote connection to net.Conn")
				return
			}
			localConn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
				Port: 4222,
			})
			if err != nil {
				log.WithError(err).Error("connecting to local nats server")
				return
			}
			stop := make(chan bool, 1)
			go func() {
				buf := make([]byte, 8192)
				for {
					read, err := localConn.Read(buf)
					if err != nil {
						log.WithError(err).Warn("reading from local")
						stop <- false
						return
					}
					if log.IsLevelEnabled(log.DebugLevel) {
						log.WithField("data", string(buf[:read])).Debug("read from nats")
					}
					_, err = remoteConn.Write(buf[:read])
					if err != nil {
						log.WithError(err).Warn("writing to remote")
						stop <- false
						return
					}
				}
			}()
			go func() {
				buf := make([]byte, 8192)
				for {
					read, err := remoteConn.Read(buf)
					if err != nil {
						log.WithError(err).Warn("reading from remote")
						stop <- false
						return
					}
					if log.IsLevelEnabled(log.DebugLevel) {
						log.WithField("data", string(buf[:read])).Debug("read from remote")
					}
					_, err = localConn.Write(buf[:read])
					if err != nil {
						log.WithError(err).Warn("writing to nats")
						stop <- false
						return
					}
				}
			}()
			<-stop
		}
		http.HandleFunc("/", warpstone.Listen(crypto, chanHandler))
		log.Info("listening on http:" + strconv.Itoa(port))
		err = http.ListenAndServe("127.0.0.1:"+strconv.Itoa(port), nil)
		if err != nil {
			log.WithError(err).Error("listening on http")
		}
	},
}

func main() {
	_ = rootCmd.Execute()
}

var verbose = false
var port = 5678

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 5678, "local port to listen to for ws connections")
}
