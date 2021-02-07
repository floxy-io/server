/*
Based on implementation of https://gist.github.com/codref/473351a24a3ef90162cf10857fac0ff3
Go-Language implementation of an SSH Reverse Tunnel, the equivalent of below SSH command:
   ssh -R 8085:127.0.0.1:1325 -R 8086:127.0.0.1:1326 app@ip
Copyright 2020, Daniel Kanczuk
MIT License, http://www.opensource.org/licenses/mit-license.php
*/

package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var FingerPrint string
var PrivateKey string
var SshHost string
var RemotePassword string
var Kind string

func main() {
	proxyHost := flag.String("h", "", "a proxyHost")
	flagKind := flag.String("k", "", "kind of proxy")
	flagPassword := flag.String("p", "", "remote password protected")
	//externalDomain := flag.String("e", "", "external domain")

	flag.Parse()

	if FingerPrint == "" {
		FingerPrint = os.Getenv("FLOXY_FINGERPRINT")
	}
	if PrivateKey == "" {
		PrivateKey = os.Getenv("FLOXY_KEY")
	}
	if SshHost == "" {
		SshHost = os.Getenv("FLOXY_SSH_HOST")
	}
	log.Println(fmt.Sprintf("init Floxy binary fingerprint firsts digits: %s", strings.Split(FingerPrint, "-")[0]))

	if Kind == "" {
		if flagKind == nil || *flagKind == "" {
			log.Fatal("Must use flag -k to specify local or remote")
		}else {
			Kind = *flagKind
		}
	}

	finalProxyHost := generateProxyHost(proxyHost)

	var err error

	switch Kind {
	case "local":
		for {
			err = startLocalProxy(localProxyConfig{
				PrivateKey:  PrivateKey,
				Fingerprint: FingerPrint,
				SshHost:     SshHost,
				ProxyHost:   finalProxyHost,
			})
			log.Println("error on local connection: ", err)
			time.Sleep(1 * time.Second)
		}
	case "remote":
		if RemotePassword != "" && *flagPassword != RemotePassword {
			log.Fatal("password protected remote, please use flag -p with a correct password")
		}
		for {
			err = startRemoteProxy(remoteProxyConfig{
				PrivateKey:  PrivateKey,
				Fingerprint: FingerPrint,
				SshHost:     SshHost,
				ProxyHost:   finalProxyHost,
			})
			log.Println("error on remote connection: ", err)
			time.Sleep(1 * time.Second)
		}
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
	ProxyHost   string
}

func generateProxyHost(input *string)string{
	finalInput := ""
	if input == nil || *input == "" {
		fmt.Print("-> specify host:port or port (if localhost): ")
		reader := bufio.NewReader(os.Stdin)
		readPort, _ := reader.ReadString('\n')
		finalInput = readPort
		finalInput = strings.ReplaceAll(finalInput, "\n", "")
		finalInput = strings.ReplaceAll(finalInput, "\r", "")
	}else{
		finalInput = *input
	}
	finalInputSpl := strings.Split(finalInput, ":")
	switch len(finalInputSpl) {
	case 1:
		if _, err := strconv.Atoi(finalInputSpl[0]); err != nil {
			log.Fatal("You must specify host:port or at least a correct port number")
		}
		finalInput = fmt.Sprintf("localhost:%s",finalInputSpl[0])
	case 2:
		break
	default:
		log.Fatal("You must specify host:port or at least a correct port number")
	}
	return finalInput
}

func startRemoteProxy(config remoteProxyConfig) error {

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
	if !ok {
		return fmt.Errorf("server respond: %s", string(b))
	}
	serverPort := string(b)

	// listening to server reverse proxy
	l, err := serverClient.Listen("tcp", fmt.Sprintf("localhost:%s", serverPort))
	if err != nil {
		return fmt.Errorf("server error: %s", err)
	}

	log.Println("FloxyR success connect to server!")

	for {
		serverConn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("server cannot accept conn: %s", err)
		}

		go handleRemoteConn(serverConn, config.ProxyHost)
	}
}

func handleRemoteConn(serverConn net.Conn, host string) {
	proxyConn, err := net.Dial("tcp", host)
	if err != nil {
		return
	}
	waitUntilEnd := make(chan bool)
	go func() {
		_, _ = io.Copy(serverConn, proxyConn)
		serverConn.Close()
		waitUntilEnd <- true

	}()

	_, _ = io.Copy(proxyConn, serverConn)
	proxyConn.Close()
	<-waitUntilEnd
}

type localProxyConfig struct {
	PrivateKey  string
	Fingerprint string
	SshHost     string
	ProxyHost   string
}

func startLocalProxy(config localProxyConfig) error {
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
	hostListener, err := net.Listen("tcp", config.ProxyHost)
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("(%s) FloxyL success connect to server!", time.Now()))

	for {
		listenerConn, err := hostListener.Accept()
		if err != nil {
			return err
		}

		go handleLocal(hostListener, listenerConn, serverClient, serverPort)
	}
}

func handleLocal(hostListener net.Listener, listenerConn net.Conn, serverClient *ssh.Client, port string){
	serverConn, err := serverClient.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		log.Println(fmt.Sprintf("(%s) cannot call proxy server!", time.Now()))
		hostListener.Close()
		return
	}

	waitUntilEnd := make(chan bool)

	go func() {
		_, _ = io.Copy(listenerConn, serverConn)
		listenerConn.Close()
		waitUntilEnd <- true
	}()

	_, _ = io.Copy(serverConn, listenerConn)
	serverConn.Close()
	<-waitUntilEnd
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
