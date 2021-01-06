package main

import (
	"fmt"
	"github.com/danielsussa/freeport"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"io"
	"log"
)

type keyProxy struct {
	Key  string
	Port int
}

var pKeyMap map[string]*keyProxy


func init(){
	pKeyMap = make(map[string]*keyProxy)
	key := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCx1vJJsiAQmmT5smcNAVTsymemTCCu79MfZDfFySEdgcWIDG2zD9adk6xXbcuXF+Rrl92NrY4Yw+SNXQfpLkymNjRlFcM9NJPoHC0f3cYU+Ke9ipVmvefRNkSfjbqt4LAkGtzBXoPjY0t5Bc0qckjZjHk4xa3IGN4WX+2OWw/MKA1mYRXNDpQoFa7Al+SNSBHtDJtwI2ECauiu5Q7KtMNbknTzsUwfxcPzi0wkLtTD/3/XywgyzEZFumGgqW1kaUPfglJNCYriF5DbEVRpO3PbM0d6fQUYMd53lg+iz5VAcigeDt4nmTH7RsNBl8fGw/zNZA3pScO9VxF+In2xcNtx daniel.kanczuk@pismo.io"

	parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		panic(err)
	}

	pKeyMap[string(parsedKey.Marshal())] = &keyProxy{
		Key: key,
	}
}

func main() {
	log.Println("starting ssh server on port 2222...")

	forwardHandler := &ssh.ForwardedTCPHandler{}

	server := ssh.Server{
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			log.Println("Accepted forward", host, port)
			return true
		}),
		Addr: ":2222",
		SessionRequestCallback: func(sess ssh.Session, requestType string) bool {
			proxy,_ := pKeyMap[string(sess.PublicKey().Marshal())]

			for _, command := range sess.Command(){
				switch command {
				case "allocate-reverse-port":
					if proxy.Port == 0 || !freeport.CheckPortIsFree(proxy.Port) {
						freePort, err := freeport.GetFreePort()
						if err != nil {
							log.Fatal(err)
						}
						log.Println("new port allocated", freePort)
						proxy.Port = freePort
					}
					io.WriteString(sess, fmt.Sprintf("%d\n", proxy.Port))
				}
			}
			return false
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
			"allocate-reverse-port": func(ctx ssh.Context, srv *ssh.Server, req *ssh2.Request) (ok bool, payload []byte) {
				log.Println("start allocate-reverse-port")
				proxy := ctx.Value("keyProxy").(*keyProxy)
				if proxy.Port == 0 || !freeport.CheckPortIsFree(proxy.Port) {
					freePort, err := freeport.GetFreePort()
					if err != nil {
						log.Fatal(err)
					}
					log.Println("new port allocated", freePort)
					proxy.Port = freePort
				}
				log.Println("allocating reverse port", proxy.Port)
				return true, []byte(fmt.Sprintf("%d", proxy.Port))
			},
			"allocate-local-port": func(ctx ssh.Context, srv *ssh.Server, req *ssh2.Request) (ok bool, payload []byte) {
				log.Println("start allocate-local-port")
				proxy := ctx.Value("keyProxy").(*keyProxy)
				if proxy.Port == 0 {
					return false, []byte("no ports are available\n")
				}
				log.Println("allocating local port", proxy.Port)
				return true, []byte(fmt.Sprintf("%d", proxy.Port))
			},
		},
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			// TODO check if can handle this port
			log.Println("attempt to bind", host, port, "granted")
			return true
		}),
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			log.Println("start PublicKeyHandler")
			val,ok := pKeyMap[string(key.Marshal())]
			if !ok {
				return false
			}
			ctx.SetValue("keyProxy", val)
			return true
		},
	}

	log.Fatal(server.ListenAndServe())
}