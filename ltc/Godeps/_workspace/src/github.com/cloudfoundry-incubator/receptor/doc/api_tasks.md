# Tasks API Reference

This reference does not cover the JSON payload supplied to each endpoint.  That is documented in detail in [Understanding Tasks](tasks.md).

We recommend using the [Receptor http client](https://github.com/cloudfoundry-incubator/receptor) to communicate with Diego's API.  The methods on the client are self-explanatory.

## Creating Tasks

To create a Task submit a valid [`TaskCreateRequest`](tasks.md#describing-tasks) via:

```
POST /v1/tasks
```

## Fetching Tasks

### Fetching all Tasks

To fetch *all* Tasks:

```
GET /v1/tasks
```

This returns an array of [`TaskResponse`](tasks.md#retreiving-tasks) objects


### Fetching Tasks by Domain

To fetch all Tasks in a given [`domain`](tasks.md#domain-required):

```
GET /v1/tasks?domain=domain-name
```

This returns an array of [`TaskResponse`](tasks.md#retreiving-tasks) objects

### Fetching a Specific Task

To fetch a Task by [`task_guid`](tasks.md#task_guid-required):

```
GET /v1/tasks/:task_guid
```

This returns a single [`TaskResponse`](tasks.md#retreiving-tasks) object or `404` if none is found.

## Resolving Completed Tasks

When a Task enters the `COMPLETED` state (see [The Task Lifecycle](tasks.md#the-task-lifecycle) for details) you are responsible for resolving it.

This is done via:

```
DELETE /v1/tasks/:task_guid
```

You can only resolve a task in the `COMPLETED` state.  Anything else is an error.

## Cancelling Inflight Tasks

Tasks in the `PENDING` and `RUNNING` states (see [The Task Lifecycle](tasks.md#the-task-lifecycle) for details) can be cancelled. This results in a Task in the `COMPLETED` state with `failed = true`.

To cancel a task:

```
POST /v1/tasks/:task_guid/cancel
```
Note, you must resolve the cancelled Task once it enters the `COMPLETED` state.

[back](README.md)
