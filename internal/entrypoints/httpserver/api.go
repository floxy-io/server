package httpserver

import (
	"fmt"
	"github.com/danielsussa/floxy/internal/pkg/env"
	"github.com/danielsussa/floxy/internal/pkg/store"
	"io"
	"log"
	"net"
)

var listener net.Listener

func Shutdown() error {
	return listener.Close()
}

func Start(s *store.Engine) chan bool {
	chStart := make(chan bool)
	baseDns := env.GetOrDefault(env.ServerBaseDns, "localhost")
	go func() {
		listener, err := net.Listen("tcp", "localhost:8081")
		if err != nil {
			panic(err)
		}
		chStart <- true
		log.Println("start listening at port 8081")
		for {
			serverConn, err := listener.Accept()
			if err != nil {
				fmt.Println("conn err: ", err)
				continue
			}
			go handleCall(s, serverConn, baseDns)
		}
	}()
	return chStart
}

func handleCall(s *store.Engine, serverConn net.Conn, baseDns string) error {
	defer serverConn.Close()

	info, reader, bRemain := extractHttpInfo(serverConn)

	log.Println(fmt.Sprintf("dns: %s", info.Subdomain()))

	reg, ok := s.Get(info.Subdomain())
	if !ok {
		return fmt.Errorf("user not found")
	}

	clientConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", reg.Port))
	if err != nil {
		return err
	}
	defer clientConn.Close()

	_, err = clientConn.Write(bRemain)
	if err != nil {
		return err
	}

	chDone := make(chan bool)
	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(clientConn, reader)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(serverConn, clientConn)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()

	<-chDone
	return nil
}

func handleClient(client, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()

	<-chDone
}
