package httpserver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractSubdomain(t *testing.T) {
	{
		h := HttpInfo{host: "my-sub.ssh1.floxy.io"}
		assert.Equal(t, "my-sub", h.Subdomain())
	}
	{
		h := HttpInfo{host: "localhost"}
		assert.Equal(t, "localhost", h.Subdomain())
	}
}
