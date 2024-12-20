This document provides information about integration options with any external systems using syslog
format (for example any SIEM systems).

# Table of Content

* [Table of Content](#table-of-content)
* [Syslog](#syslog)
  * [Before you begin](#before-you-begin)
  * [Common limits/restrictions](#common-limitsrestrictions)
    * [Messages length limits](#messages-length-limits)
    * [Multiline messages](#multiline-messages)
  * [Send logs using syslog in FluentD](#send-logs-using-syslog-in-fluentd)
    * [Filter logs and send only filtered to Syslog for FluentD](#filter-logs-and-send-only-filtered-to-syslog-for-fluentd)
  * [Send logs using syslog in FluentBit](#send-logs-using-syslog-in-fluentbit)
    * [Filter logs and send only filtered to Syslog for FluentBit](#filter-logs-and-send-only-filtered-to-syslog-for-fluentbit)

# Syslog

Syslog has been around for a number of decades and provides a protocol used for transporting event messages
between computer systems and software applications. The Syslog protocol utilizes a layered architecture,
which allows the use of any number of transport protocols for the transmission of Syslog messages.
It also provides a message format that allows vendor-specific extensions to be provided in a structured way.

## Before you begin

* You should know the address and port of the Syslog server or SIEM system
* You should know the protocol of the Syslog server (TCP or UDP)
* You should know other Syslog settings that require your Syslog server
* You should verify that Fluent has access from Kubernetes to the Syslog server

## Common limits/restrictions

This section describes various limits and restrictions applied to both logging agents, FluentD and FluentBit.

### Messages length limits

In the case of using the `syslog` output plugins, you need to remember that it has a configurable message limit.

It can be configured using the parameter, like the following:

* FluentD

    ```yaml
    @type remote_syslog
    host <address>
    port <port>
    protocol <protocol>
    ...
    packet_size 1024  # default value is 1024 (in bytes)
    ```

* FluentBit

    ```yaml
    [OUTPUT]
      name                 syslog
      match                *
      host                 <address>
      port                 <port>
      mode                 <protocol>
      ...
      syslog_format        rfc3164
      syslog_maxsize       1024
    ```

    **Note:** If no value is provided, the default size is set depending of the protocol version specified by `syslog_format`.
    `rfc3164` sets max size to `1024` bytes, `rfc5424` sets the size to `2048` bytes.

This parameter sets the maximum size of messages that can be sent using the syslog protocol.

Moreover, the syslog protocol has a prefix in the message that also requires some amount of symbols.
Usually, the syslog message should be generated as:

```bash
<priority>id timestamp hostname tag content
```

**Note:** Format can be different because there are some RFC and versions of `syslog` protocol.
It's just an example, use official RFCs to read more about `syslog` format.

Usually, the prefix can take up approximately 300 bytes from the whole message limit. So using the default limit
in `1024 bytes` your message can contain approximately `700 bytes`.

FluentD is written in Ruby which means that by default it uses `ASCII-8bit` encoding that uses 1 byte to encode
Latin symbols. All other symbols from `UTF-8` will be encoded by 2 bytes.

Based on this information you can calculate which limit should be set to set the maximum allowed per message size.

### Multiline messages

The `syslog` protocol doesn't allow send in messages control symbols. It described in RFC5424
[https://datatracker.ietf.org/doc/html/rfc5424#section-8.2](https://datatracker.ietf.org/doc/html/rfc5424#section-8.2).

It means that symbols like `\n` (new line) or `\t` tabulation also don't allow and multiline messages can't be sent
as it.

By default, all `syslog` output plugins will send each new line from the multiline log as a separate message.
For example, the java stacktrace:

```bash
Exception in thread "main" java.lang.NullPointerException: Fictitious NullPointerException
  at ClassName.methodName1(ClassName.java:lineNumber)
  at ClassName.methodName2(ClassName.java:lineNumber)
  at ClassName.methodName3(ClassName.java:lineNumber)
  at ClassName.main(ClassName.java:lineNumber)
```

will be send as `5` separated messages, like:

```bash
<14>1 2021-07-12T14:37:35.569848Z myhost myapp 1234 ID98 [namespace="mynamespace"] Exception in thread "main" java.lang.NullPointerException: Fictitious NullPointerException
<14>1 2021-07-12T14:37:35.569848Z myhost myapp 1234 ID98 [namespace="mynamespace"] at ClassName.methodName1(ClassName.java:lineNumber)
<14>1 2021-07-12T14:37:35.569848Z myhost myapp 1234 ID98 [namespace="mynamespace"] at ClassName.methodName2(ClassName.java:lineNumber)
<14>1 2021-07-12T14:37:35.569848Z myhost myapp 1234 ID98 [namespace="mynamespace"] at ClassName.methodName3(ClassName.java:lineNumber)
<14>1 2021-07-12T14:37:35.569848Z myhost myapp 1234 ID98 [namespace="mynamespace"] at ClassName.main(ClassName.java:lineNumber)
```

Some external systems that can receive logs using the `syslog` format provide a function to build from separated messages
one multiline message. Please refer to the documentation of your Logging system.

## Send logs using syslog in FluentD

Now the docker image provided by Platform includes two syslog plugins for FluentD:

* [fluent-plugin-remote_syslog](https://github.com/fluent-plugins-nursery/fluent-plugin-remote_syslog)
* [fluent-plugin-syslog_rfc5424](https://github.com/cloudfoundry/fluent-plugin-syslog_rfc5424)

The full list of parameters for the plugins you run is in the official documentation.

> **Warning!**
>
> FluentD apply `custom output` before default output to Graylog. Also FluentD stop process logs after reach a first
> output. So it means that if you specify output in the `custom output` section, FluentD won't send logs to
> default Graylog output.

To configure the Syslog output plugin in FluentD you can add the following `custom output` config in FluentD config:

* for `fluent-plugin-remote_syslog`

    ```yaml
    fluentd:
      customOutputConf: |-
        <match parsed.**>
           @type copy
           @log_level fatal

           @type remote_syslog
           host <address>
           port <port>
           protocol <protocol>
           program fluentd
           hostname ${tag[1]}
           flush_interval 10s

            <format>
              @type single_value
              message_key message
            </format>
        </match>
    ```

* for `fluent-plugin-syslog_rfc5424`

    ```yaml
    fluentd:
      customOutputConf: |-
        <match parsed.**>
           @type copy
           @log_level fatal

           @type syslog_rfc5424
           host <address>
           port <port>
           transport <protocol>
           flush_interval 10s
        </match>
    ```

To send logs in two outputs, to Syslog server and Graylog, you can use the following config:

* for `fluent-plugin-remote_syslog`

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

           @type remote_syslog
           host <address>
           port <port>
           protocol <protocol>
           program fluentd
           hostname ${tag[1]}
           flush_interval 10s

            <format>
              @type single_value
              message_key message
            </format>
        </match>
    ```

* for `fluent-plugin-syslog_rfc5424`

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

           @type syslog_rfc5424
           host <address>
           port <port>
           transport <protocol>
           flush_interval 10s
        </match>
    ```

### Filter logs and send only filtered to Syslog for FluentD

In some cases, for example, to send logs in the SIEM system, you may want to filter them and send only
some specific scope of logs to the system.

In FluentD all logs routing based on tags and their name patterns. So you need to set a special value
of tags using `match` or `filter` sections.

```yaml
fluentd:
  customFilterConf: |-
    <match parsed.kubernetes.**>
      @type rewrite_tag_filter
      <rule>
        key namespace_name
        pattern ^idp$
        tag siem.${tag}
      </rule>
    </match>
```

or

```yaml
fluentd:
  customFilterConf: |-
    <match parsed.kubernetes.**>
      @type rewrite_tag_filter
      <rule>
        key 'container_name'
        pattern ^(access-control|identity-provider)$
        tag siem.${tag}
      </rule>
    </match>
```

And next, in the `custom output` section you need to use `match` by early set tag to route marked logs
to the necessary output. For example:

* for `fluent-plugin-remote_syslog`

    ```yaml
    fluentd:
      customOutputConf: |-
        <match siem.**>
           @type copy
           @log_level fatal

           @type remote_syslog
           host <address>
           port <port>
           protocol <protocol>
           program fluentd
           hostname ${tag[1]}
           flush_interval 10s

            <format>
              @type single_value
              message_key message
            </format>
        </match>
    ```

* for `fluent-plugin-syslog_rfc5424`

    ```yaml
    fluentd:
      customOutputConf: |-
        <match siem.**>
           @type copy
           @log_level fatal

           @type syslog_rfc5424
           host <address>
           port <port>
           transport <protocol>
           flush_interval 10s
        </match>
    ```

## Send logs using syslog in FluentBit

FluentBit has one built-in plugin to send logs in syslog format:

* [Syslog](https://docs.fluentbit.io/manual/pipeline/outputs/syslog)

The full list of parameters for the plugin you run is in the official documentation.

To configure the Syslog output plugin in FluentBit you can add the following `custom output` config
in the FluentBit config:

```yaml
fluentbit:
  customOutputConf: |-
    [OUTPUT]
        name                 syslog
        match                parsed.*
        host                 <address>
        port                 <port>
        mode                 <protocol>
```

### Filter logs and send only filtered to Syslog for FluentBit

In some cases, for example, to send logs in the SIEM system, you may want to filter them and send only
some specific scope of logs to the system.

In FluentBit all logs routing based on tags and their name patterns. So you need to set a special value
of tags using `match` or `filter` sections.

```yaml
fluentbit:
  customFilterConf: |-
    [FILTER]
        Name                  rewrite_tag
        Match                 parsed.kubernetes.**
        Rule                  $message /Invalid username or password|Successful Login|Successful Logout|Failed to look up user based on cookie/ parsed.var.log.audit.audit.log false
        Emitter_name          re_grafana
        Emitter_Storage.type  filesystem
        Emitter_Mem_Buf_Limit 10M
```

And next, in the `custom output` section you need to use `match` by early set tag to route marked logs
to the necessary output. For example:

```yaml
fluentbit:
  customOutputConf: |-
    [OUTPUT]
        name                 syslog
        match                re_grafana
        host                 <address>
        port                 <port>
        mode                 <protocol>
```
