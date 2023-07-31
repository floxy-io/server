package store

import (
	"github.com/gliderlabs/ssh"
	"net"
	"sync"
)

type Engine struct {
	dnsToPort sync.Map
}

type Register struct {
	Ln   net.Listener
	Port int64
}

func (e *Engine) Get(dns string) (Register, bool) {
	val, ok := e.dnsToPort.Load(dns)
	if !ok {
		return Register{}, false
	}

	return val.(Register), true
}

func (e *Engine) Add(ctx ssh.Context, r Register) {
	e.dnsToPort.Store(ctx.User(), r)
}

func (e *Engine) Remove(ctx ssh.Context) {
	val, ok := e.dnsToPort.Load(ctx.User())
	if !ok {
		return
	}
	_ = val.(Register).Ln.Close()
	e.dnsToPort.Delete(ctx.User())
}

func New() *Engine {
	return &Engine{
		dnsToPort: sync.Map{},
	}
}
