This document describes how you can customize a data pipeline
`Input -> filter 1 -> ... -> filter N -> Output` in the logging agents (FluentD, Fluent-bit).

# Table of Contents

* [Table of Contents](#table-of-contents)
* [FluentD](#fluentd)
  * [Input customization](#input-customization)
    * [Customization of the out-of-box configuration](#customization-of-the-out-of-box-configuration)
    * [Custom input configuration](#custom-input-configuration)
  * [Filters customization](#filters-customization)
    * [Customization of the out-of-box configuration](#customization-of-the-out-of-box-configuration-1)
    * [Append fields to every log message](#append-fields-to-every-log-message)
    * [Custom filter configuration](#custom-filter-configuration)
  * [Output customization](#output-customization)
    * [Customization of the out-of-box configuration](#customization-of-the-out-of-box-configuration-2)
    * [Custom output configuration](#custom-output-configuration)
  * [Scenarios](#scenarios)
    * [Send a specific type of logs to custom output](#send-a-specific-type-of-logs-to-custom-output)
* [FluentBit](#fluentbit)
  * [Input customization](#input-customization-1)
    * [Customization of the out-of-box configuration](#customization-of-the-out-of-box-configuration-3)
    * [Custom input configuration](#custom-input-configuration-1)
  * [Filters customization](#filters-customization-1)
    * [Customization of the out-of-box configuration](#customization-of-the-out-of-box-configuration-4)
    * [Append fields to every log message](#append-fields-to-every-log-message-1)
    * [Custom filter configuration](#custom-filter-configuration-1)
  * [Output customization](#output-customization-1)
    * [Customization of the out-of-box configuration](#customization-of-the-out-of-box-configuration-5)
    * [Custom output configuration](#custom-output-configuration-1)
  * [Scenarios](#scenarios-1)
    * [Send a specific type of logs to custom output](#send-a-specific-type-of-logs-to-custom-output-1)
* [FluentBit Aggregator](#fluentbit-aggregator)

# FluentD

## Input customization

### Customization of the out-of-box configuration

Parameters that affect out-of-box input configuration:

* `containerRuntimeType`
* `fluentd.systemLogging`
* `fluentd.excludePath`
* `fluentd.systemLogType`

You can find the full list of FluentD parameters in
[the installation guide](../installation.md#fluentd).

### Custom input configuration

You can add your own custom part of the input pipeline configuration by using `fluentd.customInputConf`.
Example of the custom input configuration:

```yaml
fluentd:
  install: true
  #...
  customInputConf: |-
    <source>
      custom_input_configuration
    </source>
```

## Filters customization

### Customization of the out-of-box configuration

Parameters that affect out-of-box filters configuration:

* `containerRuntimeType`
* `fluentd.cloudEventsReaderFormat`
* `fluentd.billCycleConf`
* `fluentd.multilineFirstLineRegexp`

You can find the full list of FluentD parameters in
[the installation guide](../installation.md#fluentd).

### Append fields to every log message

You can add your own custom fields (labels) to every log messages, for example, to identify the source of sent logs
on the Graylog server later.

You can use `fluentd.extraFields` parameter to add custom key-value pairs to every log message processed by
the FluentD agent. Example:

```yaml
fluentd:
  install: true
  #...
  extraFields:
    foo_key: foo_value
    bar_key: bar_value
```

**Warning:** This filter is based on `record_transformer` plugin, so it can override existing fields if their keys are
identical.

This filter works after all other filters except the custom filter.

### Custom filter configuration

You can add your own custom part of the filtering pipeline configuration by using `fluentd.customFilterConf`.
Example of the custom filtration configuration:

```yaml
fluentd:
  install: true
  #...
  customFilterConf: |-
    <filter raw.kubernetes.var.log.**>
      custom_filter_configuration
    </filter>
```

## Output customization

### Customization of the out-of-box configuration

Parameters that affect out-of-box output configuration:

* `fluentd.graylogOutput`
* `fluentd.fileStorage`
* `fluentd.totalLimitSize`
* `fluentd.tls`

You can find the full list of FluentD parameters in
[the installation guide](../installation.md#fluentd).

### Custom output configuration

You can add your own custom part of the output pipeline configuration by using `fluentd.customOutputConf`.
Example of the custom output configuration:

```yaml
fluentd:
  install: true
  #...
  customOutputConf: |-
    <store ignore_error>
      custom_output_configuration
    </store>
```

## Scenarios

### Send a specific type of logs to custom output

**Objective**:

Some of your pods have logs that you want to send in your custom output. For example, access logs that are
represented like this:

<!-- markdownlint-disable line-length -->
```text
2024-05-20T09:57:38.406 CEF:0|Qubership|access-control|5.13.0|IMPORT_RULES|ABAC configuration was imported|6|suser=Some User
```
<!-- markdownlint-enable line-length -->

You can define these type of logs by including `access-control` marker in log (or stricter rule).

**Configuration**:

1. Add your custom filter that will rewrite tag in parameter `fluentd.customFilterConf`
   Documentation of `rewrite_tag_filter` link is [here](https://docs.fluentd.org/output/rewrite_tag_filter).
   For example:

   ```yaml
   fluentd:
     customFilterConf: |-
       <match parsed.kubernetes.var.log.your-app**>
         @type rewrite_tag_filter
         <rule>
           key log
           pattern /access-control*/
           tag my-tag.var.log.access-pods.log
         </rule>
       </match>
   ```

   The filter in the example checks if the logs from `your-app*` pods fits regular expression. If true, a new tag
   `my-tag.var.log.access-pods.log` applied to the log and emitted again.

   **Note:** Rewriting tag increases time of processing log, because log is sent to the beginning of pipeline and each
   filter applies to log again.

2. Configure custom output in Fluent Bit. There are different custom output integration described in documentation.

Now we can send logs with a new tag `my-tag.var.log.access-pods.log`. All we need to do is to set parameter
`fluentd.customOutputConf`.

For instance, you need to send logs with the new tag to Splunk:

```yaml
  fluentd:
    customOutputConf: |-
      <match my-tag.var.log.access-pods.log>
        @type copy
        @log_level fatal

        @type splunk_hec
        protocol http
        insecure_ssl false
        hec_host <splunk_host>
        hec_port <splunk_port>
        hec_token <splunk_token>
      </match>
```

In case if you need to send these logs both to Splunk and Graylog:

```yaml
fluentd:
  customOutputConf: |-
     <match my-tag.var.log.access-pods.log>
       @type copy

       <store ignore_error>
         @type gelf
         host <graylog_host>
         port <graylog_port>
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
    </match>
  ```

  **Note:** These are just examples, not recommended configurations.

# FluentBit

## Input customization

### Customization of the out-of-box configuration

Parameters that affect out-of-box input configuration:

* `containerRuntimeType`
* `fluentbit.systemLogging`
* `fluentbit.systemLogType`

You can find the full list of FluentBit parameters in
[the installation guide](../installation.md#fluentbit).

### Custom input configuration

You can add your own custom part of the input pipeline configuration by using `fluentbit.customInputConf`.
Example of the custom input configuration:

```yaml
fluentbit:
  install: true
  #...
  customInputConf: |-
    [INPUT]
      Name   <name>
```

## Filters customization

### Customization of the out-of-box configuration

Parameters that affect out-of-box filters configuration:

* `containerRuntimeType`
* `fluentbit.billCycleConf`
* `fluentbit.multilineFirstLineRegexp`
* `fluentbit.multilineOtherLinesRegexp`

You can find the full list of FluentBit parameters in
[the installation guide](../installation.md#fluentbit).

### Append fields to every log message

You can add your own custom fields (labels) to every log messages, for example, to identify the source of sent logs
on the Graylog server later.

You can use `fluentbit.extraFields` parameter to add custom key-value pairs to every log message
processed by the FluentBit agent. Example:

```yaml
fluentbit:
  install: true
  #...
  extraFields:
    foo_key: foo_value
    bar_key: bar_value
```

This filter works after all other filters except the custom filter. The filter is based on `record_modifier` plugin.

### Custom filter configuration

You can add your own custom part of the filtering pipeline configuration by using `fluentbit.customFilterConf`.
Example of the custom filtration configuration:

```yaml
fluentbit:
  install: true
  #...
  customFilterConf: |-
    [FILTER]
      Name record_modifier
      Match *
      Record testField fluent-bit
```

## Output customization

### Customization of the out-of-box configuration

Parameters that affect out-of-box output configuration:

* `fluentbit.graylogOutput`
* `fluentbit.graylogHost`
* `fluentbit.graylogPort`
* `fluentbit.graylogProtocol`
* `fluentbit.totalLimitSize`
* `fluentbit.tls`

You can find the full list of FluentBit parameters in
[the installation guide](../installation.md#fluentbit).

### Custom output configuration

You can add your own custom part of the output pipeline configuration by using `fluentbit.customOutputConf`.
Example of the custom output configuration:

```yaml
fluentbit:
  install: true
  #...
  customOutputConf: |-
    [OUTPUT]
      Name null
      Match fluent.*
```

## Scenarios

### Send a specific type of logs to custom output

**Objective**:

Some of your pods have logs that you want to send in your custom output. For example, access logs that are
represented like this:

<!-- markdownlint-disable line-length -->
```text
2024-05-20T09:57:38.406 CEF:0|Qubership|access-control|5.13.0|IMPORT_RULES|ABAC configuration was imported|6|suser=Some User
```
<!-- markdownlint-enable line-length -->

You can define these type of logs by including `access-control` marker in log (or stricter rule).

**Configuration**:

1. Add your custom filter that will rewrite tag in parameter `fluentbit.customFilterConf`
   Documentation of `rewrite_tag` filter link is [here](https://docs.fluentbit.io/manual/pipeline/filters/rewrite-tag).
   For example:

   ```yaml
   fluentbit:
     customFilterConf: |-
       [FILTER]
           Name                  rewrite_tag
           Match                 parsed.raw.kubernetes.var.log.pods.your-app*
           Rule                  $log /access-control*/ my-tag.var.log.access-pods.log true
           Emitter_name          re_my_audit
           Emitter_Storage.type  filesystem
           Emitter_Mem_Buf_Limit 10M
   ```

   The filter in the example checks if the logs from `your-app*` pods fits regular expression. If true, the log is
   copied with a new tag `my-tag.var.log.access-pods.log` and sent to the beginning of Fluent Bit pipeline to be
   processed (about Fluent Bit pipeline you can read
   [here](https://docs.fluentbit.io/manual/concepts/data-pipeline/router). In brief, each log message has a tag that is
   needed to decide should any filter be applied to log and to route log to output according to tag). If the log should
   be sent both in Graylog and your custom output, the flag `true` set in `Rule` parameter.

   **Note:** Rewriting tag increases time of processing log, because log is sent to the beginning of pipeline and each
   filter applies to log again.

2. Configure custom output in Fluent Bit. There are different custom output integration described in documentation.

  Now we can send logs with a new tag `my-tag.var.log.access-pods.log`. All we need to do is to set parameter
  `fluentbit.customOutputConf`.

  For instance, you need to send logs with the new tag to Splunk:

  ```yaml
  fluentbit:
    customOutputConf: |-
      [OUTPUT]
          Name         splunk
          Match        my-tag.var.log.access-pods.log
          Host         <splunk_host>
          Port         <splunk_port>
          Splunk_Token <splunk_token>
          TLS          On
          TLS.Verify   Off
  ```

  **Note:** These are just examples, not recommended configurations.

# FluentBit Aggregator

The way to customize the data pipeline in the FluentBit in Aggregator mode is almost the same as for
the standard [FluentBit](#fluentbit) except the following differences:

* The aggregator configuration have fixed input config
* The aggregator configuration doesn't have `billCycleConf`
* You have to set parameters into the different `fluentbit.aggregator` section instead of `fluentbit`
