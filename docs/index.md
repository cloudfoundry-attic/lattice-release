# Lattice FAQ

## <a name="what-is-lattice"><a href="#what-is-lattice">What is Lattice?</a></a>

Lattice is an open source project for running containerized workloads on a cluster.  A Lattice cluster is comprised of a number of Lattice Cells (VMs that run containers) and a 'Brain' that monitors the Cells.

Lattice includes built-in http load-balancing, a cluster scheduler, log aggregation with log streaming and health management.

Lattice containers are described as long-running processes or temporary tasks. Lattice includes support for Linux Containers expressed either as Docker Images or by composing applications as binary code on top of a root file system. Lattice's container pluggability will enable other backends such as [Windows](https://www.youtube.com/watch?v=S4U_YyzC5z4) or [Rocket](http://blog.pivotal.io/cloud-foundry-pivotal/news-2/launching-rockets-collaborating-on-next-level-linux-containers) in the future.

All this functionality leverages components of [Cloud Foundry](https://github.com/cloudfoundry). The new parts in Lattice are the CLI to interact with the component APIs and the installers.

## <a name="why-did-we-create-lattice"><a href="#why-did-we-create-lattice">Why did we create Lattice?</a></a>

We wanted to be able to leverage components of Cloud Foundry independently to demonstrate the capabilities and facilitate more rapid experimentation. We believe these components are already useful in their own right, but will also improve with more wide adoption and exposure.

We also have a strong opinions on the developer and operator experience for scalable applications that we hoped to express with Lattice. Our foundational principles are:

- **Simple**: you should be able to get started in minutes.
- **Comprehensive**: include capabilities commonly required for scalable applications.
- **Extensible**: you should be able to use it how you like, adding and taking away things.


Lattice lives up to these principles.

- **Simple**: a [small Vagrant VM](https://github.com/cloudfoundry-incubator/lattice#local-deployment) easily fits on most laptops. [Terraform scripts](https://github.com/cloudfoundry-incubator/lattice#clustered-deployment) enable you to easily start a Lattice cluster on your cloud of choice. Very few new concepts are required to use Lattice and you can [get started in minutes](/docs/getting-started.md).
- **Comprehensive**: Lattice currently includes load balancing, aggregated logs, health management, and cluster scheduling.
- **Extensible**: Docker is very popular and we are supporting Docker Images. We also believe in a [pluggable container backend](https://github.com/cloudfoundry-incubator/garden) so there are options to support [Windows](https://www.youtube.com/watch?v=S4U_YyzC5z4), [Rocket](http://blog.pivotal.io/cloud-foundry-pivotal/news-2/launching-rockets-collaborating-on-next-level-linux-containers) or other backends.

## <a name="how-do-i-deploy-lattice"><a href="#how-do-i-deploy-lattice">How do I deploy Lattice?</a></a>

You can run Lattice easily on your laptop with a [Vagrant VM](https://github.com/cloudfoundry-incubator/lattice#local-deployment) or deploy a [cluster of machines](https://github.com/cloudfoundry-incubator/lattice#clustered-deployment) with [AWS](https://github.com/cloudfoundry-incubator/lattice/blob/v0.3.1/terraform/aws/README.md), [Digital Ocean](https://github.com/cloudfoundry-incubator/lattice/blob/v0.3.1/terraform/digitalocean/README.md), [Google Cloud](https://github.com/cloudfoundry-incubator/lattice/blob/v0.3.1/terraform/google/README.md) or [Openstack](https://github.com/cloudfoundry-incubator/lattice/blob/v0.3.1/terraform/openstack/README.md).

Up-to-date installation instructions are on the [GitHub README](https://github.com/cloudfoundry-incubator/lattice/tree/v0.3.1).

## <a name="how-do-i-use-lattice"><a href="#how-do-i-use-lattice">How do I use Lattice?</a></a>

Start with the [getting-started tutorial](/docs/getting-started.md) for a brief introduction.

Lattice provides an [HTTP API](/docs/lattice-api.md) for scheduling and monitoring work.  [`ltc` (the Lattice CLI)](/docs/ltc.md)  wraps this API and provides basic access to Lattice’s feature-set.

## <a name="im-having-trouble-launching-my-docker-image---help"><a href="#im-having-trouble-launching-my-docker-image---help">I'm having trouble launching my Docker image - help!</a></a>

Check out the [troubleshooting section](/docs/troubleshooting.md).

## <a name="is-lattice-intended-for-multi-tenant-deployments"><a href="#is-lattice-intended-for-multi-tenant-deployments">Is Lattice intended for multi-tenant deployments?</a></a>

No.  While built on top of a tech stack that supports secure multi-tenancy, Lattice is geared towards the developers or small teams with a high trust culture that need cluster computing solutions.  Lattice gives you "root access" to your cluster and minimizes as many barriers to entry as possible - including security barriers!

## <a name="what-software-components-make-up-lattice"><a href="#what-software-components-make-up-lattice">What software components make up Lattice?</a></a>

Lattice components are primarily written in Go.

Lattice includes Cloud Foundry components [Diego](https://github.com/cloudfoundry-incubator/diego-design-notes), [Loggregator](https://github.com/cloudfoundry/loggregator) and [Router](https://github.com/cloudfoundry/gorouter).  Diego is responsible for scheduling and running containerized workloads.  Loggregator is responsible for streaming logs out of running containers.  The Router is responsible for load balancing HTTP traffic across running containers.

## <a name="what-is-the-relationship-between-lattice-and-diego"><a href="#what-is-the-relationship-between-lattice-and-diego">What is the relationship between Lattice and Diego?</a></a>

[Diego](https://github.com/cloudfoundry-incubator/diego-design-notes) is Cloud Foundry's next generation elastic runtime.  Lattice is a distribution of Diego that eschews most of Cloud Foundry's 'enterprise' features.  Lattice is built to let you easily deploy and interact with Diego.  New features that make their way into Diego will be available on Lattice almost immediately.

To build a deep understanding of how Lattice works you'll need to learn about Diego.  The [design notes](https://github.com/cloudfoundry-incubator/diego-design-notes) are a good starting point for building a mental model of what is going on under the hood.

## <a name="what-is-the-relationship-between-lattice-and-docker"><a href="#what-is-the-relationship-between-lattice-and-docker">What is the relationship between Lattice and Docker?</a></a>

Lattice supports Docker images as a format for distributing container root filesystems.  Currently, Docker images must be publicly hosted on the [Docker Hub Registry](https://registry.hub.docker.com) or published to a signed private Docker registry without authentication.  Lattice uses Docker's libraries to fetch Docker image metadata and image layers, but it does *not* use the Docker daemon to run and manage containers.  Instead, Lattice uses [Garden](https://github.com/cloudfoundry-incubator/garden).  Garden provides a *platform-agnostic* API for launching and managing containers and is built to be consumed by a distributed container scheduler like Diego.  [Garden-Linux](https://github.com/cloudfoundry-incubator/garden-linux) is an implementation of the Garden API that provides containers on the Linux platform using kernel namespaces and cgroups - the same technologies used by Docker, LXC and Rocket.

More details about how Lattice works with Docker images can be found [here](/docs/troubleshooting.md#how-does-lattice-work-with-docker-images).

## <a name="what-about-cloud-foundry-elastic-runtime-kubernetes-mesos-and-other-clustered-docker-projects"><a href="#what-about-cloud-foundry-elastic-runtime-kubernetes-mesos-and-other-clustered-docker-projects">What about Cloud Foundry Elastic Runtime, Kubernetes, Mesos and other clustered Docker projects?</a></a>

- **Cloud Foundry**: For small teams and some use cases, a complete Cloud Foundry deployment is overkill, can be difficult to get started and requires a footprint that does not fit on most developer laptops. We also wanted to make it possible to use Cloud Foundry components like the Router on their own and experiment more easily with new ideas before graduating them to full Cloud Foundry.  Lattice reuses Diego, Cloud Foundry’s next generation Elastic Runtime.

- **Kubernetes**: Kubernetes, from Google, has a lot of great ideas and opinions that overlap with Lattice and some that don't. Kubernetes requires overlay networking and does not include features we believe are important like DNS load balancing and aggregated logging. At this time, Kubernetes only includes simple cluster scheduling in open source and does not use the cluster scheduler we imagine Google uses for production workloads. Kubernetes also requires awareness and implementations of concepts specific to Kubernetes like pods, replicaControllers and services that do not fit the simple user experience we envision.

- **Mesos**: Apache Mesos is also a great project which provides distributed scheduling. Mesos is managing more of lower level primitives in the data center and requires applications be written for Mesos or the use of frameworks built on top of Mesos to target specific use cases. Some frameworks like Marathon provide higher level abstractions, but do not include features we wanted and do not fit the simpler developer and operator experience we are targeting.

- **Other clustered container projects**: Most of the clustered Docker projects are very Docker-centric, missing features we wanted and do not have a pluggable container solution that would easily accommodate Windows or Rocket. There is obviously overlap and differences with many other projects. Lattice is an attempt to balance maximizing the utility of scalable clustering with a minimal feature set in line with our goals for easy deployment and a great user experience.

## <a name="is-lattice-going-to-replace-the-existing-cloud-foundry-elastic-runtime"><a href="#is-lattice-going-to-replace-the-existing-cloud-foundry-elastic-runtime">Is Lattice going to replace the existing Cloud Foundry Elastic Runtime?</a></a>

No. Lattice and Cloud Foundry Elastic Runtime serve different purposes with some of the same building blocks. Lattice provides a single user, single-tenant, cluster root, very-few-constraints developer experience without guard rails. Cloud Foundry Elastic Runtime serves full enterprise use cases including multi-user team development, multi-tenancy, auditing, services marketplace and quotas with enterprise guard rails. Lattice is intended to be a proving ground for developer and operator experience that influences features for Cloud Foundry Elastic Runtime.

## <a name="what-cloud-foundry-projects-are-not-included-with-lattice"><a href="#what-cloud-foundry-projects-are-not-included-with-lattice">What Cloud Foundry projects are not included with Lattice?</a></a>

- **UAA and Login Server**: Lattice is intended for single-tenancy. It is not intended to provide multi-tenancy for isolation as an API user effectively has cluster root permissions. Single user “basic auth” is currently the only authentication option.
- **Cloud Controller**: Cloud Foundry’s API server has many enterprise features for multi-tenancy, a services marketplace and other enterprise abstractions that are not necessary for some basic container use cases.
- **BOSH Mandated Packaging**: Any configuration management tool should work well with Lattice. We still recommend BOSH for production, but we recognize there are many choices for configuration and cluster lifecycle management.

## <a name="how-is-lattice-governed"><a href="#how-is-lattice-governed">How is Lattice governed?</a></a>

[Pivotal](http://www.pivotal.io) plans on submitting Lattice to the [Cloud Foundry Foundation incubation process](https://docs.google.com/a/pivotal.io/document/d/1_Jh-Dobalgie-w6Yp1-QSGKf0LEpBBit-mB3oYCBbtg/edit#heading=h.gjdgxs). Product direction is provided by a Product Manager, currently Marco Nicosia. Engineering governance is handled like other projects at Cloud Foundry Foundation which include Test Driven Development, Continuous Integration and other quality standards. [See CONTRIBUTING.md](https://github.com/cloudfoundry/cf-release/blob/master/CONTRIBUTING.md) for more detail.

## <a name="how-do-i-contribute-or-get-involved-with-lattice"><a href="#how-do-i-contribute-or-get-involved-with-lattice">How do I contribute or get involved with Lattice?</a></a>

- Review the public [Lattice](https://www.pivotaltracker.com/n/projects/1183596) and [Diego](https://www.pivotaltracker.com/n/projects/1003146) Pivotal Tracker projects
- Get involved on the [mailing list](https://lists.cloudfoundry.org/mailman/listinfo/cf-lattice) ([archives](http://cf-lattice.70370.x6.nabble.com/))
- Open issues on the GitHub repository for [Lattice](https://github.com/cloudfoundry-incubator/lattice)

## <a name="what-is-the-lattice-roadmap"><a href="#what-is-the-lattice-roadmap">What is the Lattice roadmap?</a>

See the public Pivotal Tracker project for [Lattice](https://www.pivotaltracker.com/n/projects/1183596) and [Diego](https://www.pivotaltracker.com/n/projects/1003146)

## <a name="is-lattice-ready-for-production"><a href="#is-lattice-ready-for-production">Is Lattice ready for production?</a></a>

Diego, the runtime component of Lattice is not in full production yet with Cloud Foundry, but is running a subset of [Pivotal Web Services](https://run.pivotal.io/) deployments. Diego is planned for full production use sometime early in Q2 of 2015.

The Router and Loggregator components of Lattice have been in production with Cloud Foundry for over a year and we consider them production-ready.

The deployment strategy employed by Lattice emphasizes convenience and simplicity.  When deployed with Terraform none of the Lattice VMs are monitored - should a VM fail it will not be recreated.  Moreover, Lattice minimizes overhead by running single instances of several components.  These constitute single-points of failure and should not be relied upon in a production setting.  If you need to set up a productionized deployment of Lattice you will need to be sure to make all the components redundant, plus provide monitoring and management on the hosts. As the project evolves and there is more experience in the community, we hope those solutions will emerge and be shared. Depending on the use case, we also recommend considering deploying Cloud Foundry with BOSH. Instructions for this are available on the [Diego-Release GitHub repository](https://github.com/cloudfoundry-incubator/diego-release).

## <a name="how-secure-is-lattice"><a href="#how-secure-is-lattice">How secure is Lattice?</a></a>

Security is not Lattice's primary concern.  Lattice is intended to provide a cluster root experience with minimal barriers to entry.  Multi-tenant workloads are not in scope.  Additionally, no effort is made to prevent containers from communicating with other components within a Lattice cluster.  Also, no effort is made to protect log streams behind an authentication layer - they are accessible to clients with network access to Lattice servers.

As the project evolves and there is more experience in the community, we hope security solutions will emerge and be shared. Depending on the use case and security needs, we also recommend considering deploying Cloud Foundry with BOSH. Instructions for this are available on the [Diego-Release GitHub repository](https://github.com/cloudfoundry-incubator/diego-release).

## <a name="what-operating-systems-are-supported"><a href="#what-operating-systems-are-supported">What Operating Systems are supported?</a></a>

Lattice currently is developed and tested using Ubuntu Trusty 14.04 LTS.

Lattice uses Cloud Foundry's [Garden-Linux](https://github.com/cloudfoundry-incubator/garden-linux) for Linux Containers, and the current Garden-Linux Backend does not yet support CentOS7. We would like to support CentOS7 and other Linux variants with a modern Linux 3.x kernel.  Get involved on the mailing list if you are interested in helping with this.

## <a name="how-much-memory-footprint-does-lattice-have"><a href="#how-much-memory-footprint-does-lattice-have">How much memory footprint does Lattice have?</a></a>

When running all of Lattice on a single VM, the process overhead in-use is about 250MB with no containers and no activity.

When running Lattice in distributed mode with no containers and no activity, the Cell uses about 156MB of process overhead and the Brain uses about 141MB.
