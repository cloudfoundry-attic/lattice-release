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
	return &http.Client{
		Timeout: config.Timeout,
	}
}

func NewStreamingClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				KeepAlive: 10 * time.Second,
			}).Dial,
		},
	}
}
