# Troubleshooting

## How does Lattice manage applications?

A helpful step when debugging is to have an accurate mental model of the system in question.

Lattice is founded upon the notions of eventual consistency.  In particular, Lattice is constantly working to reconcile *desired* state with *actual* state.

When you issue commands via the `ltc` CLI (or the Lattice API) you are modifying Lattice's *desired* state.  Typically, you are informing Lattice about a desire to run some number of instances of an application.  Lattice updates this desired state **synchronously**.

The *actual* state (i.e. the set of running instances), however, is updated **asynchronously** as the Lattice cluster works to reconcile the current *actual* state with the *desired* state.

Typically, when the user updates *desired* state Lattice immediately takes actions to perform this reconciliation.  Should an action fail (perhaps a network partition occurs) or a running instance be lost (perhaps a Cell explodes) Lattice will eventually attempt to reconcile actual and desired state again (this happens every 30 seconds - though Lattice can detect a missing Cell within ~5 seconds).

## How does Lattice work with Docker images?

A Docker image consists of two things: a collection of layers to download and mount (the raw bits that form the file system) and metadata that describes what command should be launched (the `ENTRYPOINT` and `CMD` directives, among others, specified in the Dockerfile).

Lattice uses [Garden-Linux](https://github.com/cloudfoundry-incubator/garden-Linux) to construct Linux containers.  These containers are built on the same Linux kernel technologies that power all Linux containers: namespaces and cgroups.  When a container is created a file system must be mounted as the root file system of the container.  Garden-Linux supports mounting Docker images as root file systems for the containers it constructs.  Garden-Linux takes care of fetching and caching the individual layers associated with the Docker image and combining and mounting them as the root file system - it does this using the same libraries that power Docker.

This yields a container with contents that exactly match the contents of the associated Docker image.

Once a container is created Lattice is responsible for running and monitoring processes in the container.  The Lattice API allows the user to define exactly which commands to run within the container; in particular, it is possible to run, monitor, and route to *multiple* processes within a single container.

When launching a Docker image, `ltc` directs Lattice to create a container backed by the Docker image's root fs, and to run the command encoded in the Docker image's metadata.  It does this by fetching the metadata associated with the Docker image (using the same libraries that power Docker) and making the appropriate Lattice API calls.  `ltc` allows users to easily override the values it pulls out of the Docker image metadata.  This is outlined in detail in the [`ltc` documentation](/docs/ltc.html#ltc-start).

There are two remaining areas of Docker compatbility that we are working on:

- Removing assumptions about container contents.  Currently, Garden-Linux makes some assumptions about what is available inside the container.  Some Docker images do not satisfy these assumptions though most do (the liteweight busybox base image, for example).
- Supporting arbitrary UIDs and GIDs.  Currently Garden-Linux runs applications as the `vcap` user (a historical holdover).  One can side-step this with `--run-as-root` (see below) though this is suboptimal.  We intend to fully support the USER directive and (moreover) to improve our API around specifying which user should run the command.

## I can't run my Docker image.  Help!

Here are a few pointers to help you debug and fix some common issues:

### Increase `ltc`'s Timeout

`ltc create` will wait up to one minute for your application to start.  If this fails it may be that your Docker container is large and has not downloaded yet.  You can set the `LATTICE_CLI_TIMEOUT` environment variable (in seconds) to instruct `ltc` to wait longer.  Note that `ltc` does not remove your application when this timeout occurs, so your application may eventually start in the background.

### Increase Memory and Disk Limits

By default, `ltc` applies a memory limit of 128MB and a disk limit of 1024MB to the container.  If your process is exiting prematurely it may be attempting to consume more than 128MB of memory.  You can increase the limit using the `--memory-mb` flag.  To turn off memory and disk limits set `--memory-mb` and `--disk-mb` to `0`.

### Check the Application Logs

`ltc logs APP_NAME` will aggregate and stream application logs.  These may point you in the right direction.  In particular, if you see issues related to file permissions or a health check failing, read on...

### Run as Root

Lattice runs the process in your Docker image as an unprivileged user.  Sometimes this user does not have privileges to execute the requested process - you can try using the `--run-as-root` flag to get around this limitation.

> We have plans to build more robust support for specifying the user/uid/group/gid to run the container as.

### Disable Health Monitoring

By default, `ltc` requests that Lattice perform a periodic health check agains the running application.  This health check verifies that the application is listening on a port.  For applications that do not listen on ports (e.g. a worker that does not expose an endpoint) you can disable the health check via the `--no-monitor` flag.

### Watch Lattice Component Logs

If you're still stuck you can try streaming the Lattice debug-logs with `ltc debug-logs` while launching your application. [The veritas cli](https://github.com/pivotal-cf-experimental/veritas) is helpful for pretty printing the json log format from `ltc debug-logs` into something more human readable with a command like `ltc debug-logs | veritas chug`.

### How do I get a shell inside a lattice container.

See this detailed step-by-step set of instructions for how to [get shell access to a container running on lattice](https://docs.google.com/a/pivotal.io/document/d/1WWoQ_d5nR4-P6VfLbAAbzOZIvRj-Xdff2hsjM_ZWRUQ/edit#heading=h.hwnzq0ni9hoj).

## How do I communicate with my containers over TCP?

The Lattice router only supports HTTP communication at this time.  If you would like to use TCP instead you will need to communicate with the container by IP and port, which you can get via `ltc status`.  For a local vagrant deployed Lattice, the containers can be reached at `192.168.11.11`.  On AWS, you will need to configure your ELB to route traffic to the Cells in your VPC.

## How do I communicate between containers?

Lattice does not apply any firewall rules between containers.  Any container can freely communicate with any other container.  All you need is to identfiy the IP and Port - information available via `ltc status` or the [Receptor API](https://github.com/cloudfoundry-incubator/receptor/blob/master/doc/README.md).

## How do I do service discovery?

Outside of the HTTP router, Lattice does not ship with a service discovery solution.  It is relatively straightforward, however, to build a solution on top of the Receptor API.  We have plans to explore this space soon after release.

## How do I upgrade Lattice?

Lattice [does not support rolling upgrades](/docs/#is-lattice-ready-for-production).
