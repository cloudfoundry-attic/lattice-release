package whetstone_test

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	domain  string
	timeout int
	numCpu  int
)

func init() {
	numCpu = runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	flag.StringVar(&domain, "domain", "", "Domain of Diego - REQUIRED")
	flag.IntVar(&timeout, "timeout", 30, "How long whetstone will wait for docker apps to start")
}

func TestWhetstone(t *testing.T) {
	if domain == "" {
		fmt.Fprintf(os.Stderr, "To run this test suite, you must set the required flags.\nUsage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Whetstone Suite")
}
