package testrunner

import (
	"os/exec"
	"strconv"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	RegisterWithRouter bool
	DomainNames        string
	Address            string
	TaskHandlerAddress string
	EtcdCluster        string
	ConsulCluster      string
	Username           string
	Password           string
	NatsAddresses      string
	NatsUsername       string
	NatsPassword       string
	CORSEnabled        bool
	EtcdClientCert     string
	EtcdClientKey      string
	EtcdCACert         string
	BBSAddress         string
}

func (args Args) ArgSlice() []string {
	return []string{
		"-registerWithRouter=" + strconv.FormatBool(args.RegisterWithRouter),
		"-domainNames", args.DomainNames,
		"-address", args.Address,
		"-taskHandlerAddress", args.TaskHandlerAddress,
		"-etcdCluster", args.EtcdCluster,
		"-username", args.Username,
		"-password", args.Password,
		"-natsAddresses", args.NatsAddresses,
		"-natsUsername", args.NatsUsername,
		"-natsPassword", args.NatsPassword,
		"-corsEnabled=" + strconv.FormatBool(args.CORSEnabled),
		"-consulCluster", args.ConsulCluster,
		"-etcdCertFile", args.EtcdClientCert,
		"-etcdKeyFile", args.EtcdClientKey,
		"-etcdCaFile", args.EtcdCACert,
		"-bbsAddress", args.BBSAddress,
	}
}

func New(binPath string, args Args) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:       "receptor",
		Command:    exec.Command(binPath, args.ArgSlice()...),
		StartCheck: "receptor.started",
	})
}
