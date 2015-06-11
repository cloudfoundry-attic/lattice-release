package config

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
)

type Data struct {
	Target     string         `json:"target"`
	Username   string         `json:"username,omitempty"`
	Password   string         `json:"password,omitempty"`
	BlobTarget BlobTargetInfo `json:"blob_target_info"`
}

type BlobTargetInfo struct {
	TargetHost string `json:"host,omitempty"`
	TargetPort uint16 `json:"port,omitempty"`
	AccessKey  string `json:"access_key,omitempty"`
	SecretKey  string `json:"secret_key,omitempty"`
	BucketName string `json:"bucket_name,omitempty"`
}

type Config struct {
	persister persister.Persister
	data      *Data
}

func New(persister persister.Persister) *Config {
	return &Config{persister: persister, data: &Data{}}
}

func (c *Config) SetTarget(target string) {
	c.data.Target = target
}

func (c *Config) SetLogin(username string, password string) {
	c.data.Username = username
	c.data.Password = password
}

func (c *Config) Target() string {
	return c.data.Target
}

func (c *Config) Username() string {
	return c.data.Username
}

func (c *Config) Loggregator() string {
	return "doppler." + c.data.Target
}

func (c *Config) Receptor() string {
	if c.data.Username == "" {
		return "http://receptor." + c.data.Target
	}

	return "http://" + c.data.Username + ":" + c.data.Password + "@receptor." + c.data.Target
}

func (c *Config) Load() error {
	return c.persister.Load(c.data)
}

func (c *Config) Save() error {
	return c.persister.Save(c.data)
}

func (c *Config) SetBlobTarget(host string, port uint16, accessKey, secretKey, bucketName string) {
	c.data.BlobTarget.TargetHost = host
	c.data.BlobTarget.TargetPort = port
	c.data.BlobTarget.AccessKey = accessKey
	c.data.BlobTarget.SecretKey = secretKey
	c.data.BlobTarget.BucketName = bucketName
}

func (c *Config) BlobTarget() BlobTargetInfo {
	return c.data.BlobTarget
}

func (bti BlobTargetInfo) Proxy() func(req *http.Request) (*url.URL, error) {
	if bti.TargetHost == "" {
		return func(*http.Request) (*url.URL, error) {
			return nil, fmt.Errorf("missing proxy host")
		}
	}
	if bti.TargetPort == 0 {
		return func(*http.Request) (*url.URL, error) {
			return nil, fmt.Errorf("missing proxy port")
		}
	}
	return func(req *http.Request) (*url.URL, error) {
		proxy := fmt.Sprintf("http://%s:%d", bti.TargetHost, bti.TargetPort)
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy address %q: %v", proxy, err)
		}
		return proxyURL, nil
	}
}
