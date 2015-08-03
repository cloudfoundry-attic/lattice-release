package config

import (
	"github.com/cloudfoundry-incubator/lattice/ltc/config/dav_blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
)

type Data struct {
	Target    string                `json:"target"`
	Username  string                `json:"username,omitempty"`
	Password  string                `json:"password,omitempty"`
	BlobStore dav_blob_store.Config `json:"dav_blob_store"`
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

func (c *Config) SetLogin(username, password string) {
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

func (c *Config) SetBlobStore(host, port, username, password string) {
	c.data.BlobStore.Host = host
	c.data.BlobStore.Port = port
	c.data.BlobStore.Username = username
	c.data.BlobStore.Password = password
}

func (c *Config) BlobStore() dav_blob_store.Config {
	return c.data.BlobStore
}
