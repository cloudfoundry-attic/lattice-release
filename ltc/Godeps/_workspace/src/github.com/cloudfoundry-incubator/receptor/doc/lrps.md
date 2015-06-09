# LRPs: Long Running Processes

Diego can distribute and monitor multiple instances of a Long Running Process (LRP).  These instances are distributed across Diego Cells and restarted automatically if they crash or disappear.  The instances are identical (though each instance is given a unique index (in the range `0, 1, ...N-1`) and a unique instance guid).

LRPs are described by providing Diego with a `DesiredLRP`.  The `DesiredLRP` can be thought of as a manifest that describes how an LRP should be executed and monitored.

The instances that end up running on Diego cells are referred to as `ActualLRP`s.  The `ActualLRP`s contain information about the state of the instance and about the host Cell the instance is running on.

When describing a property common to both `DesiredLRP`s and `ActualLRP`s (e.g. the `process_guid`) we may refer to both notions collectively simply as LRPs.

Diego is constantly monitoring and reconciling desired state and actual state.  As such it is important to ensure that the desired state is up-to-date and accurate.  This is covered in detail in the section below on [Freshness](#freshness).

First, let's discuss DesiredLRPs.

## Describing DesiredLRPs

When desiring an LRP you `POST` a valid `DesiredLRPCreateRequest`.  The [API reference](api_lrps.md) includes the details of the request.  Here we simply describe what goes into a `DesiredLRPCreateRequest`:

```
{
    "process_guid": "some-guid",
    "domain": "some-domain",

    "instances": 17,

    "rootfs": "VALID-ROOTFS",

    "env": [
        {"name": "ENV_NAME_A", "value": "ENV_VALUE_A"},
        {"name": "ENV_NAME_B", "value": "ENV_VALUE_B"}
    ],

    "cpu_weight": 57,
    "disk_mb": 1024,
    "memory_mb": 128,
    "privileged": true,

    "setup": ACTION,
    "action":  ACTION,
    "monitor": ACTION,
    "start_timeout": N seconds,

    "ports": [8080, 5050],
    "routes": {
        "cf-router": [
            {
                "hostnames": ["a.example.com", "b.example.com"],
                "port": 8080
            }, {
                "hostnames": ["c.example.com"],
                "port": 5050
            }
        ],
        "router-key": "any opaque json payload"
    }

    "log_guid": "some-log-guid",
    "log_source": "some-log-source",
    "metrics_guid": "some-metrics-guid",
    "annotation": "arbitrary metadata",

    "egress_rules": [
        {
            "protocol": "tcp",
            "destinations": ["0.0.0.0/0"],
            "port_range": {
                "start": 1,
                "end": 1024
            }
       }
   ]
}
```

Let's describe each of these fields in turn.

#### LRP Identifiers

#### `process_guid` [required]

It is up to the consumer of Diego to provide a *globally unique* `process_guid`.  To subsequently fetch the DesiredLRP and its ActualLRP you refer to it by its `process_guid`.

- The `process_guid` must only include the characters `a-z`, `A-Z`, `0-9`, `_` and `-`.
- The `process_guid` must not be empty
- If you attempt to create a DesiredLRP with a `process_guid` that matches that of an existing DesiredLRP, Diego will attempt to update the existing DesiredLRP.  This is subject to the rules described in [updating DesiredLRPs](#updating-desiredlrps) below.

#### `domain` [required]

The consumer of Diego may organize their LRPs into groupings called Domains.  These are purely organizational (e.g. for enabling multiple consumers to use Diego without colliding) and have no implications on the ActualLRP's placement or lifecycle.  It is possible to fetch all LRPs in a given Domain.

- It is an error to provide an empty `domain`.

#### LRP Placement

In the future Diego will support the notion of Placement Pools via arbitrary tags associated with Cells.

#### Instances

#### `instances` [required]

Diego can run and manage multiple instances (`ActualLRP`s) for each `DesiredLRP`.  `instances` specifies the number of desired instances and must not be less than zero.

#### Container Contents and Environment

#### `rootfs` [required]

The `rootfs` field specifies the root filesystem to mount into the container.  Diego can be configured with a set of *preloaded* RootFSes.  These are named root filesystems that are already on the Diego Cells.

Preloaded root filesystems look like:

```
"rootfs": "preloaded:ROOTFS-NAME"
```

Diego ships with a root filesystem:
```
"rootfs": "preloaded:cflinuxfs2"
```
these are built to work with the Cloud Foundry buildpacks.

It is possible to provide a custom root filesystem by specifying a Docker image for `rootfs`:

```
"rootfs": "docker:///docker-user/docker-image#docker-tag"
```

To pull the image from a different registry than the default (Docker Hub), specify it as the host in the URI string, e.g.:

```
"rootfs": "docker://index.myregistry.gov/docker-user/docker-image#docker-tag"
```

> You *must* specify the dockerimage `rootfs` uri as specified, including the leading `docker://`!

> [Lattice](https://github.com/pivotal-cf-experimental/lattice) does not ship with any preloaded root filesystems. You must specify a Docker image when using Lattice. You can mount the filesystem provided by diego-release by specifying `"rootfs": "docker:///cloudfoundry/cflinuxfs2"`.

#### `env` [optional]

Diego supports the notion of container-level environment variables.  All processes that run in the container will inherit these environment variables.

For more details on the environment variables provided to processes in the container, read [Container Runtime Environment](environment.md)

#### Container Limits

#### `cpu_weight` [optional]

To control the CPU shares provided to a container, set `cpu_weight`.  This must be a positive number in the range `1-100`.  The `cpu_weight` enforces a relative fair share of the CPU among containers.  It's best explained with examples.  Consider the following scenarios (we shall assume that each container is running a busy process that is attempting to consume as many CPU resources as possible):

- Two containers, with equal values of `cpu_weight`: both containers will receive equal shares of CPU time.
- Two containers, one with `cpu_weight=50` the other with `cpu_weight=100`: the later will get (roughly) 2/3 of the CPU time, the former 1/3.

#### `disk_mb` [optional]

A disk quota applied to the entire container.  Any data written on top of the RootFS counts against the Disk Quota.  Processes that attempt to exceed this limit will not be allowed to write to disk.

- `disk_mb` must be an integer >= 0
- If set to 0 no disk constraints are applied to the container
- The units are megabytes

#### `memory_mb` [optional]

A memory limit applied to the entire container.  If the aggregate memory consumption by all processs running in the container exceeds this value, the container will be destroyed.

- `memory_mb` must be an integer >= 0
- If set to 0 no memory constraints are applied to the container
- The units are megabytes

#### `privileged` [optional]

If false, Diego will create a container that is in a user namespace.  Processes that succesfully obtain escalated privileges (i.e. root access) will actually only be root within the user namespace and will not be able to maliciously modify the host VM.  If true, Diego creates a container with no user namespace -- escalating to root gives the user *real* root access.

#### Actions

When an LRP instance is instantiated, a container is created with the specified `rootfs` mounted.  Diego is responsible for performing any container setup necessary to successfully launch processes and monitor said processes.

#### `setup` [optional]

After creating a container, Diego will first run the action specified in the `setup` field.  This field is optional and is typically used to download files and run (short-lived) processes that configure the container.  For more details on the available actions see [actions](actions.md).

- If the `setup` action fails the `ActualLRP` is considered to have crashed and will be restarted

#### `action` [required]

After completing any `setup` action, Diego will launch the `action` action.  This `action` is intended to launch any long running processes.  For more details on the available actions see [actions](actions.md).

#### `monitor` [optional]

If provided, Diego will monitor the long running processes encoded in `action` by periodically invoking the `monitor` action.  If the `monitor` action returns succesfully (exit status code 0), the container is deemed "healthy", otherwise the container is deemed "unhealthy".  Monitoring is quite flexible in Diego and is outlined in more detail [below](#monitoring-health).

#### `start_timeout` [optional]

If provided, Diego will give the `action` action up to `start_timeout` seconds to become healthy before marking the LRP as failed.

#### Networking

Diego can open and expose arbitrary `ports` inside the container.  There are plans to generalize this support and make it possible to build custom service discovery solutions on top of Diego.  The API is likely to change in backward-incompatible ways as we work these requirements out.

By default network access for any container is limited but some LRPs might need specific network access and that can be setup using `egress_rules` field.  Rules are evaluated in reverse order of their position, i.e., the last one takes precedence.

> Lattice users: Lattice is intended to be a single-tenant cluster environment.  In Lattice there are no network access constraints on the containers so there is no need to specify `egress_rules`.

#### `ports` [optional]

`ports` is a list of ports to open in the container.  Processes running in the container can bind to these ports to receive incoming traffic.  These ports are only valid within the container namespace and an arbitrary host-side port is created when the container is created.  This host-side port is made available on the `ActualLRP`.

#### `routes` [optional]

`routes` is a map where the keys identify route providers and the values hold information for the providers to consume.  The information in the map must be valid JSON but is not proessed by Diego.  The total length of the routing information must not exceed 4096 bytes.

##### `cf-router` [optional]

The route provider `cf-router` is used by the Diego [route emitter](https://github.com/cloudfoundry-incubator/route-emitter) to automatically register routes to the container with the [router](https://github.com/cloudfoundry/gorouter).  The routing information is a list of objects that associate a container port with a list of fully qualified host names (e.g. `"foo.example.com"`).  Consumers that attempt to access one of the hostnames via the router will be connected to one of the ActualLRP instances that is currently running.

Example: `"cf-router": [{"port":8080, "hostnames":["foo.example.com"]}}]`

#### `egress_rules` [optional]
`egress_rules` are a list of egress firewall rules that are applied to a container running in Diego

##### `protocol` [required]
The protocol of the rule that can be one of the following `tcp`, `udp`,`icmp`, `all`.

##### `destinations` [required]
The destinations of the rule that is a list of either an IP Address (1.2.3.4) or an IP range (1.2.3.4-2.3.4.5) or a CIDR (1.2.3.4/5)

##### `ports` [optional]
A list of destination ports that are integers between 1 and 65535.
> `ports` or `port_range` must be provided for `tcp` and `udp`.
> It is an error when both are provided.

##### `port_range` [optional]
- `start` [required] the start of the range as an integer between 1 and 65535
- `end` [required] the end of the range as an integer between 1 and 65535

> `ports` or `port_range` must be provided for protocol `tcp` and `udp`.
> It is an error when both are provided.

##### `icmp_info` [optional]
- `type` [required] will be an integer between 0 and 255
- `code` [required] will be an integer

> `icmp_info` is required for protocol `icmp`.
> It is an error when provided for other protocols.

##### `log` [optional]
Enable logging of the rule
> `log` is optional for `tcp` and `all`.
> It is an error to provide `log` as true when protocol is `udp` or `icmp`.

> Define all rules with `log` enabled at the end of your `egress_rules` to guarantee logging.

##### Examples
***
`ALL`
```
{
    "protocol": "all",
    "destinations": ["1.2.3.4"],
    "log": true
}
```
***
`TCP`
```
{
    "protocol": "tcp",
    "destinations": ["1.2.3.4-2.3.4.5"],
    "ports": [80, 443],
    "log": true
}
```
***
`UDP`
```
{
    "protocol": "udp",
    "destinations": ["1.2.3.4/4"],
    "port_range": {
        "start": 8000,
        "end": 8085
    }
}
```
***
`ICMP`
```
{
    "protocol": "icmp",
    "destinations": ["1.2.3.4", "2.3.4.5/6"],
    "icmp_info": {
        "type": 1,
        "code": 40
    }
}
```
***
#### Logging

Diego uses [loggregator](https://github.com/cloudfoundry/loggregator) to emit logs generated by container processes to the user.

#### `log_guid` [optional]

`log_guid` controls the loggregator guid associated with logs coming from LRP processes.  One typically sets the `log_guid` to the `process_guid` though this is not strictly necessary.

#### `log_source` [optional]

`log_source` is an identifier emitted with each log line.  Individual `RunAction`s can override the `log_source`.  This allows a consumer of the log stream to distinguish between the logs of different processes.

#### `metrics_guid` [optional]

`metrics_guid` controls the loggregator guid associated with metrics coming from LRP processes.

#### Attaching Arbitrary Metadata

#### `annotation` [optional]

Diego allows arbitrary annotations to be attached to a DesiredLRP.  The annotation must not exceed 10 kilobytes in size.

## Updating DesiredLRPs

Only a subset of the DesiredLRP's fields may updated dynamically.  In particular, changes that require the process to be restarted are not allowed - instead, you should submit a new DesiredLRP and orchestrate the upgrade path from one LRP to the next.  This provides the consumer of Diego the flexibility to pick the most appropriate upgrade strategy (blue-green, etc...)

It is possible, however, to dynamically modify the number of instances, and the routes associated with the LRP.  Diego's API makes this explicit -- when updating a DesiredLRP you provide a `DesiredLRPUpdateRequest`:

```
{
    "instances": 17,
    "routes": {
        "cf-router": [
            {
                "hostnames": ["a.example.com", "b.example.com"],
                "port": 8080
            }, {
                "hostnames": ["c.example.com"],
                "port": 5050
            }
        ],
        "router-key": "any opaque json payload"
    },
    "annotation": "arbitrary metadata"
}
```

These may be provided simultaneously in one request, or independendantly over several requests.

### Monitoring Health

It is up to the consumer to tell Diego how to monitor an LRP instance.  If provided, Diego uses the `monitor` action to ascertain when an LRP is up.

Typically, an ActualLRP instance begins in an unhealthy state (`CLAIMED`).  At this point the `monitor` action is polled every 0.5 seconds.  Eventually the `monitor` action succeeds and the instance enters a healthy state (`RUNNING`).  At this point the `monitor` action is polled every 30 seconds.  If the `monitor` action subsequently fails, the ActualLRP is considered crashed.  Diego's consumer is free to define an arbitrary `monitor` action - a `monitor` action may check that a port is accepting connections, or that a URL returns a happy status code, or that a file is present in the container.  In fact, a single `monitor` action might be a composition of other actions that can monitor multiple processes running in the container.

Normally, the `action` action on the DesiredLRP does not exit.  It is possible, however, to launch and daemonize a process in Diego.  If the `action` action exits succesfully Diego assumes the process is a daemon and continues monitoring it with the `monitor` action.  If the `action` action fails (e.g. exit with non-zero status code for a `RunAction`) Diego assumes the ActualLRP has failed and schedules it to be restarted.

Finally, it is possible to opt out of monitoring.  If no `monitor` action is specified then the health of the ActualLRP is dependent on the `action` continuing to run indefinitely.  The ActualLRP is considered `RUNNING` as soon as the `action` action begins, and is considered to have failed if the `action` action ever exits.

> Note that Diego does not currently stream back logs for processes that daemonize.

### Fetching DesiredLRPs

Diego allows consumers to fetch DesiredLRPs -- the response object (`DesiredLRPResponse`) is identical to the `DesiredLRPCreateRequest` object described above.

When fetching DesiredLRPs one can fetch *all* DesiredLRPs in Diego, all DesiredLRPs of a given `domain`, and a specific DesiredLRP by `process_guid`.

The fact that a DesiredLRP is present in Diego does not mean that the corresponding ActualLRP instances are up and running.  Diego converges on the desired state and starting/stopping ActualLRPs may take time.  The presence of a DesiredLRP in Diego signifies the consumer's intent for Diego to run instances - not that those instances are currently running.  For that you must fetch the ActualLRPs.

## Fetching ActualLRPs

As outlined above, DesiredLRPs represent the consumer's intent for Diego to run instances.  To fetch instances, consumers must [fetch ActualLRPs](api_lrps.md#fetching-actuallrps).

When fetching ActualLRPs, one can fetch *all* ActualLRPs in Diego, all ActualLRPs of a given `domain`, all ActualLRPs for a given DesiredLRP by `process_guid`, and all ActualLRPs at a given *index* for a given `process_guid`.

In all cases, the consumer is given an array of `ActualLRPResponse`:

```
[
    {
        "process_guid": "some-process-guid",
        "instance_guid": "some-instnace-guid",
        "cell_id": "some-cell-id",
        "domain": "some-domain",
        "index": 15,
        "state": "UNCLAIMED", "CLAIMED", "RUNNING" or "CRASHED"

        "address": "10.10.11.11",
        "ports": [
            {"container_port": 8080, "host_port": 60001},
            {"container_port": 5000, "host_port": 60002},
        ],

        "placement_error": "insufficient resources",

        "since": 1234567
    },
    ...
]
```

Let's describe each of these fields in turn.

#### ActualLRP Identifiers

#### `process_guid`

The `process_guid` for this ActualLRP -- this is used to correlate ActualLRPs with DesiredLRPs.

#### `instance_guid`

An arbitrary identifier unique to this ActualLRP instance.

#### `cell_id`

The identifier of the Diego Cell running the ActualLRP instance.

#### `domain`

The `domain` associated with this ActualLRP's DesiredLRP.

#### `index`

The `index` of the ActualLRP - an integer between `0` and `N-1` where `N` is the desired number of instances.

#### `state`

The state of the ActualLRP.

When an ActualLRP is first created, it enters the `UNCLAIMED` state.

Once the ActualLRP is placed onto a Cell it enters the `CLAIMED` state.  During this time a container is being created and the various processes inside the container are being spun up.

When the `action` action begins running, Diego begins periodically running the `monitor` action.  As soon as the `monitor` action reports that the processes are healthy the ActualLRP will transition into the `RUNNING` state.

#### `placement_error`

When an ActualLRP cannot be placed because there are no resources to place it, the `placement_error` is populated with the reason.

> `placement_error` is only populated when the ActualLRP is in the `UNCLAIMED` state.

#### `since`

The last modified time of the ActualLRP represented as the number of nanoseconds elapsed since January 1, 1970 UTC.

#### Networking
#### `address`

`address` contains the externally accessible IP of the host running the container.

> `address` is only populated when the ActualLRP enters the `RUNNING` state.

#### `ports`

`ports` is an array containing mappings between the `container_port`s requested in the DesiredLRP and the `host_port`s associated with said `container_port`s.  In the example above to connect to the process bound to port `5000` inside the container, a request must be made to `10.10.11.11:60002`.

> `ports` is only populated when the ActualLRP enters the `RUNNING` state.

### Killing ActualLRPs

Diego supports killing the `ActualLRP`s for a given `process_guid` at a given `index`.  This is documented [here](api_lrps.md#killing-actuallrps).  Note that this does not change the *desired* state -- Diego will simply shut down the `ActualLRP` at the given `index` and will eventually converge on desired state by restarting the (now-missing) instance.  To permanently scale down a DesiredLRP you must update the `instances` field on the DesiredLRP.

## Freshness

Diego periodically compares desired state (the set of DesiredLRPs) to actual state (the set of ActualLRPs) and takes actions to keep the actual state in sync with the desired state.  This eventual consistency model is at the core of Diego's robustness.

In order to perform this responsibility safely, however, Diego must have some way of knowing that it's knowledge of the desired state is complete and up-to-date.  In particular, consider a scenario where Diego's database has crashed and must be repopulated.  In this context it is possible to enter a state where the actual state (the ActualLRPs) are known to Diego but the desired state (the DesiredLRPs) is not.  It would be catastrophic for Diego to attempt to converge by shutting down all actual state!

To circumvent this, it is up to the consumer of Diego to inform Diego that its knowledge of the desired state is up-to-date.  We refer to this as the "freshness" of the desired state.  Consumers explicitly mark desired state as *fresh* on a domain-by-domain basis.  Failing to do so will prevent Diego from taking actions to ensure eventual consistency (in particular, Diego will refuse to stop extra instances with no corresponding desired state).

To maintain freshness you perform a simple [PUT](api_domains.md#domains).  The consumer typically supplies a TTL and attempts to bump the freshness of the domain before the TTL expires (verifying along the way, of course, that the contents of Diego's DesiredLRP are up-to-date).

It is possible to opt out of this by updating the freshness with *no* TTL.  In this case the freshness will never expire and Diego will always perform all its eventual consistency operations.

> Note: only destructive operations performed during an eventual consistency convergence cycle are gated on freshness.  Diego will continue to start/stop instances when explicitly instructed to.

[back](README.md)
