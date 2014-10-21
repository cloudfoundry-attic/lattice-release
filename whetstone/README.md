#Assumptions
- Loggregator is deployed and its traffic controller job has the allowAllAccess flag set to true. 
  Currently our forked version of loggregator (https://github.com/dajulia3/loggregator) supports this flag.  
- Diego-release is deployed
    
#Running The Whetstone Tests

For example, to Run Tests against Bosh Lite deployed Diego Release and Loggregator:
    ginkgo -r -untilItFails -- -etcdAddress="10.244.16.2:4001" -domain="10.244.0.34.xip.io" -loggregatorAddress="loggregator.10.244.0.34.xip.io"
