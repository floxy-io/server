package tcpcli

import (
	"fmt"
	"net"
)

func Start() chan bool {
	chStart := make(chan bool)

	go func() {
		ln, err := net.Listen("tcp", ":8085")
		if err != nil {
			panic(err)
		}

		conn, err := ln.Accept()
		fmt.Println(conn)

	}()

	return chStart
}
