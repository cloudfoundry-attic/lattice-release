package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/bbs"
	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/natbeat"
	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/receptor/task_handler"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/lock_bbs"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/gunk/diegonats"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/localip"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var registerWithRouter = flag.Bool(
	"registerWithRouter",
	false,
	"Register this receptor instance with the router.",
)

var serverDomainNames = flag.String(
	"domainNames",
	"",
	"Comma separated list of domains that should route to this server.",
)

var serverAddress = flag.String(
	"address",
	"",
	"The host:port that the server is bound to.",
)

var taskHandlerAddress = flag.String(
	"taskHandlerAddress",
	"127.0.0.1:1169", // "taskhandler".each_char.collect(&:ord).inject(:+)
	"The host:port for the internal task completion callback",
)

var consulCluster = flag.String(
	"consulCluster",
	"",
	"comma-separated list of consul server URLs (scheme://ip:port)",
)

var lockTTL = flag.Duration(
	"lockTTL",
	lock_bbs.LockTTL,
	"TTL for service lock",
)

var corsEnabled = flag.Bool(
	"corsEnabled",
	false,
	"Enable CORS",
)

var username = flag.String(
	"username",
	"",
	"Username for basic auth, enables basic auth if set.",
)

var password = flag.String(
	"password",
	"",
	"Password for basic auth.",
)

var natsAddresses = flag.String(
	"natsAddresses",
	"",
	"Comma-separated list of NATS addresses (ip:port).",
)

var natsUsername = flag.String(
	"natsUsername",
	"",
	"Username to connect to nats.",
)

var natsPassword = flag.String(
	"natsPassword",
	"",
	"Password for nats user.",
)

var communicationTimeout = flag.Duration(
	"communicationTimeout",
	10*time.Second,
	"Timeout applied to all HTTP requests.",
)

var bbsAddress = flag.String(
	"bbsAddress",
	"",
	"Address to the BBS Server",
)

const (
	dropsondeDestination = "localhost:3457"
	dropsondeOrigin      = "receptor"

	bbsWatchRetryWaitDuration = 3 * time.Second
)

func main() {
	cf_debug_server.AddFlags(flag.CommandLine)
	cf_lager.AddFlags(flag.CommandLine)
	etcdFlags := etcdstoreadapter.AddFlags(flag.CommandLine)
	flag.Parse()

	cf_http.Initialize(*communicationTimeout)

	logger, reconfigurableSink := cf_lager.New("receptor")
	logger.Info("starting")

	initializeDropsonde(logger)

	etcdOptions, err := etcdFlags.Validate()
	if err != nil {
		logger.Fatal("etcd-validation-failed", err)
	}

	if err := validateNatsArguments(); err != nil {
		logger.Error("invalid-nats-flags", err)
		os.Exit(1)
	}

	if err := validateBBSAddress(); err != nil {
		logger.Fatal("invalid-bbs-address", err)
	}

	legacyBBS := initializeReceptorBBS(etcdOptions, logger)

	bbs := bbs.NewClient(*bbsAddress)

	handler := handlers.New(bbs, legacyBBS, logger, *username, *password, *corsEnabled)

	worker, enqueue := task_handler.NewTaskWorkerPool(legacyBBS, logger)
	taskHandler := task_handler.New(enqueue, logger)

	members := grouper.Members{
		{"server", http_server.New(*serverAddress, handler)},
		{"worker", worker},
		{"task-complete-handler", http_server.New(*taskHandlerAddress, taskHandler)},
	}

	if *registerWithRouter {
		registration := initializeServerRegistration(logger)
		natsClient := diegonats.NewClient()
		members = append(members, grouper.Member{
			Name:   "background-heartbeat",
			Runner: natbeat.NewBackgroundHeartbeat(natsClient, *natsAddresses, *natsUsername, *natsPassword, logger, registration),
		})
	}

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func validateBBSAddress() error {
	if *bbsAddress == "" {
		return errors.New("bbsAddress is required")
	}
	return nil
}

func validateNatsArguments() error {
	if *registerWithRouter {
		if *natsAddresses == "" || *serverDomainNames == "" {
			return errors.New("registerWithRouter is set, but nats addresses or domain names were left blank")
		}
	}
	return nil
}

func initializeDropsonde(logger lager.Logger) {
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}

func initializeReceptorBBS(etcdOptions *etcdstoreadapter.ETCDOptions, logger lager.Logger) Bbs.ReceptorBBS {
	workPool, err := workpool.NewWorkPool(100)
	if err != nil {
		logger.Fatal("failed-to-construct-etcd-adapter-workpool", err, lager.Data{"num-workers": 100}) // should never happen
	}

	etcdAdapter, err := etcdstoreadapter.New(etcdOptions, workPool)

	if err != nil {
		logger.Fatal("failed-to-construct-etcd-tls-client", err)
	}

	client, err := consuladapter.NewClient(*consulCluster)
	if err != nil {
		logger.Fatal("new-client-failed", err)
	}

	sessionMgr := consuladapter.NewSessionManager(client)
	consulSession, err := consuladapter.NewSession("receptor", *lockTTL, client, sessionMgr)
	if err != nil {
		logger.Fatal("consul-session-failed", err)
	}

	return Bbs.NewReceptorBBS(etcdAdapter, consulSession, *taskHandlerAddress, clock.NewClock(), logger)
}

func initializeServerRegistration(logger lager.Logger) (registration natbeat.RegistryMessage) {
	domains := strings.Split(*serverDomainNames, ",")

	addressComponents := strings.Split(*serverAddress, ":")
	if len(addressComponents) != 2 {
		logger.Error("server-address-invalid", fmt.Errorf("%s is not a valid serverAddress", *serverAddress))
		os.Exit(1)
	}

	host, err := localip.LocalIP()
	if err != nil {
		logger.Error("local-ip-invalid", err)
		os.Exit(1)
	}

	port, err := strconv.Atoi(addressComponents[1])
	if err != nil {
		logger.Error("server-address-invalid", fmt.Errorf("%s does not have a valid port", *serverAddress))
		os.Exit(1)
	}

	return natbeat.RegistryMessage{
		URIs: domains,
		Host: host,
		Port: port,
	}
}
