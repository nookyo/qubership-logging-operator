
# Table of Content

* [Table of Content](#table-of-content)
* [Overview](#overview)
  * [Graylog](#graylog)
  * [OpenSearch/ElasticSearch](#opensearchelasticsearch)
  * [MongoDB](#mongodb)
  * [FluentBit/FluentD](#fluentbitfluentd)
  * [Configmap Reload](#configmap-reload)
  * [Cloud Events Reader](#cloud-events-reader)
  * [Logging backup daemon](#logging-backup-daemon)
* [Supported deployment schemes](#supported-deployment-schemes)
  * [On-prem](#on-prem)
    * [Non-HA deployment scheme scheme](#non-ha-deployment-scheme-scheme)
    * [HA deployment scheme](#ha-deployment-scheme)
    * [DR deployment scheme](#dr-deployment-scheme)
  * [Integration with managed services](#integration-with-managed-services)
    * [AWS OpenSearch](#aws-opensearch)
    * [AWS CloudWatch](#aws-cloudwatch)
    * [Azure Log Analytic](#azure-log-analytic)
  * [Integration with other vendor solutions](#integration-with-other-vendor-solutions)
    * [Splunk](#splunk)

# Overview

The logging service is designed to collect various logs from different sources.
The main purpose is to collect application logs from the Kubernetes namespaces.

The collection of the following logs is supported:

* Application logs such as POD logs from Cloud
* System logs such as syslog logs from Kubernetes nodes
* HW nodes, if possible
* Operation services logging such as Monitoring VM, Deploy VM, and Logging VM

Custom log processing from any other source and integration with external log processing systems is also available.

The collected logs can be split into groups with different access permissions to avoid unauthorized access to sensitive data.

The logging service is based on a third-party solution, Graylog. For more information,
refer to the [Official Graylog Documentation](https://go2docs.graylog.org).

Qubership provides a log server with NC extensions as part of the solution.
This component is optional but is required for troubleshooting and operations.

Qubership extensions include:

* Extended Kubernetes-oriented logs parsing. The namespace, cloud, microservice, and other specific metrics
  are extracted as fields and are available for quick filtering and histogram display.
* Pattern-based event detection, and transmission to monitoring for further processing and alarming.
  For example, unsuccessful authentication attempts.
* OOB monitoring dashboards that monitor Graylog's health and performance metrics.
* Automated deployment in various schemas such as standalone, DR, HA, and so on.
* Basic configuration OOB such as logs segregation, users, logs rotation policies, and so on.
* Operations such as backup/restore, troubleshooting, maintenance, and so on.

All Graylog parts are deployed on a dedicated VM or Cloud in docker containers.
Graylog supports many input protocols, and can interact with almost all log agents such as the following:

* Cloud Logs Collection - By default, `fluentbit` is used in the solution as Kubernetes DaemonSet for application
  logs collection. For more information about Fluentd, refer to the _Official Fluentd Documentation_
  at [https://www.fluentd.org/architecture](https://www.fluentd.org/architecture).
  We provide fluentbit configured for collection application logs, system logs, cloud audit logs.
* Operations Services Logs Collection - For services that are deployed on dedicated VMs such as Logging, Monitoring,
  and Deploy VM, `td-agent`, an RPM version of Fluentd is installed on these VMs.

## Graylog

The log processing engine. It receives logs from collectors such as `FluentBit`/`FluentD` in this case, processes them,
and sends them to the storage. It also contains Web UI for configuring and viewing the logs.

Official documentation:

* Graylog [https://go2docs.graylog.org/5-2/home.htm](https://go2docs.graylog.org/5-2/home.htm)

## OpenSearch/ElasticSearch

Used by Graylog as a log storage.

`OpenSearch` is the flexible, scalable, open-source way to build solutions for data-intensive applications.
Was forked from `ElasticSearch`, improved and now supported by Amazon.

`ElasticSearch` is a distributed, RESTful search and analytics engine capable of addressing a growing number of use cases.
As the heart of the Elastic Stack, it centrally stores your data for lightning-fast search, fine‑tuned relevancy,
and powerful analytics that scale with ease.

Official documentation:

* OpenSearch [https://opensearch.org/docs/latest/](https://opensearch.org/docs/latest/)
* ElasticSearch [https://www.elastic.co/guide/index.html](https://www.elastic.co/guide/index.html)

## MongoDB

The small instance of `MongoDB` used by `Graylog` as a settings storage.

`MongoDB` is a document database designed for ease of application development and scaling.

## FluentBit/FluentD

`FluentBit` is a CNCF sub-project under the umbrella of `FluentD`.

`FluentBit` is an open-source telemetry agent specifically designed to efficiently handle the challenges of collecting
and processing telemetry data across a wide range of environments, from constrained systems to complex cloud infrastructures.
Managing telemetry data from various sources and formats can be a constant challenge, particularly
when performance is a critical factor.

Rather than serving as a drop-in replacement, `FluentBit` enhances the observability strategy for your infrastructure
by adapting and optimizing your existing logging layer, as well as metrics and trace processing.
Furthermore, Fluent Bit supports a vendor-neutral approach, seamlessly integrating with other ecosystems such as
Prometheus and OpenTelemetry. Trusted by major cloud providers, banks, and companies in need of a ready-to-use
telemetry agent solution, Fluent Bit effectively manages diverse data sources and formats while
maintaining optimal performance.

`FluentD` is an open-source data collector for a unified logging layer.
`FluentD` allows you to unify data collection and consumption for better use and understanding of data.

Official documentation:

* FluentBit [https://docs.fluentbit.io/](https://docs.fluentbit.io/)
* FluentD [https://docs.fluentd.org/](https://docs.fluentd.org/)

## ConfigMap Reload

`ConfigMap-reload` is a simple binary to trigger a reload when Kubernetes ConfigMaps or Secrets,
mounted into pods, are updated. It watches mounted volume dirs and notifies the target process that
the config map has been changed.
`ConfigMap-reload` allows you to update Fluents' config without restarting pods.

Official documentation:

* ConfigMap-reload [https://github.com/jimmidyson/configmap-reload/pkgs/container/configmap-reload](https://github.com/jimmidyson/configmap-reload/pkgs/container/configmap-reload)

## Cloud Events Reader

Cloud Events Reader is a deployment that observes for Kubernetes events and prints it to logs in predefined format
(to be processed by Fluentd/FluentBit). It is deployed as a part of the Logging stack.

It implements Kubernetes controller that watches for kind Event with API version events.k8s.io/v1 adding and modifying
and prints formatted data of event in stdout.

## Logging backup daemon

# Supported deployment schemes

## On-prem

This section describes the Logging stack deployment in the on-premise Kubernetes in different schemes.

### Non-HA deployment scheme scheme

Non-HA deployment supports for both clients: FluentD and FluentBit. Each client sends processed messages directly in
Graylog.

![Non-HA](/docs/images/architecture/cloud-fluentbit.png)

### HA deployment scheme

Now the Logging stack partially supports High-Availability deploy in the Cloud.

HA deployment support only for FluentBit. In this schema, FluentBit will deploy in two parts:

* FluentBit forwarder - will deploy as DaemonSet to read logs from all Kubernetes nodes
* FluentBit aggregator - will deploy as StatefulSet (with PV) to receive logs from the forwarder, process them
  and send them to the destination

Other components can't be deployed in more than 1 replica (or it has no sense):

* Graylog - currently we don't support it deploying with more than 1 replica
* MongoDB - currently deploying as a sidecar of Graylog so can't be run in more than 1 replica
* Cloud Event Reader - stateless service and can be run in N replicas, but it has no make sense

![Non-HA](/docs/images/architecture/cloud-fluentbit-ha.png)

### DR deployment scheme

Currently Logging in the Cloud has no specific deployment schema.
It means that on both sites all Logging components will deploy in the Cloud independently. And you will have
two (or more) Logging stacks on your right and left Kubernetes.

If you try to plan your Kubernetes DR deployment, you need to multiply the required resources by 2
(or more if you have more Kubernetes in the DR schema).

## Integration with managed services

This section describes the abilities of Logging stack deploy using Public Cloud Managed Services.

### AWS OpenSearch

**Type:** Graylog integration

In the case of deploy Graylog in the Cloud and AWS, you can use `AWS OpenSearch` as a log storage.

Graylog in the Cloud requires separated OpenSearch/ElasticSearch to store logs. Because AWS provides
a lot of managed services (including AWS OpenSearch) you can replace self-managed OpenSearch clusters
with AWS Managed Service.

### AWS CloudWatch

**Type:** FluentBit integration

FluentBit can send data to `AWS CloudWatch` in two schemes:

* Together with sending logs to Graylog
* Send logs only to `AWS CloudWatch`

For more details, on how to configure integration with `AWS CloudWatch` please refer to the integration guide
[Amazon CloudWatch](https://docs.fluentbit.io/manual/pipeline/outputs/cloudwatch).

### Azure Log Analytic

**Type:** FluentBit integration

FluentBit can send data to `Azure Log Analytic` in two schemes:

* Together with sending logs to Graylog
* Send logs only to `Azure Log Analytic`

For more details, on how to configure integration with `Azure Log Analytic` please refer to the integration guide
[Azure Log Analytic](https://docs.fluentbit.io/manual/pipeline/outputs/azure).

## Integration with other vendor solutions

This section describes possible integrations of the Logging stack with solutions from other vendors.
For example, Splunk, New Relic, Datadog and so on.

### Splunk

**Type:** FluentBit integration

FluentBit can send data to `Splunk` in two schemes:

* Together with sending logs to Graylog
* Send logs only to `Splunk`

For more details, on how to configure integration with `Splunk` please refer to the integration guide
[Splunk](integrations/splunk.md).
