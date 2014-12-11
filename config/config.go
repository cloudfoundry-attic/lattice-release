package config

import (
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
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

func (c *Config) SetTarget(target string) error {
	c.data.Target = target
	return c.save()
}

func (c *Config) SetLogin(username string, password string) error {
	c.data.Username = username
	c.data.Password = password
	return c.save()
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

func (c *Config) Target() string {
	return c.data.Target
}

func (c *Config) save() error {
	return c.persister.Save(c.data)
}
