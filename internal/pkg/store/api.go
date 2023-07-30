package store

import (
	"fmt"
	"sync"
)

type Engine struct {
	dnsToPort sync.Map
}

func New() Engine {
	return Engine{
		dnsToPort: sync.Map{},
	}
}

var (
	dnsToPort sync.Map
)

func Remove(user string) {
	dnsToPort.Delete(user)
}

func Add(user string, port int64) {
	dnsToPort.Store(user, port)
}

func Get(dns string) (int64, error) {
	val, ok := dnsToPort.Load(dns)
	if !ok {
		return 0, fmt.Errorf("not found")
	}

	return val.(int64), nil
}

func Has(dns string) bool {
	_, ok := dnsToPort.Load(dns)
	return ok
}
