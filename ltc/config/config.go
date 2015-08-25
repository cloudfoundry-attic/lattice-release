package config

import (
	"github.com/cloudfoundry-incubator/lattice/ltc/config/persister"
)

type BlobStoreConfig struct {
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type S3BlobStoreConfig struct {
	Region     string `json:"region,omitempty"`
	AccessKey  string `json:"access_key,omitempty"`
	SecretKey  string `json:"secret_key,omitempty"`
	BucketName string `json:"bucket_name,omitempty"`
}

type Data struct {
	Target   string `json:"target"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	BlobStore   BlobStoreConfig   `json:"dav_blob_store,omitempty"`
	S3BlobStore S3BlobStoreConfig `json:"s3_blob_store,omitempty"`
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

func (c *Config) SetS3BlobStore(accessKey, secretKey, bucketName, region string) {
	c.data.S3BlobStore.AccessKey = accessKey
	c.data.S3BlobStore.BucketName = bucketName
	c.data.S3BlobStore.SecretKey = secretKey
	c.data.S3BlobStore.Region = region
}

func (c *Config) BlobStore() BlobStoreConfig {
	return c.data.BlobStore
}

func (c *Config) S3BlobStore() S3BlobStoreConfig {
	return c.data.S3BlobStore
}
