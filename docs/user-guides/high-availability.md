The section provides information about HA schema for log collection. This schema of application can take high load of
logs in Cloud, process and send its to Graylog without loss of log messages.
The key idea of the HA schema is that there are daemon set and deployment of fluentbit pods that are configured to
send/receive data from each other.
Fluentbit pods deployed by daemon set are forwarders, these pods collect logs from each node and send unprocessed logs
to fluentbit aggregators (pods deployed as deployment) which have high resources to filter and send logs to Graylog.

To enable deploy of HA logs collection section `fluentbit.aggregator` must be filled.
The example of HA schema configuration section is below:

```yaml
fluentbit:
  install: true
  aggregator:
    install: true
    replicas: 2
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 2Gi
    graylogHost: x.x.x.x
    graylogPort: 12201
    graylogProtocol: tcp
    volume:
      bind: true
      storageClassName: "storage-class"
      storageSize: 200Mi
```

**Note:** only `fluentbit` supports HA schema, `fluentd` does not.
