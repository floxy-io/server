/*
Based on implementation of https://gist.github.com/codref/473351a24a3ef90162cf10857fac0ff3
Go-Language implementation of an SSH Reverse Tunnel, the equivalent of below SSH command:
   ssh -R 8085:127.0.0.1:1325 -R 8086:127.0.0.1:1326 app@146.148.21.125
Copyright 2017, Daniel Kanczuk
MIT License, http://www.opensource.org/licenses/mit-license.php
*/

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var FingerPrint string
var PrivateKey string
var SshHost string
var RemotePassword string

func main() {
	proxyHost := flag.String("h", "", "a proxyHost")
	flagKind := flag.String("k", "", "kind of proxy")
	flagPassword := flag.String("p", "", "remote password protected")


	flag.Parse()

	if RemotePassword != "" && *flagPassword != RemotePassword{
		log.Fatal("password protected remote, please use flag -p to pass a password")
	}

	if FingerPrint == "" {
		FingerPrint = os.Getenv("FLOXY_FINGERPRINT")
	}
	if PrivateKey == "" {
		PrivateKey = os.Getenv("FLOXY_KEY")
	}
	if SshHost == "" {
		SshHost = os.Getenv("FLOXY_SSH_HOST")
	}
	log.Println(fmt.Sprintf("init %s on fingerprint !", FingerPrint))

	if flagKind == nil || *flagKind == "" {
		log.Fatal("Must use flag -kind to specify local or remote")
	}

	if proxyHost == nil || *proxyHost == "" {
		log.Fatal("Must use flag -host host:port")
	}
	if len(strings.Split(*proxyHost, ":")) != 2 {
		log.Fatal("You must specify host:port")
	}

	var err error

	switch *flagKind {
	case "local":
		err = startLocalProxy(localProxyConfig{
			PrivateKey:  PrivateKey,
			Fingerprint: FingerPrint,
			SshHost:     SshHost,
			ProxyHost:   proxyHost,
		})
	case "remote":
		err = startRemoteProxy(remoteProxyConfig{
			PrivateKey:  PrivateKey,
			Fingerprint: FingerPrint,
			SshHost:     SshHost,
			ProxyHost:   proxyHost,
		})
	default:
		log.Fatal("cannot find kind")
	}

	if err != nil {
		log.Fatal(err)
	}
}

type remoteProxyConfig struct {
	PrivateKey  string
	Fingerprint string
	SshHost     string
	ProxyHost   *string
}

func startRemoteProxy(config remoteProxyConfig) error{

	// ssh config
	sshConfig := &ssh.ClientConfig{
		User: config.Fingerprint,
		Auth: []ssh.AuthMethod{
			// put here your private PrivateKey path
			publicKey(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Connect to SSH remote server using serverEndpoint
	serverClient, err := ssh.Dial("tcp", config.SshHost, sshConfig)
	if err != nil {
		return fmt.Errorf("dial INTO remote server error: %s", err)
	}

	ok, b, err := serverClient.SendRequest("allocate-reverse-port", true, nil)
	if err != nil {
		return fmt.Errorf("cannot setup port: %s", err)
	}
	if ! ok {
		return fmt.Errorf("server respond: %s", string(b))
	}
	serverPort := string(b)

	// listening to server reverse proxy
	l, err := serverClient.Listen("tcp", fmt.Sprintf("localhost:%s", serverPort))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("FloxyR success connect to server!")

	for {
		serverConn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		proxyConn, err := net.Dial("tcp", *config.ProxyHost)
		if err != nil {
			continue
		}
		handleConn(serverConn, proxyConn)
	}
}

func handleConn(serverConn , proxyConn net.Conn) {
	waitUntilEnd := make(chan bool)
	go func() {
		_, _ = io.Copy(serverConn, proxyConn)
		serverConn.Close()
		waitUntilEnd <- true

	}()

	_, _ = io.Copy(proxyConn, serverConn)
	proxyConn.Close()
	<- waitUntilEnd
}

type localProxyConfig struct {
	PrivateKey  string
	Fingerprint string
	SshHost     string
	ProxyHost   *string
}

func startLocalProxy(config localProxyConfig) error{
	// ssh config
	sshConfig := &ssh.ClientConfig{
		User: config.Fingerprint,
		Auth: []ssh.AuthMethod{
			// put here your private PrivateKey path
			publicKey(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Connect to SSH remote server using serverEndpoint
	serverClient, err := ssh.Dial("tcp", config.SshHost, sshConfig)
	if err != nil {
		return fmt.Errorf("(%s) dial INTO remote server error: %s", time.Now(), err)
	}

	// allocate port
	ok, b, err := serverClient.SendRequest("allocate-local-port", true, nil)
	if err != nil {
		return fmt.Errorf("(%s) cannot setup port: %s", time.Now(), err)
	}
	if !ok {
		return fmt.Errorf("server respond: %s", string(b))
	}
	serverPort := string(b)

	// start listen to proxy
	hostListener, err := net.Listen("tcp", *config.ProxyHost)
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("(%s) FloxyL success connect to server!", time.Now()))

	for {
		listenerConn, err := hostListener.Accept()
		if err != nil {
			return err
		}

		serverConn, err := serverClient.Dial("tcp", fmt.Sprintf("localhost:%s", serverPort))
		if err != nil {
			log.Println(fmt.Sprintf("(%s) cannot call proxy server!", time.Now()))
			continue
		}

		waitUntilEnd := make(chan bool)

		go func() {
			_, err = io.Copy(listenerConn, serverConn)
			listenerConn.Close()
			waitUntilEnd <- true
		}()

		_, err = io.Copy(serverConn, listenerConn)
		serverConn.Close()
		<- waitUntilEnd
	}
}

func publicKey() ssh.AuthMethod {
	dst, err := base64.StdEncoding.DecodeString(PrivateKey)
	if err != nil {
		panic(err)
	}

	key, err := ssh.ParsePrivateKey(dst)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(key)
}