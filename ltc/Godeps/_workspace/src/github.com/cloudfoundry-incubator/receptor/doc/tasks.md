# Tasks

Diego can run one-off work in the form of Tasks.  When a Task is submitted Diego allocates resources on a Cell, runs the Task, and then reports on the Task's results.  Tasks are guaranteed to run at most once.

## Describing Tasks

When submitting a Task you `POST` a valid `TaskCreateRequest`.  The [API reference](api_tasks.md) includes the details of the request.  Here we simply describe what goes into a `TaskCreateRequest`:

```
{
    "task_guid": "some-guid",
    "domain": "some-domain",

    "rootfs": "docker:///docker-org/docker-image",
    "env": [
        {"name": "ENV_NAME_A", "value": "ENV_VALUE_A"},
        {"name": "ENV_NAME_B", "value": "ENV_VALUE_B"}
    ],

    "cpu_weight": 57,
    "disk_mb": 1024,
    "memory_mb": 128,
    "privileged": true,

    "action":  ACTION (see below),

    "result_file": "/path/to/return",
    "completion_callback_url": "http://optional/callback/url",

    "log_guid": "some-log-guid",
    "log_source": "some-log-source",

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

#### Task Identifiers

#### `task_guid` [required]

It is up to the consumer of Diego to provide a *globally unique* `task_guid`.  To subsequently fetch the Task you refer to it by its `task_guid`.

- It is an error to attempt to create a Task whose `task_guid` matches that of an existing Task.
- The `task_guid` must only include the characters `a-z`, `A-Z`, `0-9`, `_` and `-`.
- The `task_guid` must not be empty

#### `domain` [required]

The consumer of Diego may organize their Tasks into groupings called Domains.  These are purely organizational (e.g. for enabling multiple consumers to use Diego without colliding) and have no implications on the Task's placement or lifecycle.  It is possible to fetch all Tasks in a given Domain.

- It is an error to provide an empty `domain`.

#### Container Contents and Environment

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
"rootfs": "docker:///docker-org/docker-image#docker-tag"
```

To pull the image from a different registry than the default (Docker Hub), specify it as the host in the URI string, e.g.:

```
"rootfs": "docker://index.myregistry.gov/docker-org/docker-image#docker-tag"
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

#### `action` [required]

Encodes the action to run when running the Task.  For more details see [actions](actions.md)

#### Task Completion and Output

When the `action` on a Task terminates the Task is marked as `COMPLETED`.

#### `result_file` [optional]

When a Task completes succesfully Diego can fetch and return the contents of a file in the container.  This is made available in the `result` field of the `TaskResponse` (see [below](#retrieving-tasks)).

To do this, set `result_file` to a valid absolute path in the container.

- Diego only returns the first 10KB of the `result_file`.  If you need to communicate back larger datasets, consider using an `UploadAction` to upload the result file to a blob store.

#### `completion_callback_url` [optional]

Consumers of Diego have two options to learn that a Task has `COMPLETED`: they can either poll the action or register a callback.

If a `completion_callback_url` is provided Diego will `POST` to the provided URL as soon as the Task completes.  The body of the `POST` will include the `TaskResponse` (see [below](#retrieving-tasks)).

- Any response from the callback (be it success or failure) will resolve the Task (removing it from Diego).
- However, if the callback responds with `503` or `504` Diego will immediately retry the callback up to 3 times.  If the `503/504` status persists Diego will try again after a period of time (typically within ~30 seconds).
- If the callback times out or a connection cannot be established, Diego will try again after a period of time (typically within ~30 seconds).
- Diego will eventually (after ~2 minutes) give up on the Task if the callback does not respond succesfully.

#### Networking
By default network access for any container is limited but some tasks might need specific network access and that can be setup using `egress_rules` field.

Rules are evaluated in reverse order of their position, i.e., the last one takes precedence.

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

Diego uses [doppler](https://github.com/cloudfoundry/loggregator) to emit logs generated by container processes to the user.

#### `log_guid` [optional]

`log_guid` controls the doppler guid associated with logs coming from Task processes.  One typically sets the `log_guid` to the `task_guid` though this is not strictly necessary.

#### `log_source` [optional]

`log_source` is an identifier emitted with each log line.  Individual `RunAction`s can override the `log_source`.  This allows a consumer of the log stream to distinguish between the logs of different processes.

#### Attaching Arbitrary Metadata

#### `annotation` [optional]

Diego allows arbitrary annotations to be attached to a Task.  The annotation must not exceed 10 kilobytes in size.

## Retrieving Tasks

To learn that a Task is completed you must either register a `completion_callback_url` or periodically poll the API to fetch the Task in question.  In both cases, you will receive an object that includes **all the fields on the `TaskCreateRequest`** and the following additional fields:

```
{
    ... all TaskCreateRequest fields...

    "state": "RUNNING",

    "cell_id": "cell-identifier",

    "failed": true/false,
    "failure_reason": "why it failed",
    "result": "the contents of result_file",
}
```

Let's describe each of these fields in turn.

#### `state`

Tasks travel through a series of state transitions throughout their lifecycle.  These are described in [The Task Lifecycle](#the-task-lifecycle) below.

`state` will be a string and one of `INVALID`, `PENDING`, `CLAIMED`, `RUNNING`, `COMPLETED`, `RESOLVING`.

#### `cell_id`

Once claimed, a Task will include the ID of the Diego cell it is running on.

#### `failed`

Once a Task enters the `COMPLETED` state, `failed` will be a boolean indicating whether the Task completed succesfully or unsuccesfully.

#### `failure_reason`

If `failed` is `true`, `failure_reason` will be a short string indicating why the Task failed.  Sometimes, in the case of a `RunAction` that has failed this will simply read (e.g.) `exit status 1`.  To debug the Task you will need to fetch the logs from doppler.

#### `result`

If `result_file` was specified and the Task has completed succesfully, `result` will include the first 10KB of the `result_file`.

## The Task lifecycle

Tasks in Diego undergo a simple lifecycle encoded in the Tasks's state:

- When first created a Task enters the `PENDING` state.
- When succesfully allocated to a Diego Cell the Task will enter the `CLAIMED` state.  At this point the Task's `cell_id` will be populated.
- When the Cell begins to create the container and run the Task action, the Task enters the `RUNNING` state.
- When the Task completes, the Cell annotates the `TaskResponse` with `failed`, `failure_reason`, and `result`, and puts the Task in the `COMPLETED` state.

At this point it is up to the consumer of Diego to acknowledge and resolve the completed Task.  This can either be done via a completion callback (described [above](#completion_callback_url-optional)) or by [deleting](api_tasks.md#resolving-completed-tasks) the Task.  When the Task is being resolved it first enters the `RESOLVING` state and is ultimately removed from Diego.

Diego will automatically reap Tasks that remain unresolved after 2 minutes.

> The `RESOLVING` state exists to ensure that the `completion_callback_url` is initially called at most once per Task.

> There are a variety of timeouts associated with the `PENDING` and `CLAIMED` states.  It is possible for a Task to jump directly from `PENDING` or `CLAIMED` to `COMPLETED` (and `failed`) if any of these timeouts expire.  If you would like to impose a time limit on how long the Task is allowed to run you can use a `TimeoutAction`.

## Cancelling Tasks

Diego supports cancelling inflight tasks.  More documentation on this is available [here](api_tasks.md#cancelling-inflight-tasks).

[back](README.md)
