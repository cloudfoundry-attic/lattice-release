package cf_http

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
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

func NewTLSConfig(certFile, keyFile, caCertFile string) (*tls.Config, error) {
	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: false,
	}

	certBytes, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}

	if caCertFile != "" {
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
			return nil, errors.New("Unable to load caCert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}
