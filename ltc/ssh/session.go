package ssh

//go:generate counterfeiter -o mocks/fake_session.go . Session
type Session interface {
	KeepAlive() (stopChan chan<- struct{})
	Resize(width, height int) error
	Shell() error
	Run(string) error
	Wait() error
	Close() error
}

type SSHAPISessionFactory struct{}

func (*SSHAPISessionFactory) New(client Client, width, height int, desirePTY bool) (Session, error) {
	return client.Open(width, height, desirePTY)
}
