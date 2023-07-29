package httpserver

import (
	"fmt"
	"github.com/danielsussa/floxy/internal/env"
	"github.com/danielsussa/floxy/internal/sshserver"
	"io"
	"log"
	"net"
	"strings"
)

var listener net.Listener

func Shutdown() error {
	return listener.Close()
}

func Start() chan bool {
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
			go handleCall(serverConn, baseDns)
		}
	}()
	return chStart
}

func handleCall(serverConn net.Conn, baseDns string) error {
	defer serverConn.Close()

	info, reader, bRemain := extractHttpInfo(serverConn)

	dns := strings.ReplaceAll(strings.Split(info.Host(), baseDns)[0], ".", "")

	user, err := sshserver.GetUserByDomain(dns)
	if err != nil {
		return err
	}

	clientConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", user.Port))
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

//func handleCall(serverConn net.Conn, baseDns string) error {
//	defer serverConn.Close()
//
//	info := extractHttpInfo(serverConn)
//
//	dns := strings.ReplaceAll(strings.Split(info.Host(), baseDns)[0], ".", "")
//
//	user, err := sshserver.GetUserByDomain(dns)
//	if err != nil {
//		return err
//	}
//
//	clientConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", user.Port))
//	if err != nil {
//		return err
//	}
//
//	_, err = fmt.Fprint(clientConn, info.Buffer())
//	if err != nil {
//		return err
//	}
//
//	handleClient(clientConn, serverConn)
//	return nil
//}

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
