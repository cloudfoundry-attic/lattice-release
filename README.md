#Whetstone

Integration tests for Lattice

##Assumptions
- Lattice is deployed with a publicly accessible Receptor API
    
##Setup

     go get github.com/pivotal-cf-experimental/whetstone
     go get github.com/pivotal-cf-experimental/lattice-cli/ltc
     go get -v -t ./...

     go get github.com/onsi/ginkgo/ginkgo
     go get github.com/onsi/gomega


##Running The Whetstone Tests

To run against [Lattice](https://github.com/pivotal-cf-experimental/lattice) with a 30 sec app start timeout

    ginkgo -- -domain="192.168.11.11.xip.io" -timeout=30

To run against Lattice with a username and password

    ginkgo -- -domain="12.345.130.132.xip.io" -timeout=30 --username <USERNAME> --password <PASSWORD>
