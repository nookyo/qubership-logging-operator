This section describes the Logging Service dashboards.

# Table of Content

* [Table of Content](#table-of-content)
* [Overview](#overview)
* [Monitoring](#monitoring)
  * [Metrics](#metrics)
  * [Dashboards](#dashboards)
* [Logging](#logging)
* [Tracing](#tracing)
* [Profiler](#profiler)

# Overview

| Observability part        | Integration status                |
| ------------------------- | --------------------------------- |
| [Monitoring](#monitoring) | ✅ Support                         |
| [Logging](#logging)       | ✅ Support                         |
| [Tracing](#tracing)       | ⛔️ No, Graylog has no this feature |
| [Profiler](#profiler)     | ⛔️ No, due the license             |

# Monitoring

All Logging components that have valuable metrics are already integrated with Monitoring.
It means that:

* their metrics enabled and exposed the endpoint with metrics in Prometheus format
* during deployment will create ServiceMonitors/PodMonitors to integrate with the Monitoring

Components with metrics:

* Graylog
* MongoDB
* OpenSearch
* FluentBit
* FluentD

## Metrics

* Graylog - [https://go2docs.graylog.org/5-0/interacting_with_your_log_data/metrics.html#PrometheusMetricExporting](https://go2docs.graylog.org/5-0/interacting_with_your_log_data/metrics.html#PrometheusMetricExporting)
* FluentBit - [https://docs.fluentbit.io/manual/administration/monitoring](https://docs.fluentbit.io/manual/administration/monitoring)
* FluentD - [https://docs.fluentd.org/monitoring-fluentd/monitoring-prometheus](https://docs.fluentd.org/monitoring-fluentd/monitoring-prometheus)

# Logging

The Logging agent, FluentBit or FluentD in the Cloud by default collects logs from all nodes.
So the logging agents will collect all logs from Logging components in the Cloud:

* Graylog (if deployed in the Cloud)
  * MongoDB
* FluentBit/FluentD
  * Configmap Reload
* Cloud Events Reader

In the case of deploying Graylog on the VM logs from Graylog and other components on the VM will collected
by the logging agent (FluentBit/FluentD).

Log format examples for different tools in the Logging stack:

* Graylog (text format)

    ```bash
    [2024-01-01T00:00:00,329][INFO]Successfully ensured index template gray_audit-template
    [2024-01-01T00:00:01,292][INFO]Waiting for allocation of index <gray_audit_1>.
    [2024-01-01T00:00:03,745][INFO]Index <gray_audit_1> has been successfully allocated.
    [2024-01-01T00:00:03,745][INFO]Pointing index alias <gray_audit_deflector> to new index <gray_audit_1>.
    [2024-01-01T00:00:04,099][INFO]Submitted SystemJob <b73d0ad0-a838-11ee-bd6b-52ae94ffd025> [org.graylog2.indexer.indices.jobs.SetIndexReadOnlyAndCalculateRangeJob]
    [2024-01-01T00:00:04,099][INFO]Successfully pointed index alias <gray_audit_deflector> to index <gray_audit_1>.
    ```

* MongoDB (JSON format)

    ```bash
    {"t":{"$date":"2024-01-04T17:43:31.543+00:00"},"s":"I",  "c":"STORAGE",  "id":22430,   "ctx":"Checkpointer","msg":"WiredTiger message","attr":{"message":"[1704390211:543503][7:0x7f9631a66700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 1554590, snapshot max: 1554590 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0) base write gen: 2867"}}
    {"t":{"$date":"2024-01-04T17:44:31.569+00:00"},"s":"I",  "c":"STORAGE",  "id":22430,   "ctx":"Checkpointer","msg":"WiredTiger message","attr":{"message":"[1704390271:569453][7:0x7f9631a66700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 1554754, snapshot max: 1554754 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0) base write gen: 2867"}}
    ```

* OpenSearch (text format)

    ```bash
    [2024-01-04T17:50:58,716][INFO ][o.o.j.s.JobSweeper       ] [opensearch-0] Running full sweep
    [2024-01-04T17:51:09,665][INFO ][sgaudit                  ] [opensearch-0] {"audit_trace_task_parent_id":"cKWOrPT-TZKwEPjOQcgG5A:8488341","audit_cluster_name":"opensearch","audit_transport_headers":{"_opendistro_security_remote_address_header":"rO0ABXNyABpqYXZhLm5ldC5JbmV0U29ja2V0QWRkcmVzc0ZxlGFv+apFAwADSQAEcG9ydEwABGFkZHJ0ABZMamF2YS9uZXQvSW5ldEFkZHJlc3M7TAAIaG9zdG5hbWV0ABJMamF2YS9sYW5nL1N0cmluZzt4cgAWamF2YS5uZXQuU29ja2V0QWRkcmVzc0hh9mL0l51qAgAAeHAAAObUc3IAFGphdmEubmV0LkluZXRBZGRyZXNzLZtXr5/j69sDAANJAAdhZGRyZXNzSQAGZmFtaWx5TAAIaG9zdE5hbWVxAH4AAnhwCoH8lQAAAAJweHB4","_opendistro_security_initial_action_class_header":"GetMappingsRequest","_opendistro_security_origin_header":"REST","_opendistro_security_user_header":"rO0ABXNyACFvcmcub3BlbnNlYXJjaC5zZWN1cml0eS51c2VyLlVzZXKzqL2T65dH3AIABloACmlzSW5qZWN0ZWRMAAphdHRyaWJ1dGVzdAAPTGphdmEvdXRpbC9NYXA7TAAEbmFtZXQAEkxqYXZhL2xhbmcvU3RyaW5nO0wAF29wZW5EaXN0cm9TZWN1cml0eVJvbGVzdAAPTGphdmEvdXRpbC9TZXQ7TAAPcmVxdWVzdGVkVGVuYW50cQB+AAJMAAVyb2xlc3EAfgADeHAAc3IAEWphdmEudXRpbC5IYXNoTWFwBQfawcMWYNEDAAJGAApsb2FkRmFjdG9ySQAJdGhyZXNob2xkeHA/QAAAAAAAAHcIAAAAEAAAAAB4dAAGbmV0Y3Jrc3IAEWphdmEudXRpbC5IYXNoU2V0ukSFlZa4tzQDAAB4cHcMAAAAED9AAAAAAAACdAAQbWFuYWdlX3NuYXBzaG90c3QACmFsbF9hY2Nlc3N4cHNxAH4ACHcMAAAAED9AAAAAAAABdAAFYWRtaW54","_opendistro_security_remotecn":"opensearch"},"audit_node_name":"opensearch-0","audit_trace_task_id":"-oFFf1ZaT-6lz_kLJId0GQ:10511485","audit_transport_request_type":"GetMappingsRequest","audit_category":"INDEX_EVENT","audit_request_origin":"REST","audit_node_id":"-oFFf1ZaT-6lz_kLJId0GQ","audit_request_layer":"TRANSPORT","@timestamp":"2024-01-04T17:51:09.664+00:00","audit_format_version":4,"audit_request_remote_address":"1.2.3.4","audit_request_privilege":"indices:admin/mappings/get","audit_node_host_address":"1.2.3.4","audit_request_effective_user":"test","audit_trace_indices":["graylog_0"],"audit_trace_resolved_indices":["graylog_0"],"audit_node_host_name":"1.2.3.4"}
    ```

* FluentBit (text format)

    ```bash
    Fluent Bit v2.1.10
    * Copyright (C) 2015-2022 The Fluent Bit Authors
    * Fluent Bit is a CNCF sub-project under the umbrella of Fluentd
    * https://fluentbit.io
    ```

* FluentD (text format)

    ```bash
    2024-01-04 16:20:23 +0000 [info]: #1 following tail of /var/log/pods/monitoring_monitoring-tests-7668c54bb4-pvltd_adf47226-7234-4cd6-8dd9-1b0546a0bf79/monitoring-tests/247.log
    2024-01-04 16:22:23 +0000 [info]: #1 following tail of /var/log/pods/rabbitmq-saja_rabbitmq-integration-tests-6b97d7687d-8b49k_7d7ba982-1c39-453b-8fd4-12dd1b57c0c5/rabbitmq-integration-tests/210.log
    2024-01-04 16:26:23 +0000 [info]: #1 following tail of /var/log/pods/mongodb_datars12-0_63751eca-0e1b-4bdc-bee7-98414f6a2985/datars12/0.log
    ```

  * Configmap Reload (text format)

      ```bash
    2024/09/12 09:00:31 Watching directory: "/fluentd/etc"
    2024/09/12 09:03:57 config map updated
    2024/09/12 09:03:57 performing webhook request (1/1)
    2024/09/12 09:03:57 successfully triggered reload
      ```

* Cloud Events Reader (JSON format)

    ```bash
    {"time":"2024-03-07T09:23:17.349","level":"INFO","source":"main.go:43","msg":"Starting Cloud Events Reader..."}
    {"time":"2024-03-07T09:23:17.349","level":"WARN","source":"event_format.go:17","msg":"Template format is not set. Default is used."}
    {"time":"2024-03-07T09:23:17.455","level":"INFO","source":"events_reader_controller.go:107","msg":"Started workers"}
    {"time":"2024-03-07T09:22:58.432","involvedObjectKind":"GrafanaFolder","involvedObjectNamespace":"monitoring","involvedObjectName":"public-stats-folder","involvedObjectUid":"80acdb27-cde8-4b96-acd5-cff761072684","involvedObjectApiVersion":"integreatly.org/v1alpha1","involvedObjectResourceVersion":"47301469","reason":"Success","type":"Normal","message":"folder monitoring/public-stats-folder successfully submitted","kind":"KubernetesEvent"}
    ```

# Tracing

Graylog has no integration with Tracing, neither with Jaeger nor with OpenTelemetry.

# Profiler

Graylog has no integration with Cloud Diagnostic Toolset / Profiler.

Graylog uses an `SSPL` license, so we can't modify it base image to add the CDT agent inside.
