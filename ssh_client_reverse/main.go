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

var key = `
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAsdbySbIgEJpk+bJnDQFU7Mpnpkwgru/TH2Q3xckhHYHFiAxtsw/W
nZOsV23Llxfka5fdja2OGMPkjV0H6S5MpjY0ZRXDPTST6BwtH93GFPinvYqVZr3n0TZEn4
26reCwJBrcwV6D42NLeQXNKnJI2Yx5OMWtyBjeFl/tjlsPzCgNZmEVzQ6UKBWuwJfkjUgR
7QybcCNhAmroruUOyrTDW5J087FMH8XD84tMJC7Uw/9/18sIMsxGRbphoKltZGlD34JSTQ
mK4heQ2xFUaTtz2zNHen0FGDHed5YPos+VQHIoHg7eJ5kx+0bDQZfHxsP8zWQN6UnDvVcR
fiJ9sXDbcQAAA9AhQnbRIUJ20QAAAAdzc2gtcnNhAAABAQCx1vJJsiAQmmT5smcNAVTsym
emTCCu79MfZDfFySEdgcWIDG2zD9adk6xXbcuXF+Rrl92NrY4Yw+SNXQfpLkymNjRlFcM9
NJPoHC0f3cYU+Ke9ipVmvefRNkSfjbqt4LAkGtzBXoPjY0t5Bc0qckjZjHk4xa3IGN4WX+
2OWw/MKA1mYRXNDpQoFa7Al+SNSBHtDJtwI2ECauiu5Q7KtMNbknTzsUwfxcPzi0wkLtTD
/3/XywgyzEZFumGgqW1kaUPfglJNCYriF5DbEVRpO3PbM0d6fQUYMd53lg+iz5VAcigeDt
4nmTH7RsNBl8fGw/zNZA3pScO9VxF+In2xcNtxAAAAAwEAAQAAAQEArksBbqS6tEr9B5OH
V8GkX+YHK36U0Z6OBcgMuTVj1S1oUOwNX174cbtXPuCGlfB+l8xhAQfFqhPjHYC9zhToXk
Xll+R6UrQC+YsT1pVeGxOQIj1+KxGX7v0GyHD5MoxxVRFWpdVh4Sthvpym9SDIsR3xeGiU
5vUoRDmD8u7gTq6V4tH4qAVkrF17dAjxOp94iuyDw5QubdzmcDh6JDaG5fHThmqf5M7BFg
lul4BKjvjIDWk3x2R/p9CwPcvyFhlFPGjFOoW8cxFjl7t7lcFE81e31CSnbII9UIFHVkWC
ZN1Wil6OZJbLXF8LRBib+S0Ux0xWizvX5zLb3QB5fTdI2QAAAIAdoBnE1DBhJgBZ6v1KP+
OH7PYaFvdksjC6tJ9Qfkq1CpxkIlVPys3kUGY4dyehPNqWbG3WLTb4eGTdfV/Hb9UbRVym
dUrlFi+kcJwxIfj7ygF1E71DhN6TuA/mbpAAkLP/mUi2A49aOG5b7k+C9AENDSHVMMdF9r
oXf9Eh1SPoVgAAAIEA7MPruZdz11ArzHMuYooYcKOFGdfMTi2KdolYu2S/pzZTtNpS8YOn
z1rik8qvgEmUpFnxWLIJpMAfufy4EWZ9IabfHqDh6H0M70ZY/IutF9MK9ykb70sgBegH6N
+hpm5aEqHDSZwGSsGDQjgXr04RnjZvbxKvStgr0h1dud9gSA8AAACBAMBJh9T3qrWhW2ee
AkgiWevfpEGTpzeu1Qea5hTul31tDl/WCLFK5vxFA3q8PCnDPQSdc6GtnzfQAeI2Q6PekM
78U4OAG/fofWJW+D6LIwnSJBOHZjBE77B8XamM2e2ZfZXlsCba1hP44BEdYaUOG4swGMNh
dHHBtZYPHd/txSR/AAAAF2RhbmllbC5rYW5jenVrQHBpc21vLmlvAQID
-----END OPENSSH PRIVATE KEY-----
`

func publicKey() ssh.AuthMethod {
	key, err := ssh.ParsePrivateKey([]byte(key))
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
			// put here your private key path
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

func main() {
	host := flag.String("host", "", "a host")
	kind := flag.String("K", "", "a kind")
	flag.Parse()

	config := SshConfiguration{
		User: "app",
		Host: "localhost",
		Port: 2222,
	}

	if *kind == "L"{
		NewSshConnection(config).
			executeGetPort().
			execLocal(*host)
	}else{
		NewSshConnection(config).
			executeAllocatePort().
			execRemote(*host)
	}

}
