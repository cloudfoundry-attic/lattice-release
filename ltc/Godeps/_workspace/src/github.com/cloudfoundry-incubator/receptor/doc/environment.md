# Container Runtime Environment

Diego sets up a hierarchy of environment variables available to processes running in containers.  This hierarchy includes the following layers in order:

- Environment variables baked into the Docker image (for Docker image based containers only)
- Container-level environment variables (defined on the [`Task`](tasks.md#env) and [`LRP`](lrps.md#env) objects)
- Configuration environment variables (LRPs only, described below)
- Process-level environment variables (defined on the [`RunAction`](actions.md#runaction) object)

## Configuration-level Environment Variables

LRPs have additional runtime characteristics that are exposed to processes in the container via environment variables.

These are:

#### Instance Identifiers

- `INSTANCE_INDEX` is an integer denoting the index of the `ActualLRP`.  This will be in the range `0, 1, 2, ...N-1` where `N` is the number of instances on the `DesiredLRP`
- `INSTANCE_GUID` is a unique identifier associated with an `ActualLRP`.  This will change every time a container is created.

#### Networking Information

These environment variables are *only* provided if the operator deploying Diego has enabled `--exportNetworkEnvVars` on the Cell executor.

- `CF_INSTANCE_IP` provides the IP of the host running the container.  This is the IP used to address the container from the outside.
- `CF_INSTANCE_PORT` the host-side port corresponding to the *first* desired port in the `DesiredLRP` [`ports`](lrps.md#ports) array.
- `CF_INSTANCE_ADDR` identical to `$CF_INSTANCE_IP:$CF_INSTANCE_PORT`
- `CF_INSTANCE_PORTS` a list of the form `61012:8080,61013:5000`.  The comma delimited entries are pairs of `host-side-port:container-side-port`.  The container-side ports map onto the `DesiredLRP` [`ports`](lrps.md#ports) array.

[back](README.md)
