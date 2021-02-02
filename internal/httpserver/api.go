package httpserver

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

var listener net.Listener

func Shutdown()error {
	return listener.Close()
}

func Start()chan error{
	chErr := make(chan error)
	go func() {
		var err error
		cert, err := tls.LoadX509KeyPair( "/etc/letsencrypt/live/floxy.io/fullchain.pem","/etc/letsencrypt/live/floxy.io/privkey.pem" )
		if err != nil {
			return
		}
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			Certificates: []tls.Certificate{cert},
		}

		listener, err = tls.Listen("tcp", "localhost:8443", tlsConfig)
		if err != nil {
			chErr <- err
		}
		log.Println("start listening at port 8443")
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("conn err: ", conn)
				continue
			}
			tlsConn := conn.(*tls.Conn)
			tlsConn.VerifyHostname()

			fmt.Println("remote conn", conn.RemoteAddr().String())
			fmt.Println("local conn", conn.LocalAddr().String())

			conn.Close()
		}
	}()
	return chErr
}