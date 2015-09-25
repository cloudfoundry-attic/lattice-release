# Example LRPs

We provide some full examples of `DesiredLRPCreateRequest` JSON payloads for different types of LRPs below.

### Web server via `bash` and `nc`

This DesiredLRP runs an extremely simple HTTP server on port 8080 via `bash` and `nc`, in the preloaded `cflinuxfs2` as its distinguished unprivileged user, `vcap`:

```
{
  "process_guid": "my-nc-server-def456",
  "domain": "my-fresh-domain",
  "rootfs": "preloaded:cflinuxfs2",
  "instances": 1,
  "action": {
    "run": {
      "path": "bash",
      "args": [
        "-c",
        "set -e; mkfifo request; while true; do\n{\nread < request\n\necho -n -e \"HTTP/1.1 200 OK\\r\\nContent-Length: ${#INSTANCE_INDEX}\\r\\n\\r\\n${INSTANCE_INDEX}\"\n} | nc -l 0.0.0.0 $PORT > request;\ndone\n}"
      ],
      "env": [{"name":"PORT", "value":"8080"}],
      "user": "vcap"
    }
  },
  "ports": [
    8080
  ]
}
```

This example is based on an LRP used in Diego's [inigo integration test suite](https://github.com/cloudfoundry-incubator/inigo).

### Redis docker image

This DesiredLRP runs one instance of a Redis server via the [official Redis Docker image](https://hub.docker.com/_/redis/):

```
{
  "process_guid": "my-redis-server-abc123",
  "domain": "my-fresh-domain",
  "rootfs": "docker:///redis",
  "instances": 1,
  "action": {
    "run": {
      "path": "/entrypoint.sh",
      "args": [
        "redis-server"
      ],
      "dir": "/data",
      "user": "root"
    }
  },
  "ports": [
    6379
  ]
}
```

Many elements of the LRP specification were determined from the [redis 3.0 Dockerfile](https://github.com/docker-library/redis/blob/master/3.0/Dockerfile):

- The `rootfs` field itself refers to the official `redis` Docker image on Docker Hub.
- The RunAction takes its `path` and `args` values from the `ENTRYPOINT` and `CMD` directives in the Dockerfile, and its `dir` value from the `WORKDIR` directive. For the `user` value, we provide `root`, since this is the implicit user for a Docker image that does not specify a `USER` directive in its Dockerfile. 
- The `ports` field corresponds to the ports listed in the `EXPOSE` directives in the Dockerfile.

[back](README.md)
