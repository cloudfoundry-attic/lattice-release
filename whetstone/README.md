---Running The Whetstone Tests

To Run Tests against Bosh Lite:
ginkgo -r -untilItFails -- -etcdAddress="http://10.244.16.2:4001" -domain="10.244.0.34.xip.io"
