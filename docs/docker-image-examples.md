# Docker Image Examples

Lattice supports docker images. Lattice containers currently use ephemeral disk and therefore are only suitable for running workloads that can have persistent disk reset on restart events, which works fine for many development and testing scenarios. After a container referencing a docker image with well-known ports is running, use the `ltc status APPNAME` command to see the port mapping. See the [troubleshooting](/docs/troubleshooting.md) docs if necessary.

### Redis
`ltc create redis redis -r`

### Rabbit
`ltc create rabbit rabbitmq -r`

### MySQL
`ltc create mysql mysql -r -e MYSQL_ROOT_PASSWORD=somesecret`

### Postgres
`ltc create postgres postgres -r --no-monitor -e POSTGRES_PASSWORD=somesecret`

note: if you do not use --no-monitor, you continually see the log message:
`LOG:  incomplete startup packet`

### Mongo
`ltc create mongo mongo -r -e LC_ALL=C -- /entrypoint.sh mongod --smallfiles`

### Neo4J
`ltc create neo4j tpires/neo4j -r -m 512`

It may take some time for the container to load such that initial the ltc create operation times out, you can monitor the status with `ltc status neo4j -r 2s` or `ltc debug-logs | veritas chug` to monitor the progress.

### Ubuntu
`ltc create ubuntu library/ubuntu -- nc -l 8080`

### Nginx
`ltc start nginx library/nginx -p 80 -r`

### Spring Cloud Config Server
`ltc create --run-as-root configserver springcloud/configserver`

If you want to register with Eureka add an env var:
`ltc create --env EUREKA_SERVICE_URL=http://eureka.192.168.11.11.xip.io --run-as-root configserver springcloud/configserver`

### Spring Cloud Eureka Server
`ltc create --run-as-root eureka springcloud/eureka`

### Spring Cloud Clients
`ltc create --run-as-root --env CONFIG_SERVER_URL=http://configserver.192.168.11.11.xip.io --env EUREKA_SERVICE_URL=http://eureka.192.168.11.11.xip.io myapp mygroup/myapp`
