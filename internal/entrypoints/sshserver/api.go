package sshserver

import (
	"context"
	"fmt"
	"github.com/danielsussa/floxy/internal/pkg/env"
	"github.com/danielsussa/floxy/internal/pkg/keys"
	"github.com/danielsussa/floxy/internal/pkg/store"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"log"
)

var (
	server ssh.Server
)

// curl localhost:8081
// ssh -N -L 8081:localhost:8090 user:daa@localhost -p 2222
// ssh -N -R 8090:localhost:8080 user:daa@localhost -p 2222

func Shutdown(ctx context.Context) error {
	return server.Shutdown(ctx)
}
func Start(e *store.Engine) {
	port := env.GetOrDefault(env.ServerSshPort, "2222")
	log.Println(fmt.Sprintf("starting ssh server on port %s...", port))

	forwardHandler := &forwardedTCPHandler{
		store: e,
	}

	go func() {
		server = ssh.Server{
			Addr: fmt.Sprintf(":%s", port),
			RequestHandlers: map[string]ssh.RequestHandler{
				"tcpip-forward":        forwardHandler.HandleSSHRequest,
				"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
			},
			ChannelHandlers: map[string]ssh.ChannelHandler{
				"direct-tcpip": func(srv *ssh.Server, conn *ssh2.ServerConn, newChan ssh2.NewChannel, ctx ssh.Context) {
					forwardHandler.directTCPIPHandler(srv, conn, newChan, ctx)
				},
			},
			ReversePortForwardingCallback: func(ctx ssh.Context, host string, port uint32) bool {
				if port != 0 {
					return false
				}

				ctx.SetValue("user", ctx.User())
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
