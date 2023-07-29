package userstore

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

func Add(user string, port int) {
	dnsToPort.Store(user, port)
}

func Get(dns string) (int, error) {
	val, ok := dnsToPort.Load(dns)
	if !ok {
		return 0, fmt.Errorf("not found")
	}

	return val.(int), nil
}

func Has(dns string) bool {
	_, ok := dnsToPort.Load(dns)
	return ok
}
