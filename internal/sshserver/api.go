package sshserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/danielsussa/floxy/internal/env"
	"github.com/danielsussa/floxy/internal/infra/keys"
	"github.com/danielsussa/freeport"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	ssh2 "golang.org/x/crypto/ssh"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	server        ssh.Server
	sshUserMap    sync.Map
	idUserMap     sync.Map
	domainUserMap sync.Map
)

type ProxyUserMap struct {
	Id          string
	Port        int
	CreatedAt   time.Time
	ConnectedAt *time.Time
	Password    *string
	PublicKeys  *[]string
	SubDns      *string
}

func (p ProxyUserMap) Expired(t time.Time) bool {
	if p.ConnectedAt == nil {
		return t.Sub(p.CreatedAt).Minutes() > 10
	}
	return t.Sub(*p.ConnectedAt).Hours() > 5
}

type AddNewUserRequest struct {
	PublicKeys *[]string
	GenDomain  bool
}

//func GetUserByDomain(dns string) (*ProxyUserMap, error) {
//	return &ProxyUserMap{
//		Port: 1323,
//	}, nil
//}

func GetUserByDomain(dns string) (*ProxyUserMap, error) {
	val, ok := domainUserMap.Load(dns)
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return val.(*ProxyUserMap), nil
}

func GetUserById(id string) (*ProxyUserMap, error) {
	val, ok := idUserMap.Load(id)
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return val.(*ProxyUserMap), nil
}

func AddNewUser(req AddNewUserRequest) (*ProxyUserMap, error) {
	keysToRemove := make([]string, 0)
	mapPort := make(map[int]bool)
	sshUserMap.Range(func(key, value any) bool {
		pUser := value.(*ProxyUserMap)
		if pUser.Expired(time.Now()) {
			keysToRemove = append(keysToRemove, key.(string))
		}
		mapPort[pUser.Port] = true
		return false
	})
	for _, key := range keysToRemove {
		sshUserMap.Delete(key)
	}

	selectedPort := 0
	for i := 0; i <= 10; i++ {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			log.Println("error allocate port: ", err)
			return nil, fmt.Errorf("error to allocate port")
		}
		_, ok := mapPort[freePort]
		if ok {
			continue
		}
		selectedPort = freePort
		break
	}
	if selectedPort == 0 {
		return nil, fmt.Errorf("error to allocate port")
	}

	var pUser *ProxyUserMap

	if req.PublicKeys == nil {
		password := strings.Replace(uuid.New().String(), "-", "", -1)
		pUser = &ProxyUserMap{
			Port:      selectedPort,
			CreatedAt: time.Now(),
			Id:        uuid.New().String(),
			Password:  &password,
		}
		sshUserMap.Store(password, pUser)
	} else {
		pUser = &ProxyUserMap{
			Port:      selectedPort,
			Id:        uuid.New().String(),
			CreatedAt: time.Now(),
		}
		for _, pkey := range *req.PublicKeys {
			b, err := base64.StdEncoding.DecodeString(pkey)
			if err != nil {
				return nil, fmt.Errorf("pKey is not base64 result")
			}

			pk, _, _, _, err := ssh.ParseAuthorizedKey(b)
			if err != nil {
				return nil, fmt.Errorf("not valid public key")
			}

			sshUserMap.Store(base64.StdEncoding.EncodeToString(pk.Marshal()), pUser)
		}
	}

	if req.GenDomain {
		for {
			domain := strings.Split(uuid.New().String(), "-")[0]
			_, ok := domainUserMap.Load(domain)
			if ok {
				continue
			}
			domainUserMap.Store(domain, pUser)
			pUser.SubDns = &domain
			break
		}
	}

	idUserMap.Store(pUser.Id, pUser)

	return pUser, nil
}

// curl localhost:8081
// ssh -N -L 8081:localhost:8090 user:daa@localhost -p 2222
// ssh -N -R 8090:localhost:8080 user:daa@localhost -p 2222

func Shutdown(ctx context.Context) error {
	return server.Shutdown(ctx)
}
func Start() {
	port := env.GetOrDefault(env.ServerSshPort, "2222")
	log.Println(fmt.Sprintf("starting ssh server on port %s...", port))

	forwardHandler := &forwardedTCPHandler{}

	go func() {
		server = ssh.Server{
			LocalPortForwardingCallback: func(ctx ssh.Context, host string, port uint32) bool {
				val := ctx.Value("user")
				if val == nil {
					return false
				}
				pUser := val.(*ProxyUserMap)
				if pUser.Port != int(port) {
					return false
				}
				return true
			},
			Addr: fmt.Sprintf(":%s", port),
			RequestHandlers: map[string]ssh.RequestHandler{
				"tcpip-forward":        forwardHandler.HandleSSHRequest,
				"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
			},
			ChannelHandlers: map[string]ssh.ChannelHandler{
				"direct-tcpip": func(srv *ssh.Server, conn *ssh2.ServerConn, newChan ssh2.NewChannel, ctx ssh.Context) {
					directTCPIPHandler(srv, conn, newChan, ctx)
				},
			},
			ReversePortForwardingCallback: func(ctx ssh.Context, host string, port uint32) bool {
				if port != 0 {
					return false
				}

				ctx.SetValue("user", ctx.User())
				return true
			},
			PasswordHandler: func(ctx ssh.Context, password string) bool {
				//ctx.SetValue("password", password)
				return true
			},
		}
		s, err := ssh2.NewSignerFromKey(keys.LoadKey())
		if err != nil {
			panic(err)
		}
		server.AddHostKey(s)
		log.Fatal(server.ListenAndServe())
	}()
}

type HostResponse struct {
	PrivateKey string
	Port       int
	PublicKey  []byte
}
