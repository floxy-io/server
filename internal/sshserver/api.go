package sshserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/db"
	"github.com/danielsussa/freeport"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"log"
	"sync"
	"time"
)


type sshKeyProxy struct {
	PublicKey   []byte
	Fingerprint string
	Port        int
}

var (
	server ssh.Server
)

func Shutdown(ctx context.Context)error {
	return server.Shutdown(ctx)
}
func Start() {
	log.Println("starting ssh server on port 2222...")

	forwardHandler := &ssh.ForwardedTCPHandler{}

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
				_ = setUpdatedAt(proxy.Fingerprint)

				log.Println("Accepted forward", host, port)
				return true
			},
			Addr: ":2222",
			//SessionRequestCallback: func(sess ssh.Session, requestType string) bool {
			//	proxy,_ := pKeyMap[string(sess.PublicKey().Marshal())]
			//
			//	for _, command := range sess.Command(){
			//		switch command {
			//		case "allocate-reverse-port":
			//			if proxy.Port == 0 || !freeport.CheckPortIsFree(proxy.Port) {
			//				freePort, err := freeport.GetFreePort()
			//				if err != nil {
			//					log.Fatal(err)
			//				}
			//				log.Println("new port allocated", freePort)
			//				proxy.Port = freePort
			//			}
			//			io.WriteString(sess, fmt.Sprintf("%d\n", proxy.Port))
			//		}
			//	}
			//	return false
			//},
			RequestHandlers: map[string]ssh.RequestHandler{
				"tcpip-forward":        forwardHandler.HandleSSHRequest,
				"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
				"allocate-reverse-port": func(ctx ssh.Context, srv *ssh.Server, req *ssh2.Request) (ok bool, payload []byte) {
					log.Println("start allocate-reverse-port")
					proxy, err := getProxy(ctx)
					if err != nil {
						return false, []byte("")
					}
					if proxy.Port == 0 || !freeport.CheckPortIsFree(proxy.Port) {
						freePort, err := freeport.GetFreePort()
						if err != nil {
							log.Fatal(err)
						}
						log.Println("new port allocated", freePort)
						err = setPort(proxy.Fingerprint, freePort)
						if err != nil {
							log.Fatal(err)
						}
					}
					log.Println("allocating reverse port", proxy.Port)
					return true, []byte(fmt.Sprintf("%d", proxy.Port))
				},
				"allocate-local-port": func(ctx ssh.Context, srv *ssh.Server, req *ssh2.Request) (ok bool, payload []byte) {
					log.Println("start allocate-local-port")
					proxy := ctx.Value("keyProxy").(*sshKeyProxy)
					if proxy.Port == 0 {
						return false, []byte("no ports are available\n")
					}
					log.Println("allocating local port", proxy.Port)
					return true, []byte(fmt.Sprintf("%d", proxy.Port))
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
				_ = setUpdatedAt(proxy.Fingerprint)
				log.Println("attempt to bind", host, port, "granted")
				return true
			},
			PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
				log.Println("start PublicKeyHandler")
				val, err := getByUserAndKey(ctx.User(), key.Marshal())
				if err != nil {
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

	port, err := freeport.GetFreePort()
	if err != nil {
		return HostResponse{}, err
	}

	err = addKey(sshKeyProxy{PublicKey: publicKey.Marshal(), Fingerprint: fingerprint, Port: port})
	if err != nil {
		return HostResponse{}, err
	}

	return HostResponse{
		PrivateKey: toBase64PrivateKey(privatekey),
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

var insertOnTable = `
INSERT INTO sshPair (fingerprint,publicKey, port, createdAt)
VALUES(?,?,?,?);
`

func addKey(k sshKeyProxy)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare(insertOnTable)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(k.Fingerprint,base64.StdEncoding.EncodeToString(k.PublicKey), k.Port, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func getByUserAndKey(user string, public []byte)(sshKeyProxy, error){
	key := base64.StdEncoding.EncodeToString(public)

	dbConn := db.Get()
	row, err := dbConn.Query("SELECT fingerprint,publicKey,port FROM sshPair WHERE fingerprint=? AND publicKey=?", user, key)
	if err != nil {
		return sshKeyProxy{}, err
	}
	defer row.Close()

	for row.Next() {
		var fingerprint string
		var publicKey string
		var port int
		err = row.Scan(&fingerprint, &publicKey, &port)
		if err != nil {
			return sshKeyProxy{}, err
		}

		publicDec, err := base64.StdEncoding.DecodeString(publicKey)
		if err != nil {
			return sshKeyProxy{}, err
		}
		return sshKeyProxy{Fingerprint: fingerprint, PublicKey: publicDec, Port: port}, nil
	}
	return sshKeyProxy{}, fmt.Errorf("no scan")
}

func setPort(fingerprint string, port int)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare("UPDATE sshPair SET port=? WHERE fingerprint=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(port, fingerprint)
	if err != nil {
		return err
	}
	return nil
}

func setUpdatedAt(fingerprint string)error{
	dbConn := db.Get()
	stmt, err := dbConn.Prepare("UPDATE sshPair SET updatedAd=? WHERE fingerprint=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(time.Now(), fingerprint)
	if err != nil {
		return err
	}
	return nil
}

func getProxy(ctx ssh.Context)(*sshKeyProxy, error){
	proxy := ctx.Value("keyProxy")
	if proxy == nil {
		return nil, fmt.Errorf("cannot find proxy")
	}
	return proxy.(*sshKeyProxy), nil
}