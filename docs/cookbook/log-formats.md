This document describes log format agreements.

# Table of Content

* [Table of Content](#table-of-content)
* [Logs Collection in the Cloud](#logs-collection-in-the-cloud)
* [Log Format for Applications](#log-format-for-applications)
  * [Logs Traceability](#logs-traceability)
  * [Complex Format Example](#complex-format-example)
  * [JSON logs](#json-logs)
  * [Special Log Categories](#special-log-categories)
    * [Audit Logs](#audit-logs)
      * [JSON Audit logs](#json-audit-logs)
    * [System Logs](#system-logs)
* [Collecting Logs from Other Source](#collecting-logs-from-other-source)
* [Best Practices for Logging](#best-practices-for-logging)
  * [Log Amount and Content](#log-amount-and-content)
  * [Logging Levels](#logging-levels)
  * [Log level configuration](#log-level-configuration)

# Logs Collection in the Cloud

This section provides information on how the logs are collected in the Cloud.

The high-level flow of logs collection is as follows:

1. The application in the docker container writes logs to stdout.
2. The docker log driver, by default a JSON file, writes it to OpenShift node file system.
3. The Log collector, Fluentd, reads this file.
4. The Fluentd processes the log file content and sends the parsed logs to Graylog.

There is another flow for direct logs streaming from application to Graylog.
It can be done using the log4j/slf4j/logback GELF plugins.

# Log Format for Applications

It is necessary to have the same log format in all the applications to allow:

* Parsing of a log message into several fields for convenience of searching and filtering operations.
* Concatenation of several lines of multiline log messages into one entry.

The following log format is a requirement:

`[%TIME%][%LEVEL%]%MESSAGE%`

Where:

* `%TIME% and %LEVEL%` are common core fields that contain the time of the event in `ISO8601` format
  and the log level respectively.
* `%MESSAGE%` is actual log message (can be multiline). It may contain any additional number of extra fields
  in the following format:

  * `[%KEY%=%VALUE%]` the key cannot contain spaces or be empty.

These fields are processed by the fluent-plugin-fields-parser plugin.

The base pattern for log4j2 is as follows:

`[%d{ISO8601}][%p]%m%n`

The following pattern should be used for earlier versions of log4j due to the incorrect date format when `%d{ISO8601}`
is used:

`[%d{yyyy-MM-dd'T'HH:mm:ss.SSS}][%p]%m%n`

The main requirement is to have `[%d{yyyy-MM-dd'T'HH:mm:ss.SSS}][%p]` at the beginning and `%n` at the end.
In this case all the features work.

**Note:** There are two important things in this pattern datetime format and two sections in square brackets at the beginning.

**Note:** In order to keep compatibility with SaaS microservice-framework,
it also supports multiline logs for `YYYY-MM-DD %m%n` format.

## Logs Traceability

In order to track operations across services, it is *mandatory* to have the `X-Request-ID` HTTP header in the logs.

You can easily add it to the log pattern in the `%X{requestId:-no_request_id}` format.
The pattern example is applicable for "logback" java library.

This field is parsed separately on Graylog side to `request_id` in each log entry.

**Note**: It is already included in the OOB configuration in the microservice-framework library.

## Complex Format Example

This section describes a complex format example.

The log4j patterns cheat sheet is as follows:

* %p - The log message priority.
* %C - The Java class which wrote the log message.
* %M - The method from which the log message was written.
* %t - The name of the thread.
* %m - The log message.

`[%d{yyyy-MM-dd'T'HH:mm:ss,SSS}][%p][log_id=my_special_log][class=%C] %M thread=%t %m%n`

The Fluentd sends each log message to Graylog in parsed form as a set of **field-value** pairs.
Each field is a separate column in Graylog which provides the ability to search, filter, and sort by a field.

The message with the pattern above is sent as follows:

* `[%d{yyyy-MM-dd'T'HH:mm:ss,SSS}]` at beginning and %n for a new line at the end are mandatory.
  Otherwise, Fluentd ignores the message. %d goes to `time` field.
* `[%p]` - Message level goes to `level` field.
* `[log_id=my_special_log]` - Creates a custom field `log_id` with constant value "my_special_log".
  This is helpful for searching only the special logs.
* `[class=%C]` - Creates a custom field `class` with the class name in it.
* `%M thread=%t %m` - This part interprets as a log message and goes to `message` field.
  For example, a message "MySpecialMethod thread=my_thread1 my very important log".
* It also has some metadata fields added by Kubernetes.

## JSON logs

Logs in pure JSON format also supported.
Such logs must be in single line JSON format.
All fields are parsed on fluentd level and indexed in Graylog as standalone fields.
Nice to have properties: level, message, time.

JSON log example:

```json
{"level":"info","task_name":"77bbfc6a-bda7-438f-b75f-168e3749f173","task_uuid":"ed4ca56b-4abb-4482-9e24-d70241264b7a","work_item_uuid":"0acd8972-d8aa-41e5-b3f4-e869bf5533a6","time":"2020-07-16T06:22:43Z","message":"Processing task."}
```

## Special Log Categories

Several special log categories are supported OOB. Based on some criteria logs are routed into separate streams in Graylog.

### Audit Logs

You need to add some marker for logs which contain messages that are sensitive and should be processed
in a special way by Graylog.

For example, `[logType=audit]` or any other keyword which can be easily parsed.
For more information about how audit logs processing is configured on Graylog side, refer to
*Cloud Platform Administrator's Guide*.

The following criteria for audit logs are provided with Graylog OOB:

* Services audit logs: The `message` field must match the regular expression `/access-control|PG_SERVICE|database\ssystem\sis|audit_log_type|AUDIT|logType[\":\s\t=]*audit|org\.qubership\.security\.audit.*CEF|org\.qubership\.cloud\.keycloak\.provider.*/`.
* VMs audit logs: The `tag` field must match exactly `parsed.var.log.audit.audit.log`.
* Graylog audit logs: The `container_name` field must match the regular expression `(graylog_web_1|graylog_graylog_1|graylog_mongo_1|graylog_elasticsearch_1)`.

#### JSON Audit logs

Logs in JSON format marked as Audit log if json body contains key-value `"logType":"audit"`.

### System Logs

Some logs are interpreted as system logs. The system logs are not application logs but logs of low-level parts
of the PaaS solution:

* Operation system logs (journald)
* OpenShift logs

The criteria for audit logs provided with Graylog OOB is the `tag` field must exactly match `parsed.var.log.messages|parsed.var.log.syslog|systemd`.

# Collecting Logs from Other Source

The logs from any other source, not supported by OOB, can be routed to Graylog. The configuration depends on logs source.
Following are some examples:

* Graylog OOB supports syslog format. Any application, microservice or system can send logs to Graylog via `rsyslog`.
* You can use the log4j GELF output plugin to directly send logs to Graylog from a java application.
  For example, it can be used for TOMS logs collection. This can be used for applications which are not hosted
  as OpenShift pods, but as standalone or bare-metal deployments. Note that no OpenShift related metadata
  information logs enrichment is performed.

**Note**: It is not allowed to send direct logs to Graylog's Elasticsearch. Graylog encapsulates Elasticsearch
and uses it as a backend log storage.

# Best Practices for Logging

This section provides information about best practices of logs content and configuration.

## Log Amount and Content

It is very important to understand about optimal log amount.
Too many logs with a lot of information are not useful for quick and efficient analysis of errors and incidents.
Similarly, too fewer logs are also not useful.

It is important to write only logs that are necessary.

To write optimal logs, you can follow some rules:

* Add context information to logs, UserId, TransactionId, ThreadId and so on. It helps to track potential error causes.
* Do not omit stack traces.
* Huge log amount can slow down the Logging Service.
* Use human friendly and readable messages.
* Avoid logs such as "error", "operation failed" and so on which do not provide any useful information.
  Provide helpful information for investigations later.
* Track interaction with other services/systems.

## Logging Levels

You can use different log levels. Printing trace logs after each line of code can be helpful sometimes,
but it is not possible to write such logs every time.

Typical logging level on production is `INFO`.
The `INFO` log level should contain high level information about the successful operations and detail information
about the failed operations.

In case of investigating floating and another tricky errors, it is possible to enable `DEBUG` or even `TRACE` level
which write logs about every operation in the application.

For example, log4j levels description:

* `ERROR` - Designates error events that might still allow the application to continue running
* `WARN` - Designates potentially harmful situations
* `INFO` - Designates informational messages that highlight the progress of the application at coarse-grained level
* `DEBUG` - Designates fine-grained informational events that are most useful to debug an application
* `TRACE` - Designates finer-grained informational events than `DEBUG`

## Log level configuration

The application should provide an ability to change the logging settings anytime without changing the source code
and without redeploying the application.

You can do it by exposing the logging configuration to ENVs or to config-map.

In order to apply logging configuration changes without redeploying/restarting microservice, it is possible
to create special REST endpoint which trigger logging configuration re-initializing.
