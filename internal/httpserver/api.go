package httpserver

import (
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
		listener, err = net.Listen("tcp", "localhost:3333")
		if err != nil {
			chErr <- err
		}
		log.Println("start listening at port 3333")
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("conn err: ", conn)
				continue
			}
			fmt.Println("remote conn", conn.RemoteAddr().String())
			fmt.Println("local conn", conn.LocalAddr().String())

			conn.Close()
		}
	}()
	return chErr
}