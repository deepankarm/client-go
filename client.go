package client

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jina-ai/client-go/jina"
	"github.com/viki-org/dnscache"
)

var resolver = dnscache.New(time.Minute)
var httpClient *http.Client

type CallbackType func(*jina.DataRequestProto)

type Client interface {
	POST(requests <-chan *jina.DataRequestProto, onDone, onError, onAlways CallbackType) error
}

type HealthCheckClient interface {
	HealthCheck() (bool, error)
}

// Clint & resolver taken from https://github.com/juicedata/juicefs/blob/main/pkg/object/restful.go
func updateHTTPClient() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			TLSHandshakeTimeout:   time.Second * 20,
			ResponseHeaderTimeout: time.Second * 30,
			IdleConnTimeout:       time.Second * 300,
			MaxIdleConnsPerHost:   500,
			Dial: func(network string, address string) (net.Conn, error) {
				separator := strings.LastIndex(address, ":")
				host := address[:separator]
				port := address[separator:]
				ips, err := resolver.Fetch(host)
				if err != nil {
					return nil, err
				}
				if len(ips) == 0 {
					return nil, fmt.Errorf("no such host: %s", host)
				}
				var conn net.Conn
				n := len(ips)
				first := rand.Intn(n)
				dialer := &net.Dialer{Timeout: time.Second * 10}
				for i := 0; i < n; i++ {
					ip := ips[(first+i)%n]
					address = ip.String()
					if port != "" {
						address = net.JoinHostPort(address, port[1:])
					}
					conn, err = dialer.Dial(network, address)
					if err == nil {
						return conn, nil
					}
				}
				return nil, err
			},
			DisableCompression: true,
		},
		Timeout: time.Hour,
	}
}

func init() {
	updateHTTPClient()
}
