This document provides information about integration options for logging agents (FluentD or FluentBit) as producers
for Kafka and Graylog as a consumer for Kafka.

# Table of Content

* [Table of Content](#table-of-content)
* [Kafka](#kafka)
  * [Before you begin](#before-you-begin)
  * [Configuring FluentD output to Kafka](#configuring-fluentd-output-to-kafka)
    * [Plaintext](#plaintext)
    * [SASL plaintext](#sasl-plaintext)
    * [SASL over SSL](#sasl-over-ssl)
  * [Configuring FluentBit output to Kafka](#configuring-fluentbit-output-to-kafka)
    * [Plaintext](#plaintext-1)
    * [SASL plaintext](#sasl-plaintext-1)
    * [SASL over SSL](#sasl-over-ssl-1)
  * [Configuring Graylog input from Kafka](#configuring-graylog-input-from-kafka)
    * [SASL plaintext](#sasl-plaintext-2)
    * [SASL over SSL](#sasl-over-ssl-2)
    * [Example of Kafka consumer parameters](#example-of-kafka-consumer-parameters)
* [Links](#links)

# Kafka

Apache Kafka is an open-source distributed event streaming platform used by thousands of companies for high-performance
data pipelines, streaming analytics, data integration, and mission-critical applications.

Kafka can aggregate data from producers (e.g. logging agents) and give this data to consumers.

## Before you begin

You should find out a several things before you begin:

* Address of Kafka brokers that you will use to send logs (host and port)
* Kafka security protocol (e.g. `plaintext`, `ssl`, `sasl_plaintext`, `sasl_ssl`) and auth options:
  * SASL settings (enabled or disabled, credentials, mechanism)
  * SSL settings (enabled or disabled, certificates)

Usually Kafka can have one of the following secure protocols:

1. `plaintext`: connection to the Kafka doesn't require any additional security parameters
2. `ssl`: connection to the Kafka requires managing SSL certificates
3. `sasl_plaintext`: connection to the Kafka requires setting SASL mechanism (e.g. GSSAPI, PLAIN, SCRAM-SHA-256,
   SCRAM-SHA-512, OAUTHBEARER), username and password
4. `sasl_ssl`: requires parameters from both `ssl` and `sasl_plaintext` options

We recommend creating topics in the Kafka manually, but FluentD Kafka plugin, FluentBit and Graylog can create
new topics in the Kafka by themselves in some cases.

Please notice that if you connect to Kafka in the Cloud **from outside the Cloud** (e.g. if your Graylog has been
installed on the virtual machine), you may need to take additional steps to open access to Kafka.

## Configuring FluentD output to Kafka

FluentD uses [fluent-plugin-kafka](https://github.com/fluent/fluent-plugin-kafka) to send logs to Kafka brokers.

Now FluentD cannot configure Kafka output as a separate output. So it needs to use a `custom output`
to configure it.

You can find information about Kafka output configuration
[in FluentD documentation](https://docs.fluentd.org/output/kafka) and more information with examples
[in plugin README](https://github.com/fluent/fluent-plugin-kafka#readme).

**Warning!**: FluentD applies `custom output` before default output to Graylog. Also, FluentD stops processing
logs after reach the first output. So it means that if you specify output in the `custom output` section, FluentD won't
send logs to the default Graylog output.

**NOTE:** Remember that all examples of configuration in this document are **just examples,
not recommended parameters**, so the responsibility for tuning and adapting the configuration for a specific environment
lies with the users themselves.

### Plaintext

To configure the simplest Kafka output with protocol `plaintext` in FluentD you can add the following `custom output`
config in the FluentD configuration during deploy:

```yaml
fluentd:
  customOutputConf: |-
    <match parsed.**>
      @type kafka2
      brokers [broker-1-host]:[broker-1-port],[broker-2-host]:[broker-2-port]
      use_event_time true
      <buffer [your-topic]>
        @type file
        path /var/log/td-agent/buffer/td
        flush_interval 3s
      </buffer>
      <format>
        @type json
      </format>
      topic_key [your-topic]
      default_topic [your-topic]
      required_acks -1
      compression_codec gzip

      sasl_over_ssl false
    </match>
```

**NOTE:** The `<buffer>` section can be helpful for stability of FluentD work, but it may take up a lot of disk space in
a short time, so please use this feature with caution.

### SASL plaintext

If your Kafka uses `sasl_plaintext` as a security protocol, you should add parameters `username`, `password` and
`scram_mechanism` if you use SCRAM (we use `SCRAM-SHA-512` in this example):

```yaml
fluentd:
  customOutputConf: |-
    <match parsed.**>
      @type kafka2
      brokers [broker-1-host]:[broker-1-port],[broker-2-host]:[broker-2-port]
      use_event_time true
      <buffer [your-topic]>
        @type file
        path /var/log/td-agent/buffer/td
        flush_interval 3s
      </buffer>
      <format>
        @type json
      </format>
      topic_key [your-topic]
      default_topic [your-topic]
      required_acks -1
      compression_codec gzip

      sasl_over_ssl false
      username [your-username]
      password [your-password]
      scram_mechanism sha512
    </match>
```

**WARNING:** parameter `scram_mechanism` must be written in **lower case** and in a different way than
the same parameters can be set in the other tools, for example: `sha512`.

**NOTE:** The `<buffer>` section can be helpful for stability of FluentD work, but it may take up a lot of disk space in
a short time, so please use this feature with caution.

### SASL over SSL

If your Kafka uses `sasl_ssl` as a security protocol, you have to follow the following steps:

1. Create a Kubernetes Secret in the namespace where is placed your FluentD and add the Kafka certificate to it.
   For example:

   ```yaml
   kind: Secret
   apiVersion: v1
   type: Opaque
   metadata:
     name: kafka-fluentd-ca
   data:
     ca.crt: ...base64 encoded certificate...
   ```

2. Add parameters that allow using of the certificate from the Secret by the FluentD pod and then start
   the Rolling update job. Example of deploy parameters:

   ```yaml
   fluentd:
     tls:
       enabled: false
       noVerify: true
       ...
       ca:
         secretName: kafka-fluentd-ca
         secretKey: ca.crt
   ```

   You need only `tls.ca` section to mount the certificate to FluentD pods.

   **IMPORTANT:** notice that parameters `tls.enabled`, `noVerify` and other affect only the out-of-box GELF Graylog
   output, so you can leave `tls.enabled: false` if you don't want to use TLS between FluentD and Graylog.
   Furthermore, if you wrongly set `tls.enabled` to `true` without changing the same option in the Graylog server,
   you would probably face an error in FluentD logs.

3. Switch parameter `sasl_over_ssl` to `true` and add parameter `ssl_ca_cert /fluentd/tls/ca.crt` to the custom
   output configuration:

   ```yaml
   fluentd:
     customOutputConf: |-
       <match parsed.**>
         @type kafka2
         brokers [broker-1-host]:[broker-1-port],[broker-2-host]:[broker-2-port]
         use_event_time true
         <buffer [your-topic]>
           @type file
           path /var/log/td-agent/buffer/td
           flush_interval 3s
         </buffer>
         <format>
           @type json
         </format>
         topic_key [your-topic]
         default_topic [your-topic]
         required_acks -1
         compression_codec gzip

         sasl_over_ssl true
         username [your-username]
         password [your-password]
         scram_mechanism sha512
         sasl_over_ssl true
         ssl_ca_cert /fluentd/tls/ca.crt
       </match>
   ```

   **NOTE:** The `<buffer>` section can be helpful for stability of FluentD work, but it may take up a lot of disk space
   in a short time, so please use this feature with caution.

If you want to use [cert-manager](../user-guides/tls.md#about-cert-manager), you can use the subsection `generateCerts`
for such purposes. The only requirement for this way is to set **the same `clusterIssuerName` for Kafka and for logging
VM** application to get CA certificates that will match with each other.

Example of configuration with the cert-manager:

```yaml
fluentd:
  tls:
    enabled: false
    noVerify: true
    ...
    generateCerts:
      enabled: true
      clusterIssuerName: "<kafka's_issuer>" # Required
      ...
```

## Configuring FluentBit output to Kafka

FluentBit has a built-in plugin to send collected logs to Kafka. So you only need to configure it.

Configuration parameters can be found
[in FluentBit documentation](https://docs.fluentbit.io/manual/pipeline/outputs/kafka).

**NOTE:** Remember that all examples of configuration in this document are **just examples,
not recommended parameters**, so the responsibility for tuning and adapting the configuration for a specific environment
lies with the users themselves.

### Plaintext

To configure the simplest Kafka output with protocol `plaintext` in FluentBit you can add the following `custom output`
config in the FluentBit configuration during deploy:

```yaml
fluentbit:
  customOutputConf: |-
    [OUTPUT]
        Name                                        kafka
        Match                                       parsed.*
        Brokers                                     [broker-1-host]:[broker-1-port],[broker-2-host]:[broker-2-port]
        Topics                                      [your-topic]
        Format                                      json
        rdkafka.enable.ssl.certificate.verification false
        rdkafka.security.protocol                   plaintext
```

### SASL plaintext

If your Kafka uses `sasl_plaintext` as a security protocol, you should add parameters `username`, `password` and
`scram_mechanism` if you use SCRAM (we use `SCRAM-SHA-512` in this example):

```yaml
fluentbit:
  customOutputConf: |-
    [OUTPUT]
        Name                                        kafka
        Match                                       parsed.*
        Brokers                                     [broker-1-host]:[broker-1-port],[broker-2-host]:[broker-2-port]
        Topics                                      [your-topic]
        Format                                      json
        rdkafka.enable.ssl.certificate.verification false
        rdkafka.security.protocol                   sasl_plaintext
        rdkafka.sasl.username                       [your-username]
        rdkafka.sasl.password                       [your-password]
        rdkafka.sasl.mechanism                      SCRAM-SHA-512
```

### SASL over SSL

If your Kafka uses `sasl_ssl` as a security protocol, you have to follow the following steps:

1. Create a Kubernetes Secret in the namespace where is placed your FluentBit and add the Kafka certificate to it.
   For example:

   ```yaml
   kind: Secret
   apiVersion: v1
   type: Opaque
   metadata:
     name: kafka-fluentbit-ca
   data:
     ca.crt: ...base64 encoded certificate...
   ```

2. Add parameters that allow using of the certificate from the Secret by the FluentBit pod and then start
   the Rolling update job. Example of deploy parameters:

   ```yaml
   fluentbit:
     tls:
       enabled: false
       verify: false
       ...
       ca:
         secretName: kafka-fluentbit-ca
         secretKey: ca.crt
   ```

   You need only `tls.ca` section to mount the certificate to FluentBit pods.

   **IMPORTANT:** notice that parameters `tls.enabled`, `verify` and other affect only the out-of-box GELF Graylog
   output, so you can leave `tls.enabled: false` if you don't want to use TLS between FluentBit and Graylog.
   Furthermore, if you wrongly set `tls.enabled` to `true` without changing the same option in the Graylog server,
   you would probably face an error in FluentBit logs.

3. Change parameters `rdkafka.security.protocol` to `sasl_ssl`, `rdkafka.enable.ssl.certificate.verification` to `true`
   and add parameter `rdkafka.ssl.ca.location /fluent-bit/tls/ca.crt` to the custom output configuration:

   ```yaml
   fluentbit:
     customOutputConf: |-
       [OUTPUT]
           Name                                        kafka
           Match                                       parsed.*
           Brokers                                     [broker-1-host]:[broker-1-port],[broker-2-host]:[broker-2-port]
           Topics                                      [your-topic]
           Format                                      json
           rdkafka.enable.ssl.certificate.verification true
           rdkafka.security.protocol                   sasl_ssl
           rdkafka.sasl.username                       [your-username]
           rdkafka.sasl.password                       [your-password]
           rdkafka.sasl.mechanism                      SCRAM-SHA-512
           rdkafka.ssl.ca.location                     /fluent-bit/tls/ca.crt
   ```

If you want to use [cert-manager](../user-guides/tls.md#about-cert-manager), you can use the subsection `generateCerts`
for such purposes. The only requirement for this way is to set **the same `clusterIssuerName` for Kafka and for logging
VM** application to get CA certificates that will match with each other.

Example of configuration with the cert-manager:

```yaml
fluentbit:
  tls:
    enabled: false
    verify: false
    ...
    generateCerts:
      enabled: true
      clusterIssuerName: "<kafka's_issuer>" # Required
      ...
```

## Configuring Graylog input from Kafka

You can configure your Graylog server as a Kafka consumer to collect logs from its topics. You need to configure
a specific Graylog `input` to do it.

We assume that you've already installed a Graylog server. Then you can create a new input through Graylog UI by
the following steps:

1. Log in to the Graylog UI
2. Go to `System -> Inputs`
3. Select a type of the new input to the left of the `Launch new input`. If you're not confident in the format
   of your logs, select `Raw/Plaintext Kafka`, but if you're sure then choose either `CEF Kafka`, `GELF Kafka` or
   `Syslog Kafka`.
4. Click on the `Launch new input` button

Configuration between different types of Kafka inputs is almost the same, but we'll assume that you selected
`Raw/Plaintext Kafka`.

In the opened window you should enter configuration of your connection to the Kafka. Let's look at the most
important parameters in the simplest configuration of the Kafka input:

* `Global`: check the box if you don't want to specify a particular Elastic node to store the data
* `Title`: enter a title for your input
* `Legacy mode`: uncheck the box, if only you aren't using ZooKeeper
* `Bootstrap Servers`: enter your Kafka brokers in the comma separated list format:
  `[broker-1-host]:[broker-1-port],[broker-2-host]:[broker-2-port]`
* `Topic filter regex`: enter a regex for topic(-s) that you are going to use
* `Custom Kafka properties`: you have to enter some parameters here if your Kafka is using any security protocol except
  `plaintext`

Let's look at the `Custom Kafka properties` parameters in more details. You can put any Kafka consumer parameters here.
You can find an [example of Kafka consumer parameters](#example-of-kafka-consumer-parameters) below directly from
Graylog logs.

There are a several examples for the `Custom Kafka properties` fields below.

**NOTE:** Remember that all examples of configuration in this document are **just examples,
not recommended parameters**, so the responsibility for tuning and adapting the configuration for a specific environment
lies with the users themselves.

### SASL plaintext

If the Kafka uses `sasl_plaintext` security protocol with `SCRAM-SHA-512` as a SASL mechanism:

```yaml
sasl.mechanism=SCRAM-SHA-512
sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required username="[your-username]" password="[your-password]";
security.protocol=SASL_PLAINTEXT
```

**IMPORTANT:** please notice that the `sasl.jaas.config` parameter has a specific syntax. Also, for some reasons it's a
required parameter and its location is matter. For example, during tests the Graylog input failed if this parameter is
removed or placed in the end of the parameters list.

### SASL over SSL

If your Kafka uses `sasl_ssl` as a security protocol, you have to follow the following steps:

1. Create a Kubernetes Secret in the namespace where is placed your Graylog and add the Kafka certificate to it or add
   the certificate in the Graylog CA Secret that already exists. Graylog can add all certificates from the chosen Secret
   to its keystore. We will assume that you have the Secret with the Kafka certificate in it.
   Example of the Secret:

   ```yaml
   kind: Secret
   apiVersion: v1
   type: Opaque
   metadata:
     name: graylog-ca
   data:
     ca-1.crt: ...base64 encoded certificate...
     ca-2.crt: ...base64 encoded certificate...
     ...
     kafka-ca.crt: ...base64 encoded certificate...
   ```

   We will use the Graylog keystore to add CA certificate to Kafka input settings.

2. Add parameters that allow including of the Kafka certificate from the Graylog CA Secret to the Graylog pod.
   We have to use `tls.http` section **instead of** `tls.input` to add any certificate to the Graylog keystore.
   Example of deploy parameters:

   ```yaml
   graylog:
     tls:
       http:
         enabled: false
         cacerts: graylog-ca
   ```

   All you need is to configure parameter `tls.http.cacerts`.

   **IMPORTANT:** notice that parameter `tls.http.enabled` affects any HTTP connections to the Graylog server
   **except inputs**, so you can leave `tls.enabled: false` if you don't want to use TLS for all Graylog connections.
   You also don't need to enable `tls.input`, because it affects only out-of-box GELF Graylog input managed by the
   logging-operator.

3. Configure Kafka input as you'd do it for the `SASL plaintext` protocol, but add `ssl.keystore.location` parameter
   with `=/usr/share/graylog/data/ssl/cacerts.jks` value, set password from your keystore to the `ssl.keystore.password`
   parameter (`changeit` by default) and change `security.protocol` value to `SASL_SSL`:

   ```yaml
   sasl.mechanism=SCRAM-SHA-512
   sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required username="[your-username]" password="[your-password]";
   security.protocol=SASL_SSL
   ssl.keystore.location=/usr/share/graylog/data/ssl/cacerts.jks
   ssl.keystore.password=changeit
   ```

If you want to use [cert-manager](../user-guides/tls.md#about-cert-manager), you can use the subsection `generateCerts`
for such purposes. The only requirement for this way is to set **the same `clusterIssuerName` for Kafka and for logging
VM** application to get CA certificates that will match with each other.

Example of configuration with the cert-manager:

```yaml
graylog:
  tls:
    http:
      enabled: false
      ...
      generateCerts:
        enabled: true
        clusterIssuerName: "<kafka's_issuer>" # Required
        ...
```

### Example of Kafka consumer parameters

These parameters appear in Graylog logs when you create a Kafka input:

```yaml
    allow.auto.create.topics = true
    auto.commit.interval.ms = 1000
    auto.offset.reset = latest
    bootstrap.servers = []
    check.crcs = true
    client.dns.lookup = use_all_dns_ips
    client.id = gl2-44a226cb-651c122a84432b72c19abb66-0
    client.rack = 
    connections.max.idle.ms = 540000
    default.api.timeout.ms = 60000
    enable.auto.commit = true
    exclude.internal.topics = true
    fetch.max.bytes = 52428800
    fetch.max.wait.ms = 100
    fetch.min.bytes = 5
    group.id = graylog2
    group.instance.id = null
    heartbeat.interval.ms = 3000
    interceptor.classes = []
    internal.leave.group.on.close = true
    internal.throw.on.fetch.stable.offset.unsupported = false
    isolation.level = read_uncommitted
    key.deserializer = class org.apache.kafka.common.serialization.ByteArrayDeserializer
    max.partition.fetch.bytes = 1048576
    max.poll.interval.ms = 300000
    max.poll.records = 500
    metadata.max.age.ms = 300000
    metric.reporters = []
    metrics.num.samples = 2
    metrics.recording.level = INFO
    metrics.sample.window.ms = 30000
    partition.assignment.strategy = [class org.apache.kafka.clients.consumer.RangeAssignor]
    receive.buffer.bytes = 65536
    reconnect.backoff.max.ms = 1000
    reconnect.backoff.ms = 50
    request.timeout.ms = 30000
    retry.backoff.ms = 100
    sasl.client.callback.handler.class = null
    sasl.jaas.config = null
    sasl.kerberos.kinit.cmd = /usr/bin/kinit
    sasl.kerberos.min.time.before.relogin = 60000
    sasl.kerberos.service.name = null
    sasl.kerberos.ticket.renew.jitter = 0.05
    sasl.kerberos.ticket.renew.window.factor = 0.8
    sasl.login.callback.handler.class = null
    sasl.login.class = null
    sasl.login.refresh.buffer.seconds = 300
    sasl.login.refresh.min.period.seconds = 60
    sasl.login.refresh.window.factor = 0.8
    sasl.login.refresh.window.jitter = 0.05
    sasl.mechanism = GSSAPI
    security.protocol = PLAINTEXT
    security.providers = null
    send.buffer.bytes = 131072
    session.timeout.ms = 10000
    socket.connection.setup.timeout.max.ms = 127000
    socket.connection.setup.timeout.ms = 10000
    ssl.cipher.suites = null
    ssl.enabled.protocols = [TLSv1.2]
    ssl.endpoint.identification.algorithm = https
    ssl.engine.factory.class = null
    ssl.key.password = null
    ssl.keymanager.algorithm = SunX509
    ssl.keystore.certificate.chain = null
    ssl.keystore.key = null
    ssl.keystore.location = null
    ssl.keystore.password = null
    ssl.keystore.type = JKS
    ssl.protocol = TLSv1.2
    ssl.provider = null
    ssl.secure.random.implementation = null
    ssl.trustmanager.algorithm = PKIX
    ssl.truststore.certificates = null
    ssl.truststore.location = null
    ssl.truststore.password = null
    ssl.truststore.type = JKS
    value.deserializer = class org.apache.kafka.common.serialization.ByteArrayDeserializer
```

# Links

* [fluent-plugin-kafka](https://github.com/fluent/fluent-plugin-kafka)
* [FluentD documentation about Kafka output](https://docs.fluentd.org/output/kafka)
* [FluentBit documentation about Kafka output](https://docs.fluentbit.io/manual/pipeline/outputs/kafka)
