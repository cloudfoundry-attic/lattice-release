package cf_http

import (
	"net"
	"net/http"
	"time"
)

var config Config

type Config struct {
	Timeout time.Duration
}

func Initialize(timeout time.Duration) {
	config.Timeout = timeout
}

func NewClient() *http.Client {
	return newClient(5*time.Second, 0*time.Second, config.Timeout)
}

func NewStreamingClient() *http.Client {
	return newClient(5*time.Second, 30*time.Second, 0*time.Second)
}

func newClient(dialTimeout, keepAliveTimeout, timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   dialTimeout,
				KeepAlive: keepAliveTimeout,
			}).Dial,
		},
		Timeout: timeout,
	}
}
