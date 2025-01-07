package scanner

import (
	"crypto/tls"
	"sync"

	"github.com/dsecuredcom/vhost-fuzzer/pkg/config"
	"github.com/valyala/fasthttp"
)

type ClientPool struct {
	sync.RWMutex
	clients map[string]*fasthttp.HostClient
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		clients: make(map[string]*fasthttp.HostClient),
	}
}

func (cp *ClientPool) getClient(ip string, cfg *config.Config) *fasthttp.HostClient {
	cp.RLock()
	client, exists := cp.clients[ip]
	cp.RUnlock()

	if exists {
		return client
	}

	cp.Lock()
	defer cp.Unlock()

	// Double check after acquiring write lock
	if client, exists = cp.clients[ip]; exists {
		return client
	}

	var port string
	var isTLS bool
	if cfg.Protocol == "https" {
		port = "443"
		isTLS = true
	} else {
		port = "80"
		isTLS = false
	}

	client = &fasthttp.HostClient{
		Addr:                          ip + ":" + port,
		IsTLS:                         isTLS,
		MaxConnDuration:               cfg.MaxConnDuration,
		MaxIdleConnDuration:           cfg.MaxIdleConnDuration,
		ReadTimeout:                   cfg.ReadTimeout,
		WriteTimeout:                  cfg.WriteTimeout,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		NoDefaultUserAgentHeader:      true,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	cp.clients[ip] = client
	return client
}
