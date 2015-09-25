# Container Runtime Environment

Diego provides several collections of environment variables to processes running in containers.  These include the following layers, in order:

- Environment variables baked into the Docker image (for Docker image based containers only)
- Container-level environment variables (defined on the [`Task`](tasks.md#env) and [`LRP`](lrps.md#env) objects)
- Configuration environment variables (LRPs only, described below)
- Process-level environment variables (defined on the [`RunAction`](actions.md#runaction) object)

## Configuration-level Environment Variables

LRPs have additional runtime properties that are exposed to processes in the container via environment variables. These are as follows:

#### Instance Identifiers

- `INSTANCE_INDEX` is an integer denoting the index of the `ActualLRP`.  This will be in the range `0, 1, 2, ...N-1` where `N` is the number of instances on the `DesiredLRP`.
- `INSTANCE_GUID` is a unique identifier associated with an `ActualLRP`.  This will change every time a container is created.
- `CF_INSTANCE_INDEX` and `CF_INSTANCE_GUID` are aliases of these environment variables.

#### Networking Information

These environment variables are provided *only* if the operator deploying Diego has enabled `--exportNetworkEnvVars` on the Cell executor.

- `CF_INSTANCE_IP` provides the IP of the host running the container.  This is the IP used to address the container from the outside.
- `CF_INSTANCE_PORT` contains the host-side port corresponding to the *first* port in the `DesiredLRP` [`ports`](lrps.md#ports) array.
- `CF_INSTANCE_ADDR` is identical to `$CF_INSTANCE_IP:$CF_INSTANCE_PORT`.
- `CF_INSTANCE_PORTS` contains a list of JSON objects of the form `[{"external":60413,"internal":8080},{"external":60414,"internal":2222}]`. The internal ports are the container-side ports specified in the `DesiredLRP` [`ports`](lrps.md#ports) array, and the external ports are the corresponding host-side ports.

[back](README.md)
