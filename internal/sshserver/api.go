package sshserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/repo"
	"github.com/danielsussa/freeport"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"log"
	"sync"
)


var (
	server ssh.Server
)

func Shutdown(ctx context.Context)error {
	return server.Shutdown(ctx)
}
func Start() {
	log.Println("starting ssh server on port 2222...")

	forwardHandler := &forwardedTCPHandler{}

	go func() {
		server = ssh.Server{
			LocalPortForwardingCallback: func(ctx ssh.Context, host string, port uint32) bool {
				proxy, err := getProxy(ctx)
				if err != nil {
					return false
				}
				if proxy.Port != int(port){
					return false
				}

				log.Println("Accepted forward", host, port)
				return true
			},
			Addr: ":2222",
			RequestHandlers: map[string]ssh.RequestHandler{
				"tcpip-forward":        forwardHandler.HandleSSHRequest,
				"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
				"allocate-reverse-port": func(ctx ssh.Context, srv *ssh.Server, req *ssh2.Request) (ok bool, payload []byte) {
					log.Println("start allocate-reverse-port")
					proxy, err := getProxy(ctx)
					if err != nil {
						log.Println("error get proxy: ", err)
						return false, []byte("")
					}
					if proxy.Port == 0 || !freeport.CheckPortIsFree(proxy.Port) {
						freePort, err := freeport.GetFreePort()
						if err != nil {
							log.Println("error allocate port: ", err)
							return false, []byte("error allocate port")
						}
						log.Println("new port allocated", freePort)
						err = repo.SetPort(proxy.Fingerprint, freePort)
						if err != nil {
							log.Println("error to set port: ", err)
							return false, []byte("error to set port")
						}
					}
					log.Println("allocating reverse port", proxy.Port)
					return true, []byte(fmt.Sprintf("%d", proxy.Port))
				},
				"allocate-local-port": func(ctx ssh.Context, srv *ssh.Server, req *ssh2.Request) (ok bool, payload []byte) {
					log.Println("start allocate-local-port")
					proxy, err := getProxy(ctx)
					if err != nil {
						log.Println("error to allocate port: ", err)
						return false, []byte("error to set port")
					}
					if proxy.Port == 0 {
						return false, []byte("no ports are available\n")
					}
					log.Println("allocating local port", proxy.Port)
					return true, []byte(fmt.Sprintf("%d", proxy.Port))
				},
			},
			ChannelHandlers: map[string]ssh.ChannelHandler{
				"direct-tcpip": func(srv *ssh.Server, conn *ssh2.ServerConn, newChan ssh2.NewChannel, ctx ssh.Context) {
					directTCPIPHandler(srv, conn, newChan, ctx)
				},
			},
			ReversePortForwardingCallback: func(ctx ssh.Context, host string, port uint32) bool {
				proxy, err := getProxy(ctx)
				if err != nil {
					return false
				}
				if proxy.Port != int(port){
					return false
				}
				if !proxy.Activated {
					_ = repo.ActiveProxy(proxy.Fingerprint)
				}
				log.Println("attempt to bind reverse", host, port, "granted")
				return true
			},
			PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
				log.Println("start PublicKeyHandler")
				val, err := repo.GetByUserAndKey(ctx.User(), key.Marshal())
				if err != nil {
					log.Println("cannot find user: ", err)
					return false
				}
				ctx.SetValue("keyProxy", &val)
				return true
			},
		}
		log.Fatal(server.ListenAndServe())
	}()
}

type HostResponse struct {
	PrivateKey string
	Port       int
	PublicKey  []byte
}

var mutex sync.Mutex

func AllocateNewHost() (HostResponse, error) {
	mutex.Lock()
	defer mutex.Unlock()

	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return HostResponse{}, err
	}

	publicKey, err := ssh2.NewPublicKey(&privatekey.PublicKey)
	if err != nil {
		return HostResponse{}, err
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return HostResponse{}, err
	}

	return HostResponse{
		PrivateKey: toBase64PrivateKey(privatekey),
		Port:       port,
		PublicKey:  publicKey.Marshal(),
	}, nil
}

func toBase64PrivateKey(key *rsa.PrivateKey)string{
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	return base64.StdEncoding.EncodeToString(keyPem)
}


func getProxy(ctx ssh.Context)(*repo.Floxy, error){
	proxy := ctx.Value("keyProxy")
	if proxy == nil {
		return nil, fmt.Errorf("cannot find proxy")
	}
	return proxy.(*repo.Floxy), nil
}
