---
title: ltc - The Lattice CLI
weight: 4
doc_subnav: true
---

# `ltc` - The Lattice CLI

`ltc` wraps the [Lattice API](/docs/lattice-api.html) and provides a simple interface for launching and managing applications based on Docker images.  This document is a reference detailing all of `ltc`'s subcommands and options.  You may find it helpful to also build a deeper understanding of how Lattice [manages applications](/docs/troubleshooting.html#how-does-lattice-manage-applications) and [Docker images](/docs/troubleshooting.html#how-does-lattice-work-with-docker-images).

## Targetting Lattice

### `ltc target`

You use `ltc target` to point `ltc` at a deployed Lattice installation. For a Vagrant deployed local Lattice the default should be: `ltc target 192.168.11.11.xip.io`

If the Lattice API is password protected, `ltc` will prompt you for a username and password.

Run `ltc target` with no arguments to get the current target.

## Launching and Managing Applications

### `ltc start`

`ltc start APP_NAME DOCKER_IMAGE` launches Docker image based applications in a Lattice cluster.

- `APP_NAME` is required and must be unique across the Lattice cluster.  `APP_NAME` is used to refer to the application and to route to the application.  For example, an application named `lattice-app` will be accessible at `lattice-app.192.168.11.11.xip.io`
- `DOCKER_IMAGE` is required and must match the standard Docker image format (e.g. `cloudfoundry/lattice-app`)

When launching a Docker image, `ltc` first queries the Docker registry for metadata associated with the image.  It uses this information to:

- construct the start command based on the `ENTRYPOINT` and `CMD` associated with the Docker image
- identify the working directory based on the `WORKDIR` associated with the Docker image
- open up ports based on any `EXPOSE` directives associated with the Docker image

With this metadata in hand, `ltc` submits a request to launch the application to Lattice.  This request includes information on how to monitor the health of the application. 

The default behavior of `ltc start`, outlined above, can be modified via a series of additional command line flags:

- **--working-dir=/path/to/working-dir** sets the working directory, overriding the default associated with the Docker image.
- **--run-as-root** launches the command in the process as the root user.  By default, Lattice uses a non-root user created at container-creation time.  Lattice does not yet honor the Docker USER directive.  There are plans to address this soon.  For most containers `--run-as-root` is a sufficient workaround.
- **--env NAME=VALUE** specifies environment variables. You can have multiple `--env` flags.  These are merged *on top of* the Environment variables extracted from the Docker image metadata.
- **--memory-mb=128** specifies the memory limit to apply to the container.  To allow unlimited memory usage, set this to 0.
- **--disk-mb=1024** specifies the disk limit to apply to the container.  This governs any writes *on top of* the root filesystem mounted into the container.  To allow unlimited disk usage, set this to 0.
- **--port=8080** specifies the port to open on the container.  This overrides any `EXPOSE` directives associated with the Docker image.  It is currently only possible to open one port via the `ltc` cli.  The specified port is also used for the purposes of health monitoring (see above).
- **--instances=1** specifies the number of instances of the application to launch.  This can also be modified after the application is started.
- **--no-monitor** disables health monitoring.  Lattice will consider the application crashed only if it exits.

Finally, one can override the default start command by specifiying a start command after a `--` separator.  For example:

    ltc start lattice-app cloudfoundry/lattice-app -- /lattice-app -quiet=true

> For instances with *multiple* `EXPOSE` directives, `ltc` selects the *lowest* port for the purposes of routing traffic and performing the health check.

### `ltc scale` 

`ltc scale APP_NAME NUM_INSTANCES` modifies the number of running instances of an application.

### `ltc stop`

`ltc stop APP_NAME` is equivalent to `ltc scale APP_NAME 0`

### `ltc remove`

`ltc remove APP_NAME` removes an application entirely from a Lattice deployment.

## Streaming Logs

### `ltc logs`

`ltc logs APP_NAME` attaches to a log stream for a running application.  The logstream aggregates logs from *all* instances associated with an application.

## What's Running on Lattice?

### `ltc list`

`ltc list` displays all applications currently running on the targetted Lattice deployment.  This includes information on the number of requested and running instances for each application, and routing information for accessing the application.

### `ltc status`

`ltc status APPLICATION_NAME` provides detailed information about an application running on the Lattice deployment.

The first section in the status report includes information about the *desired* state of the application: how many instances should be running, what route to associate with the application, etc..

The subsequent sections include information about individual instances of the application.  Here's some example output:

    ================================================================================
          Instance 0  [RUNNING]
    --------------------------------------------------------------------------------
    InstanceGuid    b75f58d2-4fc6-4c38-7097-42caf32ffc07
    Cell ID         lattice-cell-01
    Ip              192.168.11.11
    Ports           61001:7777;61002:9999
    Since           2015-02-06 16:52:40 (PST)
    --------------------------------------------------------------------------------

This indicates that instance 0 of the application has been `RUNNING` on `lattice-cell-01` since `2015-02-06 16:52:40 (PST)`.  The application can be reached at the indicated ip address (`192.168.11.11`).  This particular application included two EXPOSE directives, one for port  `7777` the other for port `9999`.  The `Ports` section of the report indicates the host-side ports that can be used to connect to the application at the requested container-side ports.  For example, to connect to `7777` one goes to `192.168.11.11:61001`.  To connect to `9999` one goes to `192.168.11.11:61002`.

### `ltc visualize`

`ltc visualize` displays the *distribution* of application instances across the targetted Lattice deployment.  Each running application is rendered as a green dot.  Starting applications are rendered as yellow dots.

