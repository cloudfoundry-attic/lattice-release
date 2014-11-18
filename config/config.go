package config

import (
	"github.com/pivotal-cf-experimental/diego-edge-cli/config/persister"
)

type Data struct {
	Api string
}

type Config struct {
	persister persister.Persister
	data      *Data
}

func New(persister persister.Persister) *Config {
	config := &Config{persister: persister, data: &Data{}}
	return config
}

func (c *Config) Api() string {
	return c.data.Api
}

func (c *Config) SetApi(api string) error {
	c.data.Api = api
	return c.save()
}

func (c *Config) Load() error {
	return c.persister.Load(c.data)
}

func (c *Config) save() error {
	return c.persister.Save(c.data)
}
