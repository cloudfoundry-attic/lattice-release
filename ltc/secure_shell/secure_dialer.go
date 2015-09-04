package secure_shell

import "golang.org/x/crypto/ssh"

type sshDialer struct{}

func NewSecureDialer() *sshDialer {
	return &sshDialer{}
}

type secureClient struct {
	client *ssh.Client
}

func (t *sshDialer) Dial(network, addr string, config *ssh.ClientConfig) (SecureClient, error) {
	client, err := ssh.Dial(network, addr, config)
	if err != nil {
		return nil, err
	}
	return &secureClient{client}, err
}

func (c *secureClient) NewSession() (SecureSession, error) {
	return c.client.NewSession()
}
