package scanner

import (
	"crypto/tls"
	"sync"

	"github.com/dsecuredcom/vhost-fuzzer/pkg/config"
	"github.com/valyala/fasthttp"
)

type clientCache struct {
	clients map[string]*fasthttp.HostClient
	mu      sync.Mutex
}

func newClientCache() *clientCache {
	return &clientCache{
		clients: make(map[string]*fasthttp.HostClient),
	}
}

func (cc *clientCache) getClient(ip string, cfg config.Config) *fasthttp.HostClient {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	key := ip + "|" + cfg.Protocol
	if hc, ok := cc.clients[key]; ok {
		return hc
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
	hc := &fasthttp.HostClient{
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
	cc.clients[ip] = hc
	return hc
}
