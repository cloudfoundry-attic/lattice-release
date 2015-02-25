# Diego API Overview

Diego is Cloud Foundry's next generation runtime.

Though Cloud Foundry's specific use-case (12-factor web applications staged via Heroku-style buildpacks) is the crucible in which Diego was forged, Diego has emerged as a more general distributed system capable of running and monitoring work on a cluster.

A Diego cluster is comprised of a series of VMs called Cells that can run two distinct types of work:

- [**Tasks**](tasks.md) are one-off processes that Diego guarantees will run at most once.
- [**Long-Running Processes**](lrps.md) (LRPs) are processes that Diego launches and monitors.  Diego can distribute, run, and monitor `N` instances of a given LRP.  When an LRP instance crashes, Diego restarts it automatically.

Tasks and LRPs ultimately run in [Garden](http://github.com/cloudfoundry-incubator/garden) containers on Diego Cells.  The filesystem mounted into these containers can either be a generic rootfs that ships with Diego or an arbitrary Docker image.  Processes spawned in these containers are provided with a set of [environment variables](environment.md) to aid in configuration.

In addition to launching and monitoring Tasks and LRPs, Diego can stream logs (via [doppler](http://github.com/cloudfoundry/loggregator)) out of the container processes to end users, and Diego can route (via the [router](http://github.com/cloudfoundry/gorouter)) incoming web traffic to container processes.

Consumers of Diego communicate to Diego via an http API.  This API allows you to schedule Tasks and LRPs and to fetch information about running Tasks and LRPs.  While it is possible to run multi-tenant workload on Diego, the API does not provide strong abstractions and protections around managing such work (e.g. users, organizations, quotas, etc...).  Diego simply runs Tasks and LRPs and it is up to the consumer to provide these additional abstractions.  In the case of Cloud Foundry these responsibilities fall on the [Cloud Controller](http://github.com/cloudfoundry/cloud_controller_ng)

The API is served by a component living on each Diego Cell called the [Receptor](http://github.com/cloudfoundry-incubator/receptor).  The [GitHub repository](http://github.com/cloudfoundry-incubator/receptor) includes a Golang client that consumers can use to interact with the Receptor.

[back](README.md)
