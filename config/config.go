package config

import (
	"github.com/pivotal-cf-experimental/lattice-cli/config/persister"
)

type Data struct {
	Target string
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

func (c *Config) Loggregator() string {
	return "doppler." + c.data.Target
}

func (c *Config) Receptor() string {
	return "http://receptor." + c.data.Target
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
