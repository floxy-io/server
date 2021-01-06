/*
Based on implementation of https://gist.github.com/codref/473351a24a3ef90162cf10857fac0ff3
Go-Language implementation of an SSH Reverse Tunnel, the equivalent of below SSH command:
   ssh -R 8085:127.0.0.1:1325 -R 8086:127.0.0.1:1326 app@146.148.21.125
Copyright 2017, Daniel Kanczuk
MIT License, http://www.opensource.org/licenses/mit-license.php
*/

package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"strconv"
)

type SshConfiguration struct {
	Host string
	Port int
	User string
}

func (config *SshConfiguration) String() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func publicKey() ssh.AuthMethod {
	key, err := ssh.ParsePrivateKey([]byte(PrivateKey))
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

type localRemoteListener struct {
	local string
	remote net.Listener
}

func (lrl localRemoteListener) listen(){
	defer lrl.remote.Close()
	for {
		localConn, err := net.Dial("tcp", lrl.local)
		if err != nil {
			log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
		}

		remoteConn, err := lrl.remote.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		lrl.handleRemoteClient(localConn, remoteConn)
	}
}

func(lrl localRemoteListener) handleRemoteClient(remoteConn net.Conn, localConn net.Conn) {
	defer remoteConn.Close()
	chDone := make(chan bool, 2)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(remoteConn, localConn)
		log.Println("remote->local")
		if err != nil {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(localConn, remoteConn)
		log.Println("local->remote")
		if err != nil {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()
	<-chDone
	log.Println("finish")
}

type sshConnection struct {
	client     *ssh.Client
	remotePort int
}

func NewSshConnection(config SshConfiguration)*sshConnection{
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			// put here your private PrivateKey path
			publicKey(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Connect to SSH remote server using serverEndpoint
	serverConn, err := ssh.Dial("tcp", config.String(), sshConfig)
	if err != nil {
		log.Fatalln(fmt.Printf("Dial INTO remote server error: %s", err))
	}
	return &sshConnection{client: serverConn}
}

func (sc *sshConnection) executeAllocatePort() *sshConnection {
	_, b, err := sc.client.SendRequest("allocate-reverse-port", true, nil)
	if err != nil {
		log.Fatalln(fmt.Printf("Cannot setup port: %s", err))
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		log.Fatalln(fmt.Printf("Cannot convert to port: %s", err))
	}
	log.Println("allocate port", i)
	sc.remotePort = i
	return sc
}

func (sc *sshConnection) executeGetPort() *sshConnection {
	ok, b, err := sc.client.SendRequest("allocate-local-port", true, nil)
	if err != nil {
		log.Fatalln(fmt.Printf("Cannot setup port: %s", err))
	}
	if !ok {
		log.Fatalln(fmt.Printf("Error to get port: %s", string(b)))
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		log.Fatalln(fmt.Printf("Cannot convert to port: %s", err))
	}
	log.Println("get port", i)
	sc.remotePort = i
	return sc
}


func (sc *sshConnection) execRemote(host string) {
	remoteListener, err := sc.client.Listen("tcp", fmt.Sprintf("localhost:%d",sc.remotePort))
	if err != nil {
		log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
	}
	listener := &localRemoteListener{
		local:  host,
		remote: remoteListener,
	}
	defer listener.remote.Close()
	for {
		localConn, err := net.Dial("tcp", listener.local)
		if err != nil {
			log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
		}

		remoteConn, err := listener.remote.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		listener.handleRemoteClient(remoteConn, localConn)
	}
}

func (sc *sshConnection) execLocal(host string) {
	remoteListener, err := sc.client.Listen("tcp", host)
	if err != nil {
		log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
	}
	listener := &localRemoteListener{
		local:  fmt.Sprintf("localhost:%d",sc.remotePort),
		remote: remoteListener,
	}
	defer listener.remote.Close()
	for {
		localConn, err := net.Dial("tcp", listener.local)
		if err != nil {
			log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
		}

		remoteConn, err := listener.remote.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		listener.handleRemoteClient(localConn, remoteConn)
	}
}

var FingerPrint string
var PrivateKey string
var SshHost string
var SshPort int
var SshUser string

func main() {
	log.Println("init local on fingerprint:", FingerPrint)
	host := flag.String("host", "", "a host")
	flag.Parse()

	config := SshConfiguration{
		User: SshUser,
		Host: SshHost,
		Port: SshPort,
	}

	NewSshConnection(config).
		executeGetPort().
		execLocal(*host)
}
