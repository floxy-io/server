package sshserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/db"
	"github.com/danielsussa/freeport"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"io"
	"log"
	"sync"
)


type sshKeyProxy struct {
	PublicKey   string
	Fingerprint string
}


type keyProxy struct {
	FingerPrint string
	Key         string
	Port        int
}

var pKeyMap map[string]*keyProxy
var portMap map[int]*keyProxy

var (
	server ssh.Server
)

func Shutdown(ctx context.Context)error {
	return server.Shutdown(ctx)
}
func Start() {
	pKeyMap = make(map[string]*keyProxy)
	portMap = make(map[int]*keyProxy)
	log.Println("starting ssh server on port 2222...")

	forwardHandler := &ssh.ForwardedTCPHandler{}

	go func() {
		server = ssh.Server{
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
	}()
}

type HostResponse struct {
	PKey *rsa.PrivateKey
}

var mutex sync.Mutex

func AllocateNewHost(fingerprint string) (HostResponse, error) {
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

	pKeyStr := string(publicKey.Marshal())

	err = addKey(sshKeyProxy{PublicKey: pKeyStr, Fingerprint: fingerprint})
	if err != nil {
		return HostResponse{}, err
	}

	return HostResponse{
		PKey: privatekey,
	}, nil
}

var insertOnTable = `
INSERT INTO sshPair (fingerprint,publicKey)
VALUES(?,?);
`

func addKey(k sshKeyProxy)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(insertOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(k.Fingerprint, k.PublicKey)
	if err != nil {
		return err
	}
	return nil
}

func getByKey(public string)(sshKeyProxy, error){
	dbConn := db.Get()
	row, err := dbConn.Query("SELECT fingerprint,publicKey FROM sshPair")
	if err != nil {
		return sshKeyProxy{}, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey string
		err = row.Scan(&fingerprint, &publicKey)
		if err != nil {
			return sshKeyProxy{}, err
		}

		return sshKeyProxy{Fingerprint: fingerprint, PublicKey: public}, nil
	}
	return sshKeyProxy{}, fmt.Errorf("no scan")
}