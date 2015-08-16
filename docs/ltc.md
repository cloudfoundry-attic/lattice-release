# `ltc` - The Lattice CLI

`ltc` wraps the [Lattice API](/docs/lattice-api.md) and provides a simple interface for launching and managing applications based on Docker images.  This document is a reference detailing all of `ltc`'s subcommands and options.  You may find it helpful to also build a deeper understanding of how Lattice [manages applications](/docs/troubleshooting.md#how-does-lattice-manage-applications) and [Docker images](/docs/troubleshooting.md#how-does-lattice-work-with-docker-images).

You can download the CLI from the [GitHub Releases](https://github.com/cloudfoundry-incubator/lattice/releases) page.

## Targeting Lattice

### `ltc target`

You use `ltc target` to point `ltc` at a deployed Lattice installation. For a Vagrant deployed local Lattice the default should be: `ltc target 192.168.11.11.xip.io`

If the Lattice API is password protected, `ltc` will prompt you for a username and password.

Run `ltc target` with no arguments to get the current target.  It will also indicate whether a droplet store is available.

- Starting with v0.3.0, `ltc target` looks for a droplet store bundled with Lattice to enable Buildpacks functionality.

## Launching Docker Applications

### `ltc create`

`ltc create APP_NAME DOCKER_IMAGE` launches Docker image based applications in a Lattice cluster.

- `APP_NAME` is required and must be unique across the Lattice cluster.  `APP_NAME` is used to refer to the application and to route to the application.  For example, an application named `lattice-app` will be accessible at `lattice-app.192.168.11.11.xip.io`
- `DOCKER_IMAGE` is required and must match the standard Docker image format (e.g. `cloudfoundry/lattice-app`)

When launching a Docker image, `ltc` first queries the Docker registry for metadata associated with the image.  It uses this information to:

- construct the start command based on the `ENTRYPOINT` and `CMD` associated with the Docker image
- identify the working directory based on the `WORKDIR` associated with the Docker image
- open up ports based on any `EXPOSE` directives associated with the Docker image

With this metadata in hand, `ltc` submits a request to launch the application to Lattice.  This request includes information on how to monitor the health of the application and how to route traffic to the application.

The default behavior of `ltc create`, outlined above, can be modified via a series of additional command line flags:

- **`--working-dir=/path/to/working-dir`** sets the working directory, overriding the default associated with the Docker image.
- **`--run-as-root`** launches the command in the process as the root user.  By default, Lattice uses a non-root user created at container-creation time.  Lattice does not yet honor the Docker USER directive.  There are plans to address this soon.  For most containers `--run-as-root` is a sufficient workaround.
- **`--env NAME[=VALUE]`** specifies environment variables. You can have multiple `--env` flags.  Passing an `--env` flag without explicitly setting the VALUE uses the current execution context to set the value.
- **`--cpu-weight=100`** specifies the relative CPU weight to apply to the container (scale 1-100).
- **`--memory-mb=128`** specifies the memory limit to apply to the container.  To allow unlimited memory usage, set this to 0.
- **`--disk-mb=1024`** specifies the disk limit to apply to the container.  This governs any writes *on top of* the root filesystem mounted into the container.  To allow unlimited disk usage, set this to 0.
- **`--instances=1`** specifies the number of instances of the application to launch.  This can also be modified after the application is started.
- **`--timeout=2m`** sets the maximum polling duration for starting the app.

Finally, one can override the default start command by specifiying a start command after a `--` separator.  This can be followed by any arguments one wishes to pass to the app.  For example:

    ltc create lattice-app cloudfoundry/lattice-app -- /lattice-app -quiet=true

#### Managing Mulitple Ports

By default, `ltc` requests that Lattice open up all ports specified by the `EXPOSE` directive associated with the Docker image.  It then sets up a route to send HTTP traffic to each exposed port.  For example, an application named `my-app` that exposes ports `8080` and `9000` will get the following set of default routes:

- `my-app.192.168.11.11.xip.io` will map to port `8080` (the bare `my-app` route always routes to the *lowest* exposed port)
- `my-app-8080.192.168.11.11.xip.io` will map to port `8080`
- `my-app-9000.192.168.11.11.xip.io` will map to port `9000`

You can modify all of this behavior from the command line:

- **`--ports=8080,9000`** allows you to specify the set of ports to open on the container.  This overrides any `EXPOSE` directives associated with the Docker image.  
    - When specifying multiple ports via `--port` you should also specify a `--monitor-port` or `--monitor-url` to perform the healthcheck on (or, alternatively, turn off the health-check via `--no-monitor`).
- **`--routes=8080:my-app,9000:my-app-admin`** allows you to specify the routes to map to the requested ports.  In this example, `my-app.192.168.11.11.xip.io` will map to port `8080` and `my-app-admin.192.168.11.11.xip.io` will map to port `9000`.
  - You can comma-delimit multiple routes to the same port (e.g. `--routes=8080:my-app,8080:my-app-alias`).
- **`--no-routes`** allows you to specify that no routes be registered. 
- **`--tcp-routes=6379:50000,6000:50001`** allows you to specify external ports that will be routed to container ports. In this example, given a TCP router at 192.168.11.11, then 192.168.11.11:50000 will be routed to port container port 6379, and 192.168.11.11:50001 will be routed to container port 6000. As Lattice assigned container ports starting at 60000, and the router runs on the same IP as the app container, choose an external port below 60000. You can map multiple external ports to the same container port (e.g. --tcp-routes=6379:50000,6379:50001).


#### Managing Healthchecks

In addition, `ltc` sets up a healthcheck that verifies the application responds on a given port or to a specified HTTP request. 

By default, `ltc` selects the *lowest* exposed port to healthcheck against;  if no monitoring options are specified, 8080 is the default unless `--no-monitor` is set.

- **`--monitor-port=8080`** sets the port that `ltc` performs a port healthcheck against.
- **`--monitor-url=PORT:/path/to/endpoint`** performs an HTTP roundtrip to check whether a given endpoint returns a **200 OK** result.
- **`--monitor-timeout=1s`** sets the wait time for the application to respond to the healthcheck.
- **`--no-monitor`** disables health monitoring.  Lattice will consider the application crashed only if it exits.

## Managing Applications

### `ltc remove`

`ltc remove APP1_NAME [APP2_NAME APP3_NAME...]` removes the specified applications from a Lattice deployment.  

- This operation is performed in the background. `ltc list` may still show the app as running immediately after running `ltc remove`. 
- To stop an application without removing it, try `ltc scale APP_NAME 0`.

### `ltc scale` 

`ltc scale APP_NAME NUM_INSTANCES` modifies the number of running instances of an application.

- **`--timeout=2m`** sets the maximum polling duration for scaling the app.

### `ltc update-routes`

`ltc update-routes APP_NAME PORT:ROUTE,PORT:ROUTE,...` allows you to update the routes associated with an application *after* it has been deployed.  The format is identical to the `--routes` option on `ltc create`. 

The set of routes passed into `ltc update-routes` will *override* the existing set of routes - these modification will start working shortly after the call to `update-routes`.

- **`--no-routes`** specifies that no routes be registered. 

## Building and Launching Droplets

**Note**: Buildpack support requires a Droplet Store, which is automatically configured when you run `ltc target`. You can validate that the Droplet store has been automatically detected by running `ltc target`. If Buildpack support is working correctly, you'll see two lines of output:
```
Target:		user@192.168.11.11.xip.io
Blob Store:	user@192.168.11.11.xip.io:8444
```

Lattice has the ability to build and run droplets generated by CF Buildpacks. Both the Build and Launch environments will track the [CF Stack](https://github.com/cloudfoundry/stacks).

### `ltc build-droplet`

`ltc build-droplet DROPLET_NAME http://github.com/buildpack-url` creates a droplet by running a Cloud Foundry Buildpack on an application directory.

- **`--path=.`** path to droplet source (file or folder)
- **`--env NAME[=VALUE]`** specifies environment variables. You can have multiple `--env` flags.  Passing an `--env` flag without explicitly setting the VALUE uses the current execution context to set the value.
- **`--timeout=2m`** sets the maximum polling duration for building the droplet.

  As an alternative to specifying the URL for a given Buildpack, we also support the following aliases: go, java, python, ruby, nodejs, php, binary, or staticfile.
  
  If the application directory contains a `.cfignore` file, `build-droplet` will not upload files that match the contents of the `.cfignore` file.

### `ltc launch-droplet`

`ltc launch-droplet APP_NAME DROPLET_NAME` launches a droplet as an app running on lattice

   `ltc launch-droplet` has the same options as [`ltc create`](/docs/ltc.md#ltc-create), except for `--run-as-root`.  Droplets are run as the user `vcap`.
   
   Finally, one can override the default start command by specifiying a start command after a `--` separator.  This can be followed by any arguments one wishes to pass to the app.  For example:

    ltc launch-droplet lattice-app lattice-app -- /lattice-app -quiet=true

### Managing Droplets


#### `ltc list-droplets`

`ltc list-droplets` lists the droplets available to launch on the droplet store.

#### `ltc remove-droplet`

`ltc remove-droplet DROPLET_NAME` removes a droplet from the droplet store. `remove-droplet` will error if an App based on the droplet is currently running in the cluster.

#### `ltc export-droplet`

`ltc export-droplet DROPLET_NAME` exports a droplet from the droplet store to disk.

#### `ltc import-droplet`

`ltc import-droplet DROPLET-NAME /path/droplet.tgz /path/result.json` imports a droplet from disk into the droplet store.



### `ltc submit-lrp`

`ltc submit-lrp /path/to/json` creates an application with the configuration specified in the JSON.  The syntax of the JSON can be found at the [Receptor API docs](https://github.com/cloudfoundry-incubator/receptor/blob/master/doc/lrps.md#describing-desiredlrps)

## Launching and Managing Tasks

### `ltc submit-task`

`ltc submit-task /path/to/json` creates a task with the configuration specified in the JSON.  The syntax of the task JSON can be found at the [Receptor API docs](https://github.com/cloudfoundry-incubator/receptor/blob/master/doc/tasks.md#describing-tasks)

### `ltc task`

`ltc task TASK_GUID` retrieves the assigned cell and task status, along with the result or failure if it's completed.

### `ltc delete-task`

`ltc delete-task TASK_GUID` deletes a completed task.  If a task has not compeleted yet, it will cancel and then delete the task.

## Streaming Logs

### `ltc logs`

`ltc logs APP_NAME` attaches to a log stream for a running application.  The logstream aggregates logs from *all* instances associated with an application.

## What's Running on Lattice?

### `ltc cells`

`ltc cells` lists each Lattice cell that is joined to the cluster.  It provides the available memory and disk capacity for each cell.

### `ltc list`

`ltc list` displays currently running applications and tasks not yet deleted on the targeted Lattice deployment.  For applications, this includes information on the number of requested and running instances, and routing information for accessing the application.  For tasks, the assigned cell, task status, result and/or failure reason are shown.

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
    Crash Count     0
    CPU Percentage  43.78
    Memory Usage    87M
    Disk Usage      163M
    --------------------------------------------------------------------------------

This indicates that instance 0 of the application has been `RUNNING` on `lattice-cell-01` since `2015-02-06 16:52:40 (PST)`.  The application can be reached at the indicated ip address (`192.168.11.11`).  This particular application included two EXPOSE directives, one for port  `7777` the other for port `9999`.  The `Ports` section of the report indicates the host-side ports that can be used to connect to the application at the requested container-side ports.  For example, to connect to `7777` one goes to `192.168.11.11:61001`.  To connect to `9999` one goes to `192.168.11.11:61002`.

- **`--summary`** summarizes the app instances section to one line per instance.
- **`--rate=1s`** refreshes the output at the specified time interval.

### `ltc visualize`

`ltc visualize` displays the *distribution* of application instances across the targetted Lattice deployment.  Each running application is rendered as a green dot.  Starting applications are rendered as yellow dots.

- **`--rate=1s`** refreshes the output at the specified time interval.
- **`--graphical`** uses the full terminal to display a graphical visualization.

## Is Lattice Working?

### `ltc test`

`ltc test` runs a minimal integration suite to ensure that a Lattice deploy is functioning correctly.

- **`--verbose`** verbose mode.  shows application output during test suite.
- **`--timeout=2m`** sets the wait time for Lattice to respond.

### `ltc debug-logs`

`ltc debug-logs` streams back logs from some of Lattice's key components.  This is useful for debugging situations where containers fail to get created/torn down.

- **`--raw`** prints the cluster logs with no styling.

