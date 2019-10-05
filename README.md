# linuxmetricstostatsd

A small utility that gathers system metrics and sends them to a StatsD server. 

Uses these two great libraries:

- https://github.com/shirou/gopsutil
- https://github.com/cactus/go-statsd-client

## Installation

```bash
docker run -d --name linuxmetricstostatsd --net=host -v /proc:/rootfs/proc \
-v /sys:/rootfs/sys -v /etc:/rootfs/etc -v /var/:/rootfs/var \
-e "HOST_PROC=/rootfs/proc" -e "HOST_VAR=/rootfs/var" \
-e "HOST_SYS=/rootfs/sys" -e "HOST_ETC=/rootfs/etc" \
docker.io/dgkanatsios/linuxmetricstostatsd:0.1.0
```

StatsD host/port as well as metrics polling interval can be configured during app execution.
