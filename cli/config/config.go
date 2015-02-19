package config

import (
	"github.com/cloudfoundry-incubator/lattice/cli/config/persister"
)

type Data struct {
	Target   string
	Username string
	Password string
}

type Config struct {
	persister persister.Persister
	data      *Data
}

func New(persister persister.Persister) *Config {
	config := &Config{persister: persister, data: &Data{}}
	return config
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
