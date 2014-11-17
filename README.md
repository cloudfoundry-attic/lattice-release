#Assumptions
- Loggregator is deployed and its traffic controller job has the allowAllAccess flag set to true. 
  Currently master of Loggregator supports this flag, but the version in cf-release is behind.  
- Diego-release is deployed
    
#Running The Whetstone Tests

For example, to Run Tests against Bosh Lite deployed Diego Release and Loggregator with a 30 sec app start timeout:
     
     ginkgo -- -receptorAddress="receptor.10.244.16.2.xip.io" -domain="10.244.0.34.xip.io" -loggregatorAddress="loggregator.10.244.0.34.xip.io" -timeout=30

To run against [Diego Lite](https://github.com/pivotal-cf-experimental/diego-lite)

    ginkgo -- -receptorAddress="receptor.192.168.11.11.xip.io" -domain="192.168.11.11.xip.io" -loggregatorAddress="loggregator.192.168.11.11.xip.io" -timeout=30
   

#Notes on Running against Bosh Lite:
  Cloudfoundry reccomends using xip.io with Bosh lite for DNS.
  This has been very flaky for us, resulting in no such host errors.
  The alternative that we have found is to use dnsmasq configured to resolve all xip.io addresses to the ip of the HA proxy.
  This also requires creating a /etc/resolvers/io file that points to 127.0.0.1. See further instructions [here] (http://passingcuriosity.com/2013/dnsmasq-dev-osx/). 
