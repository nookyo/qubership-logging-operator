This document provides information about integration options with Splunk logging agents such as
FluentD or FluentBit.

# Table of Content

* [Table of Content](#table-of-content)
* [Splunk](#splunk)
  * [Before you begin](#before-you-begin)
  * [Integration FluentD with Splunk](#integration-fluentd-with-splunk)
  * [Integration FluentBit with Splunk](#integration-fluentbit-with-splunk)
* [Links](#links)

# Splunk

Splunk is a software that helps capture, index and correlate real-time data in a searchable repository, from which
it can generate graphs, reports, alerts, dashboards and visualizations.

For logging cases, Splunk can collect logs from a lot of sources, and store and analyze collected logs.

## Before you begin

* Address of Splunk that you will use to send logs (host and port)
* Token to auth in Splunk and that will use to send logs
  * Splunk CloudPlatform - [Use authentication tokens](https://docs.splunk.com/Documentation/SplunkCloud/latest/Security/UseAuthTokens)
  * Splunk Enterprise - [Use authentication tokens](https://docs.splunk.com/Documentation/Splunk/9.0.4/Security/UseAuthTokens)

## Integration FluentD with Splunk

**Support since:** logging-fluentd 7.7.0

Now FluentD cannot configure Splunk as a separate output. So to configure it need to use a `custom output`
config.

> **Warning!**
>
> FluentD apply `custom output` before default output to Graylog. Also FluentD stop process logs after reach a first
> output. So it means that if you specify output in the `custom output` section, FluentD won't send logs to
> default Graylog output.

**NOTE:** Remember that all examples of configuration in this document are **just examples,
not recommended parameters**, so the responsibility for tuning and adapting the configuration for a specific environment
lies with the users themselves.

To configure the Splunk HEC output plugin in FluentD you can add the following `custom output` config in FluentD config:

```yaml
fluentd:
  customOutputConf: |-
     <match parsed.**>
       @type copy
       @log_level fatal

       @type splunk_hec
       protocol http
       insecure_ssl false
       hec_host <splunk_host>
       hec_port <splunk_port>
       hec_token <splunk_token>
       index main
       flush_interval 10s
    </match>
```

To send logs in two outputs, to Splunk and to Graylog, you can use the following config:

```yaml
fluentd:
  customOutputConf: |-
     <match parsed.**>
       @type copy
       @log_level fatal

       <store ignore_error>
         @type gelf
         host x.x.x.x
         port 12201
         protocol tcp
         retry_wait 1s
         <buffer>
           flush_interval 30s
           retry_max_interval 64
           chunk_limit_size 2m
           queue_limit_length 160
           flush_thread_count 32
           retry_forever false
         </buffer>
       </store>

       @type splunk_hec
       protocol http
       insecure_ssl false
       hec_host <splunk_host>
       hec_port <splunk_port>
       hec_token <splunk_token>
       index main
       flush_interval 10s
    </match>
```

## Integration FluentBit with Splunk

**NOTE:** Remember that all examples of configuration in this document are **just examples,
not recommended parameters**, so the responsibility for tuning and adapting the configuration for a specific environment
lies with the users themselves.

FluentBit has a built-in plugin to send collected logs to Splunk. So you need only configure it.

To configure the Splunk output plugin in FluentBit you can add the following `custom output` config
in the FluentBit config:

```yaml
fluentbit:
  customOutputConf: |-
    [OUTPUT]
        Name         splunk
        Match        parsed.*
        Host         <splunk_host>
        Port         <splunk_port>
        Splunk_Token <splunk_token>
        TLS          On
        TLS.Verify   Off
```

# Links

* [Use authentication tokens](https://docs.splunk.com/Documentation/SplunkCloud/latest/Security/UseAuthTokens)
* [Use authentication tokens](https://docs.splunk.com/Documentation/Splunk/9.0.4/Security/UseAuthTokens)
* [FluentD Splunk plugin: fluent-plugin-splunk-hec](https://github.com/splunk/fluent-plugin-splunk-hec)
* [FluentBit output configuration for Splunk](https://docs.fluentbit.io/manual/pipeline/outputs/splunk/)
