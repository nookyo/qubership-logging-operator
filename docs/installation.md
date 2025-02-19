
# Table of Content

* [Table of Content](#table-of-content)
* [Prerequisites](#prerequisites)
  * [Common](#common)
    * [Storage](#storage)
      * [OpenSearch/ElasticSearch supported versions](#opensearchelasticsearch-supported-versions)
      * [Graylog Persistence Volumes](#graylog-persistence-volumes)
      * [HostPath Persistence Volumes](#hostpath-persistence-volumes)
      * [Storage capacity planning](#storage-capacity-planning)
        * [Log `storm`](#log-storm)
        * [Indices with rotation by size](#indices-with-rotation-by-size)
        * [Indices with rotation by time or messages count](#indices-with-rotation-by-time-or-messages-count)
  * [Kubernetes](#kubernetes)
  * [OpenShift](#openshift)
  * [Amazon Web Services (AWS)](#amazon-web-services-aws)
  * [Azure](#azure)
  * [Google Cloud](#google-cloud)
* [Best practices and recommendations](#best-practices-and-recommendations)
  * [HWE](#hwe)
    * [Small](#small)
    * [Medium](#medium)
    * [Large](#large)
  * [Logging on different OS](#logging-on-different-os)
    * [System logs](#system-logs)
    * [Audit logs](#audit-logs)
    * [Kubernetes and container logs](#kubernetes-and-container-logs)
* [Parameters](#parameters)
  * [Root](#root)
  * [Graylog](#graylog)
    * [Graylog TLS](#graylog-tls)
    * [OpenSearch](#opensearch)
    * [ContentPacks](#contentpacks)
    * [Graylog Streams](#graylog-streams)
    * [Graylog Auth Proxy](#graylog-auth-proxy)
      * [Graylog Auth Proxy LDAP](#graylog-auth-proxy-ldap)
      * [Graylog Auth Proxy OAuth](#graylog-auth-proxy-oauth)
  * [FluentBit](#fluentbit)
    * [FluentBit Aggregator](#fluentbit-aggregator)
    * [FluentBit TLS](#fluentbit-tls)
  * [FluentD](#fluentd)
    * [FluentD TLS](#fluentd-tls)
  * [Cloud Events Reader](#cloud-events-reader)
  * [Integration tests](#integration-tests)
* [Installation](#installation)
  * [Before you begin](#before-you-begin)
  * [On-prem](#on-prem)
  * [Amazon Web Services (AWS)](#amazon-web-services-aws-1)
* [Post Installation Steps](#post-installation-steps)
  * [Configuring URL whitelist](#configuring-url-whitelist)
* [Upgrade](#upgrade)
* [Post Deploy Checks](#post-deploy-checks)
  * [Jobs Post Deploy Check](#jobs-post-deploy-check)
  * [Smoke test](#smoke-test)
  * [Integration tests](#integration-tests-1)
* [Frequently asked questions](#frequently-asked-questions)
* [Footnotes](#footnotes)

# Prerequisites

## Common

* Kubernetes 1.21+ or OpenShift 4.10+
* kubectl 1.21+ or oc 4.10+ CLI
* Helm 3.0+

### Storage

#### OpenSearch/ElasticSearch supported versions

In the case of deploy Graylog in the Cloud, you should use OpenSearch or ElasticSearch only specified below versions.

| Graylog version | ElasticSearch versions    | OpenSearch versions     |
| --------------- | ------------------------- | ----------------------- |
| Graylog 4.x     | `6.8.x`, `7.7.x - 7.10.x` | `1.x *`                 |
| Graylog 5.x     | `6.8.x`, `7.10.2`         | `1.x`, `2.0.x-2.5.x **` |

where:

* `*` - for Graylog 4.x OpenShift 1.x must be deployed and run **with** compatibility mode
* `**` - for Graylog 5.x OpenShift 2.x must be deployed and run **without** compatibility mode

**Note:** OpenSearch or ElasticSearch versions not specified in the table above may not be supported. Cloud Infra
Platform can't guarantee correct work Graylog with not specified OpenSearch or ElasticSearch versions.

Information about compatibility mode:

* [Moving from open-source Elasticsearch to OpenSearch](https://opensearch.org/blog/moving-from-opensource-elasticsearch-to-opensearch/)

#### Graylog Persistence Volumes

Graylog requires two Persistence Volumes (PVs):

* for in-build MongoDB, used to store Graylog configurations
* for journald, used as a cache between Graylog input and processing

NFS-like storage **doesn't support!** It means that you should use PV and dynamic PR providers that use storage with
types: NFS, AWS EFS, Azure File or any other NFS-like storage.

For Graylog `journald` storage please select a storage with enough throughput and speed. Graylog may execute a lot of
read and write operations in `journald` when it will have a high load.

#### HostPath Persistence Volumes

**Note:** In case of Graylog deployment with `hostPath` PV you **must** correctly set `nodeSelector` parameter
for unambiguously determine of node to install Graylog.

Need to execute some preparatory steps if you want to deploy Graylog (on MongoDB inside the Graylog)
on the hostPath Persistence Volume (PV).

Read more about hostPath PVs and their limitations you can in the official documentation:
[https://kubernetes.io/docs/concepts/storage/volumes/#hostpath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)

To using hostPath PV you need to grant permissions for directories inside the PV.

First, in case of deploy in the OpenShift, it by default has a range of UIDs with that it will run pods
(and containers inside) in the namespace. If you plan to deploy in the Kubernetes you can skip this step.

So first of all you need to set only one UID that OpenShift will use. You can do it using a command:

```bash
oc annotate --overwrite namespace <logging_namespace> openshift.io/sa.scc.uid-range=<uid>/<uid>
```

for example:

```bash
oc annotate --overwrite namespace logging openshift.io/sa.scc.uid-range=1100/1100
```

Second, you need to create, set ownership and grant permissions for directories. For our example let\'s assume that
hostPath PV create by path:

```bash
/mnt/graylog-0
```

and you configure `UID = 1100`.

So you need to execute the following commands:

```bash
mkdir /mnt/graylog-0/config
chown -R 1100:1100 /mnt/graylog-0
chmod 777 /mnt/graylog-0
chmod 666 /mnt/graylog-0/config
```

If you are using a OS with SELinux you may need to set SELinux security context:

```bash
chcon -R unconfined_u:object_r:container_file_t:s0 /mnt/graylog-0
```

#### Storage capacity planning

To calculate the total storage size required to store logs from you environment need to calculate
how many logs planning store in **each** Stream/Index.

Default Indices:

* Default index set (`All messages` Stream)
* Audit index set (`Audit logs` Stream)
* Graylog Events (`All events` Stream)
* Graylog System Events (`All system events` Stream)
* Kubernetes events index set (`Kubernetes events` Stream)

Except default Streams/Indices products/projects can create it's own Streams/Indices. They also should
be include in the calculation.

All received logs Graylog saved in OpenSearch/ElasticSearch. The OpenSearch/ElasticSearch has protection
to prevent OpenSearch nodes from running out of disk space.

By default, OpenSearch marks all indexes as read-only if data usage reaches a set threshold.
This threshold by default set as **95%** of available disk or volume space.

Next, for the expected log size for **N** days, you have to add **15-25%** of free space to avoid problems
with the read-only index and log rotation.

So the resulting formula will be something like this:

```bash
(Expected size of Index 1 + Expected size of Index 2 + ... + Expected size of Index N) / 0.80 = Total size of storage
```

Based on the calculated size, you have to set other parameters, like max index size, count of indexes,
and so on.
In the case of storing a big log count (by size), better to increase the max index size
from 1 Gb to 5-10-20 Gb.

##### Log `storm`

Do not forgot to add in you calculations reserve for case of `log storm`.
In some cases, problems with one service can lead to increased log generation in other related services.

THe simple example:

* there are 20 Java services that are using PostgreSQL
* all these services generating about 0-100 log lines per minute
* but in moment PostgreSQL become unavailable (due the internal or network problems)
* all 20 Java services may start generate in their logs giant stacktraces with errors about PostgreSQL unavailability
* log generation for these 20 Java services can increase in **10-100 times** for PostgreSQL unavailability period

In this example, in the case of problems with PostgreSQL in Graylog may be send 200000 logs per minute
instead of 2000 per minute which we expect during the normal work.

##### Indices with rotation by size

You **have to estimate** how many logs per day your Graylog received for each Stream/Index.

**Note:** For `All messages` Stream the simplest way is using information from Graylog UI,
navigate to `Graylog UI -> System -> Overview` and check values in histogram.

For example, if on this page you see:

* 50 GB for 1 day ago
* 100 GB for 2 days ago

you can calculate the average value or select the most pessimistic value.
Next, need to multiply the selected value by the count of days.

**Note:** You may add reserve for log `storm`, but it's optional unlike for case using rotation
by time or message count.

So the resulting formula will be something like this:

```bash
X Gb (Logs size per day) * Y days = Total size of Index
```

For example:

```bash
100 Gb * 7 days = 700 Gb of Index
```

Also, you have to keep in mind that in Graylog you have some Streams and Indexes.
So this calculation must be done for all existing Indexes.

As an alternative you can use the metrics like:

```prometheus
avg(gl_input_read_bytes_one_sec[1h])
```

or (it's not a PromQL expression, just an example):

```prometheus
rate(gl_input_read_bytes_total) / rate(gl_input_incoming_messages_total)
```

But in this case, you have to be very accurate because the load can change over time.
You can't calculate it only for one time point, you have to calculate for the time range
and next calculate the average or max value.

##### Indices with rotation by time or messages count

In the case, if you want to use Streams with Indexes rotated by time or message count,
you **have to be extremely careful** and estimate the expected incoming log flow very well.

Unlike rotation by size, using the rotation by time or message count can lead to index/disk overflow
if expected logs size will be calculated incorrect. It may affect logs storing in OpenSearch
and in all other Streams.

There is no a simple and unambiguous formula to calculate the required storage size for Indices
with rotation by time or messages count. But you can try to start from the formula:

* Rotation by time

  ```bash
  X Gb (Logs size per day) * Y days + Z Gb (log storm reserve) = Total size of Index
  ```

* Rotation by message count:

  ```bash
  X Kb (average 1 message size) * Y messages count + Z Gb (log storm reserve) = Total size of Index
  ```

  **Note:** Please keep in mind, that each message has a lot of meta information (additional fields),
  like `namespace`, `pod`, `container`, labels and others that also should be include in the 1 message size.

## Kubernetes

According to the platform\'s third-party support policy now we are supporting deploy in Kubernetes N +- 2.

The current recommended Kubernetes version is `1.28.x`, so we support:

| Kubernetes version     | Status of `0.48.0`    |
| ---------------------- | --------------------- |
| `1.25.x`               | Tested                |
| `1.26.x`               | Tested                |
| `1.28.x` (recommended) | Tested                |
| `1.29.x`               | Forward compatibility |
| `1.30.x`               | Forward compatibility |

**Note:** `Forward compatibility` means that the latest version doesn't use any Kubernetes APIs that will be removed
in the next Kubernetes versions. All Kubernetes APIs verified by official documentation
[Deprecated API Migration Guide](https://kubernetes.io/docs/reference/using-api/deprecation-guide/).

To deploy Logging in the Kubernetes/OpenShift you must have at least a namespace admin role.
You should have at least permissions like the following:

```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  namespace: <logging-namespace>
  name: deploy-user-role
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
```

**Note:** It\'s not a role that you have to create. It\'s just an example with a minimal list of permissions.

For Kubernetes 1.25+, you **must** deploy Logging using `privileged` PSS. The logging agents require using
`hostPath` PVs to mount directories with logs from Kubernetes/OpenShift nodes.

Before deploying please make sure that your namespace has the following labels:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: <logging_namespace_name>
  labels:
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/enforce-version: latest
```

[Back to TOC](#table-of-content)

## OpenShift

**Note:** OpenShift 4.x is based on Kubernetes and regularly picks up new Kubernetes releases. So compatibility with
OpenShift you can track by Kubernetes version. To understand which Kubernetes version OpenShift use you can read
it release notes. For example, `OpenShift 4.12` is based on `Kubernetes v1.25.0` about which wrote in
[Release notes](https://docs.openshift.com/container-platform/4.12/release_notes/ocp-4-12-release-notes.html).

To deploy in the OpenShift you need to:

* Run FluentBit/FluentD in `privileged` mode
* Create Security Context Constraints (SCC)

Run FluentBit/FluentD in `privileged` mode is mandatory. Otherwise logging agents can't get access to log
files on nodes.

By default, OpenShift store log files with permissions `600` and `root` ownership.

```yaml
fluentbit:
  securityContextPrivileged: true
```

or

```yaml
fluentd:
  securityContextPrivileged: true
```

[Back to TOC](#table-of-content)

## Amazon Web Services (AWS)

| AWS Managed Service | Graylog support |
| ------------------- | --------------- |
| AWS OpenSearch      | ✅ Support       |

Supported OpenSearch version can be found in the section
[OpenSearch/ElasticSearch supported versions](#opensearchelasticsearch-supported-versions).

In the case of using Graylog with AWS Managed OpenSearch, you should choose OpenSearch instance flavor with
hardware resources not less than:

* CPU - 2 core
* Memory - 4 Gb
* Storage type - SSD

Details about the required HWE can be found in the section [HWE](#hwe).

[Back to TOC](#table-of-content)

## Azure

| AWS Managed Service | Graylog support |
| ------------------- | --------------- |
| Azure OpenSearch    | N/A             |

Azure has no officially managed OpenSearch or ElasticSearch. You can find only custom solutions
in the Azure marketplace from other vendors.

[Back to TOC](#table-of-content)

## Google Cloud

| AWS Managed Service | Graylog support |
| ------------------- | --------------- |
| Azure OpenSearch    | N/A             |

Google has no officially managed OpenSearch or ElasticSearch. You can find only custom solutions
in the Google marketplace from other vendors.

[Back to TOC](#table-of-content)

# Best practices and recommendations

## HWE

The following table shows the typical throughput/HWE ratio:

Graylog:

<!-- markdownlint-disable line-length -->
| Input logs, msg/sec            | <1000  | 1000-3000 | 5000-7500  | 7500-10000 | 10000-15000 | 15000-25000 | >25000 |
| ------------------------------ | ------ | --------- | ---------- | ---------- | ----------- | ----------- | ------ |
| CPU                            | 4      | 6         | 8          | 8          | 12          | 12          | 16+    |
| Graylog heap, Gb               | 1      | 2         | 2          | 4          | 4           | 6           | 6      |
| Total RAM, Gb                  | 6      | 8         | 12         | 16         | 16          | 22          | 24+    |
| HDD volume, 1/day (very rough) | <80 Gb | 80-200 Gb | 300-600 Gb | 600-800 Gb | 0.8-1 Tb    | 1-2 Tb      | 2+ Tb  |
| Disk speed, Mb/s               | 2      | 5         | 10         | 20         | 30          | 50          | 100    |
<!-- markdownlint-enable line-length -->

OpenSearch/ElasticSearch:

<!-- markdownlint-disable line-length -->
| Input logs, msg/sec            | <1000  | 1000-3000 | 5000-7500  | 7500-10000 | 10000-15000 | 15000-25000 | >25000                     |
| ------------------------------ | ------ | --------- | ---------- | ---------- | ----------- | ----------- | -------------------------- |
| CPU                            | 4      | 6         | 8          | 8          | 12          | 12          | 16+                        |
| ES heap, Gb                    | 2      | 4         | 8          | 8          | 8           | 12          | 16+ (but less that ~32 GB) |
| Total RAM, Gb                  | 6      | 8         | 12         | 16         | 16          | 22          | 24+                        |
| HDD volume, 1/day (very rough) | <80 Gb | 80-200 Gb | 300-600 Gb | 600-800 Gb | 0.8-1 Tb    | 1-2 Tb      | 2+ Tb                      |
| Disk speed, Mb/s               | 2      | 5         | 10         | 20         | 30          | 50          | 100                        |
<!-- markdownlint-enable line-length -->

[Back to TOC](#table-of-content)

### Small

Resources in this profile were calculated for the average load `<= 3000` messages per second.

**Warning!** All resources below may require a tuning for your environment because different environment
can has different log types (for example, the average message size), different retention and so on.
So please use carefully, adjust if necessary and better to execute SVT for Logging before run in production.

| Component            | CPU Requests | Memory Requests | CPU Limits | Memory Limits |
| -------------------- | ------------ | --------------- | ---------- | ------------- |
| Graylog              | `500m`       | `1500Mi`        | `1000m`    | `1500Mi`      |
| FluentD              | `100m`       | `128Mi`         | `500m`     | `512Mi`       |
| FluentBit Forwarder  | `100m`       | `128Mi`         | `300m`     | `256Mi`       |
| FluentBit Aggregator | `300m`       | `256Mi`         | `500m`     | `1Gi`         |
| Cloud Events Reader  | `50m`        | `128Mi`         | `100m`     | `128Mi`       |

**Important!** In the case of deploy Graylog in the Cloud, you need to include in the calculation resources for
OpenSearch Cluster (recommended) or OpenSearch single instance that will deploy in the Cloud. Please refer
to the OpenSearch documentation to find OpenSearch hardware requirements.

[Back to TOC](#table-of-content)

### Medium

Resources in this profile were calculated for the average load between `> 3000` and `<= 10000` messages per second.

**Warning!** All resources below may require a tuning for your environment because different environment
can has different log types (for example, the average message size), different retention and so on.
So please use carefully, adjust if necessary and better to execute SVT for Logging before run in production.

| Component            | CPU Requests | Memory Requests | CPU Limits | Memory Limits |
| -------------------- | ------------ | --------------- | ---------- | ------------- |
| Graylog              | `1000m`      | `2Gi`           | `3000m`    | `4Gi`         |
| FluentD              | `100m`       | `256Mi`         | `1000m`    | `1Gi`         |
| FluentBit Forwarder  | `100m`       | `128Mi`         | `500m`     | `512Mi`       |
| FluentBit Aggregator | `500m`       | `512Mi`         | `1000m`    | `2Gi`         |
| Cloud Events Reader  | `50m`        | `128Mi`         | `300m`     | `256Mi`       |

**Important!** In the case of deploy Graylog in the Cloud, you need to include in the calculation resources for
OpenSearch Cluster (recommended) or OpenSearch single instance that will deploy in the Cloud. Please refer
to the OpenSearch documentation to find OpenSearch hardware requirements.

[Back to TOC](#table-of-content)

### Large

Resources in this profile were calculated for the average load `> 10000` messages per second.

**Warning!** All resources below may require a tuning for your environment because different environment
can has different log types (for example, the average message size), different retention and so on.
So please use carefully, adjust if necessary and better to execute SVT for Logging before run in production.

| Component            | CPU Requests | Memory Requests | CPU Limits | Memory Limits |
| -------------------- | ------------ | --------------- | ---------- | ------------- |
| Graylog              | `2000m`      | `4Gi`           | `6000m`    | `8Gi`         |
| FluentD              | `500m`       | `512Mi`         | `1000m`    | `1536Mi`      |
| FluentBit Forwarder  | `100m`       | `256Mi`         | `1000m`    | `1024Mi`      |
| FluentBit Aggregator | `500m`       | `1Gi`           | `1000m`    | `2Gi`         |
| Cloud Events Reader  | `100m`       | `128Mi`         | `300m`     | `512Mi`       |

**Important!** In the case of deploy Graylog in the Cloud, you need to include in the calculation resources for
OpenSearch Cluster (recommended) or OpenSearch single instance that will deploy in the Cloud. Please refer
to the OpenSearch documentation to find OpenSearch hardware requirements.

[Back to TOC](#table-of-content)

## Logging on different OS

Logging agents (Fluentd, FluentBit) in the logging-service are configured to scrape logs from certain log files
from the node: system logs, audit logs, kube logs, containers logs.
But some OS have different locations for these files or may not contain them at all.

### System logs

Different OS have different locations for their system logs files. The most important system logs are global log journal
(`/var/log/syslog` by (r)syslogd, `/var/log/messages` by systemd, `/var/log/journal` by systemd-journald)
and auth logs (`/var/log/auth.log`, `/var/log/secure`).

The following table contains frequently used and recommended OS and paths to system logs for them:

<!-- markdownlint-disable line-length -->
| OS name                                | OS versions        | Global system logs                                    | Auth logs            |
| -------------------------------------- | ------------------ | ----------------------------------------------------- | -------------------- |
| Ubuntu                                 | 20.04.x, 22.04.x   | /var/log/syslog (/var/log/journal is available too)   | /var/log/auth.log    |
| Rocky Linux                            | 9.x                | /var/log/messages                                     | /var/log/secure      |
| CentOS                                 | 8.x                | /var/log/messages                                     | /var/log/secure      |
| RHEL                                   | 8.x                | /var/log/messages                                     | /var/log/secure      |
| Oracle Linux                           | 8.x                | /var/log/messages                                     | /var/log/secure      |
| Azure Linux (CBL-Mariner)              | 2.x                | /var/log/journal                                      | /var/log/journal     |
| Amazon Linux                           | 2.x                | /var/log/messages (/var/log/journal is available too) | /var/log/secure      |
| BottleRocket OS                        | 1.x                | /var/log/journal                                      | not present[^1]      |
| COS (Container-Optimized OS by Google) | 101, 105, 109, 113 | /var/log/journal (?)[^2]                              | /var/log/journal (?) |
<!-- markdownlint-enable line-length -->

 [^1]: BottleRocket is an OS created specifically for hosting containers, and it doesn't have a standard shell.
You can manage the BottleRocket OS only through a special in-built container with privileged rights,
so auth logs on the host would be useless for such concept.

 [^2]: **COS uses journald** as a main solution for system logs, and most likely the logs are located in
the default path for journald.

### Audit logs

Audit logs are managed by `auditd` daemon that is installed by default on the most OS, but there are several exceptions.

Audit logs by `auditd` are always located on `/var/log/audit/audit.log` by default.

The following table describes which OS have auditd by default:

<!-- markdownlint-disable line-length -->
| OS name         | Is auditd present by default                                                                                                                                           |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Ubuntu          | ✓ Yes                                                                                                                                                                  |
| Rocky Linux     | ✓ Yes                                                                                                                                                                  |
| CentOS          | ✓ Yes                                                                                                                                                                  |
| RHEL            | ✓ Yes                                                                                                                                                                  |
| Oracle Linux    | ✓ Yes                                                                                                                                                                  |
| Azure Linux     | ✗ No (auditd is not installed by default)                                                                                                                              |
| Amazon Linux    | ✓ Yes                                                                                                                                                                  |
| BottleRocket OS | ✗ No (auditd is not presented due the lack of the shell)                                                                                                               |
| COS             | ✗ No (disabled by default, [can be installed by using the special DaemonSet with auditd](https://cloud.google.com/kubernetes-engine/docs/how-to/linux-auditd-logging)) |
<!-- markdownlint-enable line-length -->

### Kubernetes and container logs

The location of Kubernetes and containers logs is independent of the OS the node is running on.

The location of Kubernetes logs depends on the Kubernetes version and the type of k8s cluster (pure Kubernetes,
OpenShift).

The location of containers logs depends on the container engine (docker, containerd, cri-o).

[Back to TOC](#table-of-content)

# Parameters

The configurable parameters are described as follows.

## Root

It\'s a common section that contains some generic parameters.

All parameters in the table below should be specified on the first root level:

```yaml
name: logging-service
containerRuntimeType: containerd
...
```

<!-- markdownlint-disable line-length -->
| Parameter                    | Type              | Mandatory | Default value                    | Description                                                                                                                                                  |
| ---------------------------- | ----------------- | --------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `name`                       | string            | no        | `logging-service`                | Name of new custom resource                                                                                                                                  |
| `namespace`                  | string            | no        |                                  | Cloud namespace to deploy logging service                                                                                                                    |
| `cloudURL`                   | string            | no        | `https://kubernetes.default.svc` | Address of Kubernetes APIServer                                                                                                                              |
| `osKind`                     | string            | no        | `centos`                         | Operating system kind on Cloud nodes. Possible values: `centos` / `rhel` / `oracle` / `ubuntu`                                                               |
| `ipv6`                       | boolean           | no        | `false`                          | Set to `true` for deploy to IPv6 environment.                                                                                                                |
| `containerRuntimeType`       | String            | no        | `docker`                         | Cloud containers runtime software. Possible values: `docker` / `cri-o` / `containerd`. In fact so far he differ docker and non-docker environments           |
| `createClusterAdminEntities` | boolean           | no        | `true`                           | Set to `true` in order to create logging service entities which requires cluster-admin privileges for creation. Your user must have cluster-admin privileges |
| `operatorImage`              | string            | no        | `-`                              | Docker image of Logging-operator                                                                                                                             |
| `skipMetricsService`         | boolean           | no        | `-`                              | Set to `true` to skip step of creation metrics Service and ServiceMonitor                                                                                    |
| `nodeSelectorKey`            | string            | no        | `-`                              | NodeSelector key                                                                                                                                             |
| `nodeSelectorValue`          | string            | no        | `-`                              | NodeSelector value                                                                                                                                           |
| `affinity`                           | [core/v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#podaffinityterm-v1-core)       | no        | `-`                                                                          | It specifies the pod\'s scheduling constraints                                                                                                                                                                                                                                                                          |
| `annotations`                | map               | no        | `{}`                             | Allows to specify list of additional annotations                                                                                                             |
| `labels`                     | map               | no        | `{}`                             | Allows to specify list of additional labels                                                                                                                  |
| `pprof.install`              | boolean           | no        | `true`                           | pprof enables collecting profiling data.                                                                                                                     |
| `pprof.containerPort`        | int               | no        | `9180`                           | prot for pprof container.                                                                                                                                    |
| `pprof.service.type`         | string            | no        | `ClusterIP`                      | pprof service type.                                                                                                                                          |
| `pprof.service.port`         | int               | no        | `9100`                           | pprof port which is used in service                                                                                                                          |
| `pprof.service.portName`     | string            | no        | `http`                           | pprof port name which is used in service.                                                                                                                    |
| `pprof.service.annotations`  | map[string]string | no        | `{}`                             | Allows to specify additional annotations in service                                                                                                          |
| `pprof.service.labels`       | map[string]string | no        | `{}`                             | Allows to specify list of additional labels in service                                                                                                       |
| `priorityClassName`          | string            | no        | `-`                              | Pod priority. Priority indicates the importance of a Pod relative to other Pods and prevents them from evicting.                                             |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
name: logging-service
namespace: logging
operatorImage: qubership-logging-operator:main

cloudURL: https://kubernetes.default.svc
osKind: ubuntu
ipv6: false
containerRuntimeType: containerd

createClusterAdminEntities: true

skipMetricsService: false
pprof:
  install: true
  containerPort: 9180
  service:
    type: ClusterIP
    port: 9180
    protName: pprof
    annotations: {}
    labels: {}

nodeSelectorKey: kubernetes.io/os
nodeSelectorValue: linux
```

[Back to TOC](#table-of-content)

## Graylog

The `graylog` section contains parameters to enable and configure Graylog deployment in the Cloud.

All parameters described below should be specified under a section `graylog` as the following:

```yaml
graylog:
  install: true
  #...
```

<!-- markdownlint-disable line-length -->
| Parameter                                  | Type                                                                                                                   | Mandatory | Default value                                                                   | Description                                                                                                                                                                                           |
| ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- | --------- | ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `install`                                  | boolean                                                                                                                | no        | `false`                                                                         | Enable GRaylog deployment in the Cloud                                                                                                                                                                |
| `dockerImage`                              | string                                                                                                                 | no        | `-`                                                                             | Image to use for Graylog deployment                                                                                                                                                                   |
| `initSetupImage`                           | string                                                                                                                 | no        | `-`                                                                             | Image for init container for Graylog                                                                                                                                                                  |
| `initContainerDockerImage`                 | string                                                                                                                 | no        | `-`                                                                             | Image to initialize plugins for Graylog                                                                                                                                                               |
| `initResources`                            | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{requests: {cpu: 50m, memory: 128Mi}, limits: {cpu: 100m, memory: 256Mi}}`     | The resources describe to compute resource requests and limits for single Pods                                                                                                                        |
| `mongoDBImage`                             | string                                                                                                                 | no        | `-`                                                                             | Image of MongoDB to use for Graylog deployment                                                                                                                                                        |
| `mongoUpgrade`                             | string                                                                                                                 | no        | `false`                                                                         | Activates automatic step-by-step upgrade of the MongoDB database. Can be used only for migration from Graylog 4 to 5                                                                                  |
| `mongoDBUpgrade.mongoDBImage40`            | string                                                                                                                 | no        | `-`                                                                             | Image of MongoDB 4.0 to use for Graylog deployment. Using to migration from MongoDB 3.6 to 5.x                                                                                                        |
| `mongoDBUpgrade.mongoDBImage42`            | string                                                                                                                 | no        | `-`                                                                             | Image of MongoDB 4.2 to use for Graylog deployment. Using to migration from MongoDB 3.6 to 5.x                                                                                                        |
| `mongoDBUpgrade.mongoDBImage44`            | string                                                                                                                 | no        | `-`                                                                             | Image of MongoDB 4.0 to use for Graylog deployment. Using to migration from MongoDB 3.6 to 5.x                                                                                                        |
| `mongoPersistentVolume`                    | string                                                                                                                 | no        | `-`                                                                             | MongoDB Persistence Volume (PV) name. Using to claim already created PVs                                                                                                                              |
| `mongoStorageClassName`                    | string                                                                                                                 | no        | `-`                                                                             | MongoDB Persistence Volume Claim (PVC) storage class name. Using in case of dynamical provisioning                                                                                                    |
| `mongoResources`                           | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{requests: {cpu: 500m, memory: 256Mi}, limits: {cpu: 500m, memory: 256Mi}}`    | The resources describe to compute resource requests and limits for single Pods                                                                                                                        |
| `annotations`                              | map                                                                                                                    | no        | `{}`                                                                            | Allows to specify list of additional annotations that will be add in Graylog pod                                                                                                                      |
| `labels`                                   | map                                                                                                                    | no        | `{}`                                                                            | Allows to specify list of additional labels that will be add in Graylog pod                                                                                                                           |
| `graylogResources`                         | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{requests: {cpu: 500m, memory: 1536Mi}, limits: {cpu: 1000m, memory: 2048Mi}}` | The resources describe to compute resource requests and limits for single Pods                                                                                                                        |
| `graylogPersistentVolume`                  | string                                                                                                                 | no        | `-`                                                                             | Graylog Persistence Volume (PV) name. Using to claim already created PVs                                                                                                                              |
| `graylogStorageClassName`                  | string                                                                                                                 | no        | `""`                                                                            | Graylog Persistence Volume Claim (PVC) storage class name. Using in case of dynamical provisioning                                                                                                    |
| `storageSize`                              | string                                                                                                                 | no        | `2Gi`                                                                           | Graylog Persistence Volume size. Using for `journald` cache                                                                                                                                           |
| `priorityClassName`                        | string                                                                                                                 | no        | `-`                                                                             | Pod priority. Priority indicates the importance of a Pod relative to other Pods and prevents them from evicting.                                                                                      |
| `nodeSelectorKey`                          | string                                                                                                                 | no        | `-`                                                                             | Key of `nodeSelector`                                                                                                                                                                                 |
| `nodeSelectorValue`                        | string                                                                                                                 | no        | `-`                                                                             | Value of `nodeSelector`                                                                                                                                                                               |
| `affinity`                           | [core/v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#podaffinityterm-v1-core)       | no        | `-`                                                                          | It specifies the pod\'s scheduling constraints                                                                                                                                                                                                                                                                          |
| `securityResources.install`                | boolean                                                                                                                | no        | `false`                                                                         | Enable creating security resources as PodSecurityPolicy, SecurityContextConstraints                                                                                                                   |
| `securityResources.name`                   | string                                                                                                                 | no        | `logging-graylog`                                                               | Set a name of PodSecurityPolicy, SecurityContextConstraints objects                                                                                                                                   |
| `logLevel`                                 | string                                                                                                                 | no        | `INFO`                                                                          | Set Graylog log level for self logs                                                                                                                                                                   |
| `indexReplicas`                            | integer                                                                                                                | no        | `1`                                                                             | The number of OpenSearch/ElasticSearch replicas used per index                                                                                                                                        |
| `indexShards`                              | integer                                                                                                                | no        | `5`                                                                             | The number of OpenSearch/ElasticSearch shards used per index                                                                                                                                          |
| `elasticsearchHost`                        | string                                                                                                                 | yes       | `-`                                                                             | OpenSearch/ElasticSearch host with schema, credentials and port. For example: `http://user:password@elasticsearch.elasticsearch.svc:9200`                                                             |
| `elasticsearchMaxTotalConnections`         | integer                                                                                                                | no        | `100`                                                                           | Maximum number of total connections to OpenSearch/ElasticSearch                                                                                                                                       |
| `elasticsearchMaxTotalConnectionsPerRoute` | integer                                                                                                                | no        | `100`                                                                           | Maximum number of connections per OpenSearch/ElasticSearch route (normally this means per OpenSearch/ElasticSearch server)                                                                            |
| `createIngress`                                  | boolean                                                                                                                | no        | `true`                                                                         | Enable/Disable creation of Ingress.                                                |
| `host`                                     | string                                                                                                                 | no        | `-`                                                                             | The Graylog host for Ingress and Route. For example: `https://graylog-service.kubernetes.test.org/`                                                                                                   |
| `ingressClassName`                         | string                                                                                                                 | no        | `-`                                                                             | Name of an IngressClass that will use in created Ingress                                                                                                                                              |
| `inputPort`                                | string                                                                                                                 | no        | `12201`                                                                         | Port of default Graylog Input                                                                                                                                                                         |
| `graylogSecretName`                        | string                                                                                                                 | no        | `graylog-secret`                                                                | The name of Kubernetes Secret that store Graylog super admin credentials and OpenSearch/ElasticSearch connection string                                                                               |
| `contentDeployPolicy`                      | string                                                                                                                 | no        | `only-create`                                                                   | Strategy of applying default and new configurations during Graylog provisioning. Available values: `only-create`, `force-update`                                                                      |
| `logsRotationSizeGb`                       | integer                                                                                                                | no        | `20`                                                                            | Set maximum size of logs in `All messages` stream                                                                                                                                                     |
| `maxNumberOfIndices`                       | integer                                                                                                                | no        | `20`                                                                            | Set maximum number of indexes                                                                                                                                                                         |
| `javaOpts`                                 | string                                                                                                                 | no        | `-`                                                                             | Graylog JVM options. For example: `-Xms1024m -Xmx1024m`                                                                                                                                               |
| `contentPacks`                             | [loggingservice/v11.ContentPackPathHTTPConfig](#contentpacks)                                                          | no        | `{}`                                                                            | Links to Graylog\'s Content Packs.                                                                                                                                                                    |
| `contentPackPaths`                         | string                                                                                                                 | no        | `-`                                                                             | Links to Graylog\'s Content Packs. To specify some Context Packs use comma (`,`) as a separator                                                                                                       |
| `customPluginsPaths`                       | string                                                                                                                 | no        | `-`                                                                             | Graylog plugin path                                                                                                                                                                                   |
| `startupTimeout`                           | integer                                                                                                                | no        | `10`                                                                            | Time which operator waits for a Graylog pod to start, in minutes                                                                                                                                      |
| `ringSize`                                 | integer                                                                                                                | no        | `262144`                                                                        | Total size of ring buffers. Must be a power of 2 (512, 1024, 2048, ...)                                                                                                                               |
| `inputbufferRingSize`                      | integer                                                                                                                | no        | `131072`                                                                        | Size of input ring buffers. Must be a power of 2 (512, 1024, 2048, ...)                                                                                                                               |
| `inputbufferProcessors`                    | integer                                                                                                                | no        | `3`                                                                             | The number of cores/processes to process Input Buffer                                                                                                                                                 |
| `processbufferProcessors`                  | integer                                                                                                                | no        | `6`                                                                             | The number of cores/processes to process Processing Buffer                                                                                                                                            |
| `outputbufferProcessors`                   | integer                                                                                                                | no        | `6`                                                                             | The number of cores/processes to process Output Buffer                                                                                                                                                |
| `outputbufferProcessorThreadsMaxPoolSize`  | integer                                                                                                                | no        | `33`                                                                            | The maximum number of threads to allow in the pool                                                                                                                                                    |
| `outputBatchSize`                          | integer                                                                                                                | no        | `1000`                                                                          | Batch size for the OpenSearch/ElasticSearch output. This is the **maximum** number of messages the OpenSearch/Elasticsearch output module will get at once and write to Elasticsearch in a batch call |
| `openSearch`                               | [loggingservice/v11.OpenSearch](#opensearch)                                                                           | no        | `{}`                                                                            | Configuration of OpenSearch.                                                                                                                                                                          |
| `streams`                                  | [loggingservice/v11.GraylogStream](#graylog-streams)                                                                   | no        | `{}`                                                                            | Configuration of Graylog Streams. System and audit logs will be created by default if the section is empty.                                                                                           |
| `tls`                                      | [loggingservice/v11.GraylogTLS](#graylog-tls)                                                                          | no        | `{}`                                                                            | Configuration Graylog HTTPS/TLS for WebUI and default Inputs                                                                                                                                          |
| `authProxy`                                | [loggingservice/v11.GraylogAuthProxy](#graylog-auth-proxy)                                                             | no        | `{}`                                                                            | Configuration Graylog auth-proxy that allow use LDAP integration with Graylog groups                                                                                                                  |
| `user`                                     | string                                                                                                                 | no        | `admin`                                                                         | Username of Graylog super-admin user. Can't be empty                                                                                                                                                  |
| `password`                                 | string                                                                                                                 | no        | `admin`                                                                         | Password of Graylog super-admin user. Can't be empty                                                                                                                                                  |
| `s3Archive`                                | boolean                                                                                                                | no        | `false`                                                                         | Flag to using S3 storage in the `graylog-archiving-plugin`                                                                                                                                            |
| `awsAccessKey`                             | string                                                                                                                 | no        | `""`                                                                            | AccessKey to using S3 storage in the `graylog-archiving-plugin`                                                                                                                                       |
| `awsSecretKey`                             | string                                                                                                                 | no        | `""`                                                                            | AccessKey to using S3 storage in the `graylog-archiving-plugin`                                                                                                                                       |
| `pathRepo`                                 | string                                                                                                                 | no        | `/usr/share/opensearch/snapshots/graylog/`                                      | Path in OpenSearch/ElasticSearch to create data snapshots. These data next will upload to S3. Using in `graylog-archiving-plugin`                                                                     |
| `serviceMonitor.scrapeInterval`            | string                                                                                                                 | no        | `30s`                                                                           | Set metrics scrape interval                                                                                                                                                                           |
| `serviceMonitor.scrapeTimeout`             | string                                                                                                                 | no        | `10s`                                                                           | Set metrics scrape timeout                                                                                                                                                                            |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
graylog:
  install: true
  dockerImage: graylog/graylog:5.2.7
  createIngress: true

  # Init image settings
  initSetupImage: alpine:3.18
  initContainerDockerImage: graylog-plugins-init-container:main
  initResources:
    requests:
      cpu: 50m
      memory: 128Mi
    limits:
      cpu: 100m
      memory: 256Mi

  # MongoDB sidecar settings
  mongoDBImage: mongo:5.0.19
  mongoUpgrade: true
  mongoDBUpgrade:
    mongoDBImage40: mongo:4.0.28
    mongoDBImage42: mongo:4.2.22
    mongoDBImage44: mongo:4.4.17
  mongoPersistentVolume: pv-mongodb
  mongoStorageClassName: cinder
  mongoResources:
    requests:
      cpu: 500m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 256Mi

  # Graylog deployment settings
  annotations:
    custom/annotation: value
  labels:
    app.kubernetes.io/part-of: logging
  graylogResources:
    requests:
      cpu: 500m
      memory: 1536Mi
    limits:
      cpu: 1000m
      memory: 2048Mi
  graylogPersistentVolume: pv-graylog
  graylogStorageClassName: nginx
  storageSize: 5Gi
  graylogSecretName: graylog-secret
  priorityClassName: system-cluster-critical
  startupTimeout: 10
  nodeSelectorKey: kubernetes.io/os
  nodeSelectorValue: linux
  host: https://graylog-service.kubernetes.test.org/
  ingressClassName: nginx
  securityResources:
    install: true
    name: logging-graylog

  # Graylog settings
  logLevel: INFO
  indexReplicas: 1
  indexShards: 5
  elasticsearchHost: http://user:password@elasticsearch.elasticsearch.svc:9200
  elasticsearchMaxTotalConnections: 100
  elasticsearchMaxTotalConnectionsPerRoute: 100
  inputPort: 12201
  contentDeployPolicy: force-update
  logsRotationSizeGb: 20
  maxNumberOfIndices: 20
  javaOpts: "-Xms1024m -Xmx2048"
  contentPackPaths: http://nexus.test.org/raw/custom-content-pack.zip
  customPluginsPaths: /path/to/plugins

  # Graylog performance settings and buffer sizes
  ringSize: 262144
  inputbufferRingSize: 131072
  inputbufferProcessors: 3
  processbufferProcessors: 6
  outputbufferProcessors: 6
  outputbufferProcessorThreadsMaxPoolSize: 33
  outputBatchSize: 1000

  # Graylog super admin credentials
  user: admin
  password: admin

  # Graylog archiving plugins settings. To store archives in S3
  s3Archive: true
  awsAccessKey: s3_access_key
  awsSecretKey: s3_secret_key
  pathRepo: /usr/share/opensearch/snapshots/graylog/

  serviceMonitor:
    scrapeInterval: 30s
    scrapeTimeout: 10s

  streams:
    ...
  tls:
    ...
  authProxy:
    ...
```

[Back to TOC](#table-of-content)

### Graylog TLS

The `graylog.tls` section contains parameters to enable TLS for Graylog WebUI and default Inputs.
This section contains two subsections `http` for WebUI and `input` for Graylog\'s Inputs.

All parameters for Graylog WebUI described below should be specified under a section `graylog.tls.http` as the following:

```yaml
graylog:
  tls:
    http:
      enabled: true
      #...
```

<!-- markdownlint-disable line-length -->
| Parameter                         | Type    | Mandatory | Default value | Description                                                                                                                                                                                                          |
| --------------------------------- | ------- | --------- | ------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enabled`                         | boolean | no        | `false`       | Enable TLS for HTTP interface. If this parameter is `true`, each connection to and from the Graylog server except inputs will be secured by TLS, including API calls of the server to itself                         |
| `cacerts`                         | string  | no        | `-`           | Contains a name of Secret with CA certificates for a custom CA store. If present, all certificates from the Secret will be added to the Java keystore. The keystore can be used for TLS in the custom inputs as well |
| `keyFilePassword`                 | string  | no        | `-`           | The password to unlock the private key used for securing the HTTP interface                                                                                                                                          |
| `cert.secretKey`                  | string  | no        | `-`           | Name of Secret with the certificate                                                                                                                                                                                  |
| `cert.secretName`                 | string  | no        | `-`           | Key (filename) in the Secret with the certificate                                                                                                                                                                    |
| `key.secretKey`                   | string  | no        | `-`           | Name of Secret with the private key for the certificate                                                                                                                                                              |
| `key.secretName`                  | string  | no        | `-`           | Key (filename) in the Secret with the private key for the certificate                                                                                                                                                |
| `generateCerts.enabled`           | boolean | no        | `-`           | Enabling integration with `cert-manager` to generate certificates. Mutually exclusive with `cert` and `key` parameters                                                                                               |
| `generateCerts.secretName`        | string  | no        | `-`           | Secret name with certificates that will generate by `cert-manager`                                                                                                                                                   |
| `generateCerts.clusterIssuerName` | string  | no        | `-`           | Issuer that will use to generate certificates                                                                                                                                                                        |
| `generateCerts.duration`          | string  | no        | `-`           | Set certificates validity period                                                                                                                                                                                     |
| `generateCerts.renewBefore`       | string  | no        | `-`           | Sets the number of days before the certificates expiration day for which they will be reissued                                                                                                                       |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
graylog:
  tls:
    http:
      enabled: true
      cacerts: secret-ca
      keyFilePassword: changeit
      cert:
        secretName: graylog-http-tls-assets-0
        secretkey: graylog-http.crt
      key:
        secretName: graylog-http-tls-assets-0
        secretkey: graylog-http.key
      generateCerts:
        enabled: true
        secretName: graylog-http-cert-manager-tls
        clusterIssuerName: ""
        duration: 365
        renewBefore: 15
```

All parameters for Graylog\'s Inputs described below should be specified under a section `graylog.tls.input`
as the following:

```yaml
graylog:
  tls:
    input:
      enabled: true
      #...
```

<!-- markdownlint-disable line-length -->
| Parameter                         | Type    | Mandatory | Default value | Description                                                                                                            |
| --------------------------------- | ------- | --------- | ------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `enabled`                         | boolean | no        | `false`       | Enabling TLS for out-of-box GELF input managed by the operator                                                         |
| `keyFilePassword`                 | string  | no        | `-`           | The password to unlock the private key used for securing Graylog input                                                 |
| `ca.secretName`                   | string  | no        | `-`           | Name of Kubernetes Secret with CA certificate. Mutually exclusive with `generateCerts` section                         |
| `ca.secretKey`                    | string  | no        | `-`           | Key (filename) in the Secret with CA certificate                                                                       |
| `cert.secretKey`                  | string  | no        | `-`           | Name of Secret with the certificate. Mutually exclusive with `generateCerts` parameters                                |
| `cert.secretName`                 | string  | no        | `-`           | Key (filename) in the Secret with the certificate                                                                      |
| `key.secretKey`                   | string  | no        | `-`           | Name of Secret with the private key for the certificate. Mutually exclusive with `generateCerts` parameters            |
| `key.secretName`                  | string  | no        | `-`           | Key (filename) in the Secret with the private key for the certificate                                                  |
| `generateCerts.enabled`           | boolean | no        | `-`           | Enabling integration with `cert-manager` to generate certificates. Mutually exclusive with `cert` and `key` parameters |
| `generateCerts.secretName`        | string  | no        | `-`           | Secret name with certificates that will generate by `cert-manager`                                                     |
| `generateCerts.clusterIssuerName` | string  | no        | `-`           | Issuer that will use to generate certificates                                                                          |
| `generateCerts.duration`          | string  | no        | `-`           | Set certificates validity period                                                                                       |
| `generateCerts.renewBefore`       | string  | no        | `-`           | Sets the number of days before the certificates expiration day for which they will be reissued                         |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
graylog:
  tls:
    input:
      enabled: true
      keyFilePassword: changeit
      # Certificates from Kubernetes Secrets
      ca:
        secretName: graylog-input-tls-assets-0
        secretKey: ca.crt
      cert:
        secretName: graylog-input-tls-assets-0
        secretKey: graylog-input.crt
      key:
        secretName: graylog-input-tls-assets-0
        secretKey: graylog-input.key
      
      # Integration with cert-manager
      generateCerts:
        enabled: true
        secretName: graylog-input-cert-manager-tls
        clusterIssuerName: ""
        duration: 365
        renewBefore: 15
```

[Back to TOC](#table-of-content)

### OpenSearch

The `opensearch` section contains OpenSearch HTTP parameters.

All parameters for OpenSearch described below should be specified under a section `graylog.openSearch` as the following:

```yaml
graylog:
  openSearch:
    http:
      credentials:
        ...
      tlsConfig:
        ...
    url:
```

<!-- markdownlint-disable line-length -->
| Parameter                           | Type               | Mandatory | Default value | Description                                                                                         |
| ----------------------------------- | ------------------ | --------- | ------------- | --------------------------------------------------------------------------------------------------- |
| `http.credentials.username`         | *SecretKeySelector | no        | `-`           | The secret that contains the username for Basic authentication                                      |
| `http.credentials.password`         | *SecretKeySelector | no        | `-`           | The secret that contains the password for Basic authentication                                      |
| `http.tlsConfig.ca`                 | *SecretKeySelector | no        | `-`           | Secret name and key where Certificate Authority is stored.                                          |
| `http.tlsConfig.cert`               | *SecretKeySelector | no        | `-`           | Secret name and key where Certificate signing request is stored.                                    |
| `http.tlsConfig.key`                | *SecretKeySelector | no        | `-`           | Secret name and key where private key is stored.                                                    |
| `http.tlsConfig.insecureSkipVerify` | boolean            | no        | `-`           | InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. |
| `url`                               | string             | no        | `-`           | OpenSearch host                                                                                     |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
graylog:
  openSearch:
    http:
      credentials:
        username: 
          name: openSearch-credentials-secret
          key: httpRequestUsername
        password: 
          name: openSearch-credentials-secret
          key: httpRequestPassword
      tlsConfig:
        ca:
          name: secret-certificate
          key: cert-ca.pem
        cert:
          name: secret-certificate
          key: cert.crt
        key:
          name: secret-certificate
          key: cert.key
        insecureSkipVerify: false
    url: openSearch host 
```

[Back to TOC](#table-of-content)

### ContentPacks

The `contentpacks` section contains graylog content pack parameters.

All parameters for OpenSearch described below should be specified under a section `graylog.contentPacks` as the following:

```yaml
graylog:
  contentPacks:
    - http:
        credentials:
          ...
        tlsConfig:
          ...
      url:
    - http:
        credentials:
           ...
        tlsConfig:
           ...
      url:
```

<!-- markdownlint-disable line-length -->
| Parameter                           | Type               | Mandatory | Default value | Description                                                                                         |
| ----------------------------------- | ------------------ | --------- | ------------- | --------------------------------------------------------------------------------------------------- |
| `http.credentials.username`         | *SecretKeySelector | no        | `-`           | The secret that contains the username for Basic authentication                                      |
| `http.credentials.password`         | *SecretKeySelector | no        | `-`           | The secret that contains the password for Basic authentication                                      |
| `http.tlsConfig.ca`                 | *SecretKeySelector | no        | `-`           | Secret name and key where Certificate Authority is stored.                                          |
| `http.tlsConfig.cert`               | *SecretKeySelector | no        | `-`           | Secret name and key where Certificate signing request is stored.                                    |
| `http.tlsConfig.key`                | *SecretKeySelector | no        | `-`           | Secret name and key where private key is stored.                                                    |
| `http.tlsConfig.insecureSkipVerify` | boolean            | no        | `-`           | InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. |
| `url`                               | string             | no        | `-`           | Content pack url                                                                                    |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It's just an example of a parameter's format, not a recommended parameter.

```yaml
graylog:
  contentPacks:
    - http:
        credentials:
          username: 
            name: contentPack-credentials-secret
            key: httpRequestUsername
          password: 
            name: contentPack-credentials-secret
            key: httpRequestPassword
        tlsConfig:
          ca:
            name: secret-certificate
            key: cert-ca.pem
          cert:
            name: secret-certificate
            key: cert.crt
          key:
            name: secret-certificate
            key: cert.key
          insecureSkipVerify: false
      url: contentPack url
    - http:
        ...
```

[Back to TOC](#table-of-content)

### Graylog Streams

The `graylog.streams` section contains parameters to enable, disable or modify the retention strategy of default
Graylog's Streams.

All parameters described below should be specified under a section `graylog.streams` as the following:

```yaml
graylog:
  streams:
    - name: "System logs"
      #...
```

<!-- markdownlint-disable line-length -->
| Parameter          | Type    | Mandatory | Default value | Description                                                                                                                              |
| ------------------ | ------- | --------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `install`          | boolean | no        | `-`           | Enable or disable stream                                                                                                                 |
| `name`             | string  | no        | `-`           | The title of a Graylog's Stream. Available logs are `System logs`, `Audit logs`, `Access logs`, `Integration logs` and `Bill Cycle logs` |
| `rotationStrategy` | string  | no        | `sizeBased`   | Sets rotation strategy to IndexSet of the Stream. Available values: `sizeBased`, `timeBased`                                             |
| `rotationPeriod`   | string  | no        | `-`           | Sets rotation period to index set of the stream if `rotationStrategy` is `sizeBased`. The parameter must be set as ISO 8601 Duration     |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It's just an example of a parameter's format, not a recommended parameter.

```yaml
graylog:
  streams:
    - name: "Audit logs"
      install: true
      rotationStrategy: "timeBased"
      rotationPeriod: "P1M"
    - name: "Integration logs"
      install: false
      rotationStrategy: "timeBased"
      rotationPeriod: "P1M"
    - name: "Access logs"
      install: false
      rotationStrategy: "timeBased"
      rotationPeriod: "P1M"
    - name: "Nginx logs"
      install: false
      rotationStrategy: "timeBased"
      rotationPeriod: "P1M"
    - name: "Bill Cycle logs"
      install: false
      rotationStrategy: "timeBased"
      rotationPeriod: "P1M15D"
```

[Back to TOC](#table-of-content)

### Graylog Auth Proxy

The `graylog.authProxy` section contains parameters to enable and configure graylog-auth-proxy.
This is a proxy that allows authentication and authorization of users for the Graylog server using third-party databases
(for example, Active Directory) or OAuth authorization service (for example, Keycloak).

All parameters described below should be specified under a section `graylog.authProxy` as the following:

```yaml
graylog:
  authProxy:
    install: true
    #...
```

<!-- markdownlint-disable line-length -->
| Parameter              | Type                                                                                                                   | Mandatory | Default value                                                                      | Description                                                                                                   |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------- | --------- | ---------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `install`              | boolean                                                                                                                | no        | `false`                                                                            | Enable `graylog-auth-proxy` deployment                                                                        |
| `logLevel`             | string                                                                                                                 | no        | `INFO`                                                                             | Logging level. Allowed values: `DEBUG`, `INFO`, `WARNING`, `ERROR`, `CRITICAL`                                |
| `image`                | string                                                                                                                 | yes       | `-`                                                                                | Image of `graylog-auth-proxy`                                                                                 |
| `preCreatedUsers`      | string                                                                                                                 | no        | `admin,auditViewer,operator,telegraf_operator,graylog-sidecar,graylog_api_th_user` | Comma separated pre-created users in Graylog for which you do not need to rotate passwords                    |
| `rotationPassInterval` | integer                                                                                                                | no        | `3`                                                                                | Interval in days between password rotation for non-pre-created users                                          |
| `roleMapping`          | string                                                                                                                 | no        | `'[]'`                                                                             | Filter for mapping Graylog roles between LDAP and Graylog users by memberOf field                             |
| `streamMapping`        | string                                                                                                                 | no        | `""`                                                                               | Filter for sharing Graylog streams between LDAP and Graylog users by memberOf field                           |
| `resources`            | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{}`                                                                               | Resources describe to compute resource requests and limits for `graylog-auth-proxy` container                 |
| `requestsTimeout`      | float                                                                                                                  | no        | 30                                                                                 | A global timeout parameter affects requests to LDAP server, OAuth server and Graylog server                   |
| `authType`             | string                                                                                                                 | no        | `ldap`                                                                             | Defines which type of authentication protocol will be chosen (LDAP or OAuth 2.0). Allowed values: ldap, oauth |
| `ldap`                 | [loggingservice/v11.GraylogAuthProxyLDAP](#graylog-auth-proxy-ldap)                                                    | no        | `-`                                                                                | Configuration for LDAP or AD connection                                                                       |
| `oauth`                | [loggingservice/v11.GraylogAuthProxyOAuth](#graylog-auth-proxy-oauth)                                                  | no        | `-`                                                                                | Configuration for OAuth 2.0 connection                                                                        |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It's just an example of a parameter's format, not a recommended parameter.

```yaml
graylog:
  authProxy:
    install: true
    image: graylog-auth-proxy:main
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 200m
        memory: 256Mi

    preCreatedUsers: 3
    roleMapping: '["Reader"]'
    streamMapping: ''
    requestsTimeout: 30
    
    authType: 'ldap'

    ldap:
      ...
```

[Back to TOC](#table-of-content)

#### Graylog Auth Proxy LDAP

The `graylog.authProxy.ldap` section contains parameters to configure LDAP provider for `graylog-auth-proxy`.

All parameters described below should be specified under a section `graylog.authProxy.ldap` as the following:

```yaml
graylog:
  authProxy:
    authType: "ldap"
    ldap:
      ...
```

<!-- markdownlint-disable line-length -->
| Parameter                 | Type    | Mandatory | Default value               | Description                                                                                                     |
| ------------------------- | ------- | --------- | --------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `url`                     | string  | yes       | `-`                         | LDAP host to query users and their data                                                                         |
| `startTls`                | boolean | no        | `false`                     | Enable establishing a `STARTTLS` protected session                                                              |
| `overSsl`                 | boolean | no        | `false`                     | Establish an LDAP session over SSL                                                                              |
| `skipVerify`              | boolean | no        | `false`                     | Allow skipping verification of the LDAP server's certificate                                                    |
| `ca.secretName`           | string  | no        | `-`                         | Name of Kubernetes Secret with CA certificate                                                                   |
| `ca.secretKey`            | string  | no        | `-`                         | Key (filename) in the Secret with CA certificate                                                                |
| `cert.secretName`         | string  | no        | `-`                         | Name of Kubernetes Secret with client certificate                                                               |
| `cert.secretKey`          | string  | no        | `-`                         | Key (filename) in the Secret with client certificate                                                            |
| `key.secretName`          | string  | no        | `-`                         | Name of Kubernetes Secret with the private key for the client certificate                                       |
| `key.secretKey`           | string  | no        | `-`                         | Key (filename) in the Secret with the private key for the client certificate                                    |
| `disableReferrals`        | boolean | no        | `false`                     | Sets `ldap.OPT_REFERRALS` to zero                                                                               |
| `searchFilter`            | string  | no        | `(cn=%(username)s)`         | LDAP filter for binding users                                                                                   |
| `baseDN`                  | string  | yes       | `-`                         | LDAP base DN                                                                                                    |
| `bindDN`                  | string  | yes       | `-`                         | LDAP bind DN                                                                                                    |
| `bindPassword`            | string  | yes       | `-`                         | LDAP password for the bind DN. Mutually exclusive with `bindPasswordSecret` parameter                           |
| `bindPasswordSecret.name` | string  | no        | `graylog-auth-proxy-secret` | Kubernetes Secret name with LDAP password for the bind DN. Mutually exclusive with `bindPassword` parameter     |
| `bindPasswordSecret.key`  | string  | no        | `bindPassword`              | Field in Kubernetes Secret with LDAP password for the bind DN. Mutually exclusive with `bindPassword` parameter |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It's just an example of a parameter's format, not a recommended parameter.

```yaml
graylog:
  authProxy:
    authType: "ldap"
    ldap:
      url: ldaps://openldap.test.org:636
      startTls: false
      overSsl: true
      skipVerify: false
      ca:
        secretName: graylog-auth-proxy-ldap-ca
        secretKey: ca.crt
      disableReferrals: false
      searchFilter: (cn=%(username)s)

      baseDN: cn=admin,dc=example,dc=com
      bindDN: dc=example,dc=com
      bindPassword: very_secret_password
      bindPasswordSecret:
        name: graylog-auth-proxy-credentials
        key: password
```

[Back to TOC](#table-of-content)

#### Graylog Auth Proxy OAuth

The `graylog.authProxy.oauth` section contains parameters to configure OAuth provider for `graylog-auth-proxy`.

All parameters described below should be specified under a section `graylog.authProxy.oauth` as the following:

```yaml
graylog:
  authProxy:
    authType: "oauth"
    oauth:
      ...
```

<!-- markdownlint-disable line-length -->
| Parameter                      | Type    | Mandatory | Default value               | Description                                                                                                                                                                                                                                                        |
| ------------------------------ | ------- | --------- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `host`                         | string  | yes       | `-`                         | OAuth2 authorization server host                                                                                                                                                                                                                                   |
| `authorizationPath`            | string  | yes       | `-`                         | This path will be used to build URL for redirection to OAuth2 authorization server login page                                                                                                                                                                      |
| `tokenPath`                    | string  | yes       | `-`                         | This path will be used to build URL for getting access token from OAuth2 authorization server                                                                                                                                                                      |
| `userinfoPath`                 | string  | yes       | `-`                         | This path will be used to build URL for getting information about current user from OAuth2 authorization server to get username and entities (roles, groups, etc.) for Graylog roles and streams mapping                                                           |
| `redirectUri`                  | string  | no        | `-`                         | URI to redirect after successful logging in on the OAuth2 authorization server side. Uses "graylog.host" (the same host as in the Graylog Ingress) with /code path by default. Please make sure that your OAuth server allows using this URI as a redirection URI. |
| `clientID`                     | string  | yes       | `-`                         | OAuth2 Client ID for the proxy                                                                                                                                                                                                                                     |
| `clientSecret`                 | string  | yes       | `-`                         | OAuth2 Client Secret for the proxy. Will be stored in the secret with .oauth.clientCredentialsSecret.name at key specified in the .oauth.clientCredentialsSecret.key.                                                                                              |
| `scopes`                       | string  | no        | `openid profile roles`      | OAuth2 scopes for the proxy separated by spaces. Configured for Keycloak server by default                                                                                                                                                                         |
| `userJsonpath`                 | string  | no        | `preferred_username`        | JSONPath (by jsonpath-ng) for taking username from the JSON returned from OAuth2 server by using userinfo path. Configured for Keycloak server by default                                                                                                          |
| `rolesJsonpath`                | string  | no        | `realm_access.roles[*]`     | JSONPath (by jsonpath-ng) for taking information about entities (roles, groups, etc.) for Graylog roles and streams mapping from the JSON returned from OAuth2 server by using userinfo path. Configured for Keycloak server by default                            |
| `skipVerify`                   | boolean | no        | `false`                     | Allow skipping verification of the OAuth server's certificate                                                                                                                                                                                                      |
| `ca.secretName`                | string  | no        | `-`                         | Name of Kubernetes Secret with CA certificate                                                                                                                                                                                                                      |
| `ca.secretKey`                 | string  | no        | `-`                         | Key (filename) in the Secret with CA certificate                                                                                                                                                                                                                   |
| `cert.secretName`              | string  | no        | `-`                         | Name of Kubernetes Secret with client certificate                                                                                                                                                                                                                  |
| `cert.secretKey`               | string  | no        | `-`                         | Key (filename) in the Secret with client certificate                                                                                                                                                                                                               |
| `key.secretName`               | string  | no        | `-`                         | Name of Kubernetes Secret with the private key for the client certificate                                                                                                                                                                                          |
| `key.secretKey`                | string  | no        | `-`                         | Key (filename) in the Secret with the private key for the client certificate                                                                                                                                                                                       |
| `clientCredentialsSecret.name` | string  | no        | `graylog-auth-proxy-secret` | Kubernetes Secret name with OAuth client secret. Mutually exclusive with `clientSecret` parameter                                                                                                                                                                  |
| `clientCredentialsSecret.key`  | string  | no        | `clientSecret`              | Field in Kubernetes Secret with OAuth client secret. Mutually exclusive with `clientSecret` parameter                                                                                                                                                              |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It's just an example of a parameter's format, not a recommended parameter.

```yaml
graylog:
  authProxy:
    authType: "oauth"
    oauth:
      host: https://keycloak.server.com
      authorizationPath: /realms/test-realm/protocol/openid-connect/auth
      tokenPath: /realms/test-realm/protocol/openid-connect/token
      userinfoPath: /realms/test-realm/protocol/openid-connect/userinfo
      skipVerify: false
      ca:
        secretName: graylog-auth-proxy-oauth-ca
        secretKey: ca.crt
      
      clientID: graylog-auth-proxy
      clientSecret: <client-secret>
      scopes: "openid profile roles"
      userJsonpath: "preferred_username"
      rolesJsonpath: "realm_access.roles[*]"
      clientCredentialsSecret:
        name: graylog-auth-proxy-secret
        key: clientSecret
```

[Back to TOC](#table-of-content)

## FluentBit

The `fluentbit` section contains parameters to enable and configure FluentBit logging agent.

All parameters described below should be specified under a section `fluentbit` as the following:

```yaml
fluentbit:
  install: true
  #...
```

<!-- markdownlint-disable line-length -->
| Parameter                     | Type                                                                                                                              | Mandatory | Default value                                                                              | Description                                                                                                                                                         |
|-------------------------------|-----------------------------------------------------------------------------------------------------------------------------------| --------- | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `install`                     | boolean                                                                                                                           | no        | `false`                                                                                    | Flag for installation `logging-fluentbit`                                                                                                                           |
| `dockerImage`                 | string                                                                                                                            | no        | `-`                                                                                        | Docker image of FluentBit                                                                                                                                           |
| `configmapReload.dockerImage` | string                                                                                                                            | no        | `-`                                                                                        | Docker image of configmap_reload for FluentBit                                                        |
| `configmapReload.resources`   | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)            | no        | `{requests: {cpu: 10m, memory: 10Mi}, limits {cpu: 50m, memory: 50Mi}}`                    | The resources describe to compute resource requests and limits for single Pods                                  |
| `nodeSelectorKey`             | string                                                                                                                            | no        | `-`                                                                                        | NodeSelector key, can be multiple by OR condition, separated by comma, usually `role`                                                                               |
| `nodeSelectorValue`           | string                                                                                                                            | no        | `-`                                                                                        | NodeSelector value, can be multiple by OR condition, separated by comma, usually `compute`                                                                          |
| `tolerations`                 | [core/v1.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#toleration-v1-core)                     | no        | `[]`                                                                                       | List of tolerations applied to FluentBit Pods                                                                                                                       |
| `affinity`                           | [core/v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#podaffinityterm-v1-core)       | no        | `-`                                                                          | It specifies the pod\'s scheduling constraints                                                                                                                                                                                                                                                                          |
| `resources`                   | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)            | no        | `{requests: {cpu: 50m, memory: 128Mi}, limits {cpu: 200m, memory: 512Mi}}`                 | The resources describe to compute resource requests and limits for single Pods                                                                                      |
| `graylogOutput`               | boolean                                                                                                                           | no        | `true`                                                                                     | Flag for using Graylog output                                                                                                                                       |
| `graylogHost`                 | string                                                                                                                            | no        | `-`                                                                                        | Graylog host                                                                                                                                                        |
| `graylogPort`                 | integer                                                                                                                           | no        | `12201`                                                                                    | Graylog port                                                                                                                                                        |
| `graylogProtocol`             | string                                                                                                                            | no        | `tcp`                                                                                      | The Graylog protocol. Available values: `tcp`, `udp`                                                                                                                |
| `extraFields`                 | map[string]string                                                                                                                 | no        | `-`                                                                                        | Adds additional custom fields/labels to every log message by using filter based on record_modifier plugin.                                                          |
| `customInputConf`             | string                                                                                                                            | no        | `-`                                                                                        | Custom input configuration                                                                                                                                          |
| `customFilterConf`            | string                                                                                                                            | no        | `-`                                                                                        | Custom filter configuration                                                                                                                                         |
| `customLuaScriptConf`         | map[string]string                                                                                                                 | no        | `-`                                                                                        | Set of custom Lua scripts                                                                                                                                           |
| `customOutputConf`            | string                                                                                                                            | no        | `-`                                                                                        | Custom output configuration                                                                                                                                         |
| `multilineFirstLineRegexp`    | string                                                                                                                            | no        | `/^(\\[\\d{4}\\-\\d{2}\\-\\d{2}\|\\{\\\"\|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/`   | Custom regexp for the first line of multiline filter                                                                                                                |
| `multilineOtherLinesRegexp`   | string                                                                                                                            | no        | `/^(?!\\[\\d{4}\\-\\d{2}\\-\\d{2}\|\\{\\\"\|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/` | Custom regexp for the other lines of multiline filter                                                                                                               |
| `billCycleConf`               | boolean                                                                                                                           | no        | `false`                                                                                    | Filter for bil-cycle-logs stream                                                                                                                                    |
| `securityContextPrivileged`   | boolean                                                                                                                           | no        | `false`                                                                                    | Allows specifying securityContext.privileged for fluentd container                                                                                                  |
| `systemLogging`               | boolean                                                                                                                           | no        | `false`                                                                                    | Allows to enable collecting of system logs                                                                                                                          |
| `systemLogType`               | string                                                                                                                            | no        | `varlogmessages`                                                                           | Type of system logs to collect. Available values: `varlogmessages`, `varlogsyslog` and `systemd`                                                                    |
| `systemAuditLogging`          | boolean                                                                                                                           | no        | `true`                                                                                     | Enable input for system audit logs from `/var/log/audit/audit.log`.                                                                                                 |
| `kubeAuditLogging`            | boolean                                                                                                                           | no        | `true`                                                                                     | Enable input for Kubernetes audit logs from `/var/log/kubernetes/kube-apiserver-audit.log` and `/var/log/kubernetes/audit.log`.                                     |
| `kubeApiserverAuditLogging`   | boolean                                                                                                                           | no        | `true`                                                                                     | Enable input for Kubernetes APIServer audit logs from `/var/log/kube-apiserver/audit.log` for Kubernetes and `/var/log/openshift-apiserver/audit.log` for OpenShift. |
| `containerLogging`            | boolean                                                                                                                           | no        | `true`                                                                                     | Enable input for container logs from `/var/logs/containers` for Docker or `/var/log/pods` for other engines.                                                        |
| `totalLimitSize`              | string                                                                                                                            | no        | `1024M`                                                                                    | The size limitation of output buffer                                                                                                                                |
| `memBufLimit`                 | string                                                                                                                            | no        | `1024M`                                                                                    | Limit of allowed storage for chucks of logs before sending                                                                                                          |
| `additionalVolumes`           | [core/v1.PersistentVolumeSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#persistentvolumespec-v1-core) | no        | `{}`                                                                                       | Additional volumes for FluentBit                                                                                                                                    |
| `additionalVolumeMounts`      | [core/v1.VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#volumemount-v1-core)                   | no        | `{}`                                                                                       | Allows volume-mounts for FluentBit                                                                                                                                  |
| `securityResources.install`   | boolean                                                                                                                           | no        | `false`                                                                                    | Enable creating security resources as `PodSecurityPolicy`, `SecurityContextConstraints`                                                                             |
| `securityResources.name`      | string                                                                                                                            | no        | `logging-fluentbit`                                                                        | Set a name of `PodSecurityPolicy`, `SecurityContextConstraints` objects                                                                                             |
| `podMonitor.scrapeTimeout`    | string                                                                                                                            | no        | `30s`                                                                                      | Set metrics scrape interval                                                                                                                                         |
| `podMonitor.scrapeInterval`   | string                                                                                                                            | no        | `10s`                                                                                      | Set metrics scrape timeout                                                                                                                                          |
| `tls`                         | [loggingservice/v11.FluentBitTLS](#fluentbit-tls)                                                                                 | no        | `{}`                                                                                       | Configuration TLS for FluentBit Graylog Output                                                                                                                      |
| `annotations`                 | map                                                                                                                               | no        | `{}`                                                                                       | Allows to specify list of additional annotations                                                                                                                    |
| `labels`                      | map                                                                                                                               | no        | `{}`                                                                                       | Allows to specify list of additional labels                                                                                                                         |
| `priorityClassName`           | string                                                                                                                            | no        | `-`                                                                                        | Pod priority. Priority indicates the importance of a Pod relative to other Pods and prevents them from evicting.                                                    |
| `excludePath`                 | string                                                                                                                            | no        | `-`                                                                                        | Set one or multiple shell patterns separated by commas to exclude files matching certain criteria, e.g: *.gz,*.zip.                                                 |
| `output.loki.enabled` | boolean | Flag for enabling Loki output | no | false |
| `output.loki.host` | string | Loki host | no | `-` |
| `output.loki.tenant` | string | Loki tenant id | no | `-` |
| `output.loki.auth.token.name` | string | Authentication for Loki with token. Name of the secret where token is stored | no | `-` |
| `output.loki.auth.token.key` | string | Authentication for Loki with token. Name of key in the secret where token is stored | no | `-` |
| `output.loki.auth.user.name` | string | Basic authentication credentials for Loki. Name of the secret where username is stored | no | `-` |
| `output.loki.auth.user.key` | string | Basic authentication credentials for Loki. Name of key in the secret where username is stored | no | `-` |
| `output.loki.auth.password.name` | string | Basic authentication credentials for Loki. Name of the secret where password is stored | no | `-` |
| `output.loki.auth.password.key` | string | Basic authentication credentials for Loki. Name of key in the secret where password is stored | no | `-` |
| `output.loki.staticLabels` | string | Static labels that added as stream labels | no | `job=fluentbit` |
| `output.loki.labelsMapping` | string | Labels mappings that defines how to extract labels from each log record. Value should contain a JSON object | no | See example below |
| `output.loki.extraParams` | string | Additional configuration parameters for Loki output. See docs: [https://docs.fluentbit.io/manual/pipeline/outputs/loki#configuration-parameters](https://docs.fluentbit.io/manual/pipeline/outputs/loki#configuration-parameters) | no | See example below |
| `output.loki.tls.enabled` | boolean | Flag to enable TLS connection for Loki output | no | `false` |
| `output.loki.tls.ca.secretName` | string | Name of Secret with Loki CA certificate | no | `-` |
| `output.loki.tls.ca.secretKey` | string | Key (filename) in the Secret with Loki CA certificate | no | `-` |
| `output.loki.tls.cert.secretName` | string | Name of Secret with Loki certificate | no | `-` |
| `output.loki.tls.cert.secretKey` | string | Key (filename) in the Secret with Loki certificate | no | `-` |
| `output.loki.tls.key.secretName` | string | Name of Secret with key | no | `-` |
| `output.loki.tls.key.secretKey` | string | Key (filename) in the Secret with key | no | `-` |
| `output.loki.tls.verify` | boolean | Force certificate validation | no | `true` |
| `output.loki.tls.keyPasswd` | boolean | Optional password for private key file | no | `-` |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
fluentbit:
  install: true
  dockerImage: fluent/fluent-bit:3.1.7

  graylogOutput: true
  graylogHost: graylog.logging.svc
  graylogPort: 12201
  graylogProtocol: tcp

  extraFields:
    foo_key: foo_value
    bar_key: bar_value
  systemLogging: true
  systemLogType: varlogmessages
  systemAuditLogging: true
  kubeAuditLogging: true
  kubeApiserverAuditLogging: true
  containerLogging: true

  customInputConf: |-
    [INPUT]
      Name   random
  customFilterConf: |-
    [FILTER]
      Name record_modifier
      Match *
      Record testField fluent-bit
  customOutputConf: |-
    [OUTPUT]
      Name null
      Match fluent.*
  customLuaScriptConf:
    "script1.lua": |-
      function()
        ...
      end
    "script2.lua": |-
      function()
        ...
      end

  multilineFirstLineRegexp: "/^(\\[\\d{4}\\-\\d{2}\\-\\d{2}|\\{\\\"|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/"
  multilineOtherLinesRegexp: "/^(?!\\[\\d{4}\\-\\d{2}\\-\\d{2}|\\{\\\"|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/"
  billCycleConf: true

  securityContextPrivileged: false
  nodeSelectorKey: kubernetes.io/os
  nodeSelectorValue: linux
  tolerations:
  - key: node-role.kubernetes.io/master
    operator: Exists
  - operator: Exists
    effect: NoExecute
  - operator: Exists
    effect: NoSchedule

  totalLimitSize: 1024M
  memBufLimit: 1024M

  # FluentBit additional volumes
  additionalVolumes:
    - name: dockervolume
      hostPath:
        path: /var/lib/docker
        type: Directory
  additionalVolumeMounts:
    - name: dockervolume
      mountPath: /var/log/docker

  excludePath:
    /var/log/pods/mongo_cnfrs2*/cnfrs2/*.log,
    /var/log/pods/mongo_cnfrs0*/cnfrs0/*.log,
    /var/log/pods/mongo_cnfrs1*/cnfrs1/*.log
```

Example of FluentBit configuration with Loki output enabled:

```yaml
fluentbit:
  install: true
  dockerImage: fluent/fluent-bit:3.1.7

  graylogOutput: false

  output:
    loki:
      enabled: true
      host: loki-write.loki.svc
      tenant: dev-cloud-1
      auth:
        token:
          name: loki-secret
          key: token
        user:
          name: loki-secret
          key: user
        password:
          name: loki-secret
          key: password
      staticLabels: job=fluentbit
      labelsMapping: |-
        {
            "container": "container",
            "pod": "pod",
            "namespace": "namespace",
            "stream": "stream",
            "level": "level",
            "hostname": "hostname",
            "nodename": "nodename",
            "request_id": "request_id",
            "tenant_id": "tenant_id",
            "addressTo": "addressTo",
            "originating_bi_id": "originating_bi_id",
            "spanId": "spanId"
        }
      tls:
        enabled: true
        ca:
          secretName: secret-ca
          secretKey: ca.crt
        cert:
          secretName: secret-cert
          secretKey: certificate.crt
        key:
          secretName: secret-key
          secretKey: privateKey.key
        verify: true
        keyPasswd: secretKeyPassword
      # See docs: https://docs.fluentbit.io/manual/pipeline/outputs/loki#configuration-parameters
      extraParams: |
          workers                2
          Retry_Limit            32
          storage.total_limit_size  5000M
          net.connect_timeout 20
```

[Back to TOC](#table-of-content)

### FluentBit Aggregator

The `fluentbit.aggregator` section contains parameters to enable and configure the FluentBit aggregator.
It can be used to balance the load from FluentBit to Graylog and provide an ability to store logs in case of Graylog
unavailability.

All parameters described below should be specified under a section `fluentbit.aggregator` as the following:

```yaml
fluentbit:
  #...
  aggregator:
    install: true
    #...
```

<!-- markdownlint-disable line-length -->
| Parameter                     | Type                                                                                                                   | Mandatory | Default value                                                                              | Description                                                                                                      |
|-------------------------------| ---------------------------------------------------------------------------------------------------------------------- | --------- |--------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------|
| `install`                     | boolean                                                                                                                | no        | `false`                                                                                    | Allows to install FluentBit aggregator                                                                           |
| `dockerImage`                 | string                                                                                                                 | no        | `-`                                                                                        | Docker image of FluentBit aggregator                                                                             |
| `configmapReload.dockerImage` | string                                                                                                                 | no        | `-`                                                                                        | Docker image of configmap_reload for FluentBit aggregator                                                        |
| `configmapReload.resources`   | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{requests: {cpu: 10m, memory: 10Mi}, limits {cpu: 50m, memory: 50Mi}}`                    | The resources describe to compute resource requests and limits for single Pods                                   |
| `replicas`                    | integer                                                                                                                | no        | `2`                                                                                        | Number of FluentBit aggregator pods                                                                              |
| `graylogOutput`               | boolean                                                                                                                | no        | `true`                                                                                     | Flag for using Graylog output                                                                                    |
| `graylogHost`                 | string                                                                                                                 | no        | `-`                                                                                        | Points to Graylog host. The parameter is used if aggregator is enabled                                           |
| `graylogPort`                 | integer                                                                                                                | no        | `12201`                                                                                    | Graylog port. The parameter is used if aggregator is enabled                                                     |
| `graylogProtocol`             | string                                                                                                                 | no        | `tcp`                                                                                      | The Graylog protocol. Possible values: tcp/udp. The parameter is used if aggregator is enabled                   |
| `extraFields`                 | map[string]string                                                                                                      | no        | `-`                                                                                        | Adds additional custom fields/labels to every log message by using filter based on record_modifier plugin.       |
| `customFilterConf`            | string                                                                                                                 | no        | `-`                                                                                        | Custom filter configuration. The parameter is used if aggregator is enabled                                      |
| `customOutputConf`            | string                                                                                                                 | no        | `-`                                                                                        | Custom output configuration. The parameter is used if aggregator is enabled                                      |
| `customLuaScriptConf`         | map[string]string                                                                                                      | no        | `-`                                                                                        | Set of custom Lua scripts                                                                                        |
| `multilineFirstLineRegexp`    | string                                                                                                                 | no        | `/^(\\[\\d{4}\\-\\d{2}\\-\\d{2}\|\\{\\\"\|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/`   | Custom regexp for the first line of multiline filter                                                             |
| `multilineOtherLinesRegexp`   | string                                                                                                                 | no        | `/^(?!\\[\\d{4}\\-\\d{2}\\-\\d{2}\|\\{\\\"\|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/` | Custom regexp for the other lines of multiline filter                                                            |
| `totalLimitSize`              | string                                                                                                                 | no        | `1024M`                                                                                    | The size limitation of output buffer                                                                             |
| `memBufLimit`                 | string                                                                                                                 | no        | `5M`                                                                                       | Limit of allowed storage for chucks of logs before sending.                                                      |
| `startupTimeout`              | integer                                                                                                                | no        | `8`                                                                                        | Time the operator waits for Aggregator pod(s) to start, in minutes                                               |
| `tolerations`                 | [core/v1.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#toleration-v1-core)          | no        | `[]`                                                                                       | List of tolerations applied to FluentBit Pods                                                                    |
| `affinity`                           | [core/v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#podaffinityterm-v1-core)       | no        | `-`                                                                          | It specifies the pod\'s scheduling constraints                                                                                                                                                                                                                                                                          |
| `resources`                   | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{requests: {cpu: 500m, memory: 512Mi}, limits {cpu: 2000m, memory: 2048Mi}}`              | The resources describe to compute resource requests and limits for single Pods                                   |
| `volume.bind`                 | boolean                                                                                                                | no        | `false`                                                                                    | Allows installing PVCs for Aggregator pods                                                                       |
| `volume.storageClassName`     | string                                                                                                                 | no        | `""`                                                                                       | Aggregator PVC storage class name                                                                                |
| `volume.storageSize`          | string                                                                                                                 | no        | `2Gi`                                                                                      | Storage size limit of PVC                                                                                        |
| `securityResources.install`   | boolean                                                                                                                | no        | `false`                                                                                    | Enable creating security resources as `PodSecurityPolicy`, `SecurityContextConstraints`                          |
| `securityResources.name`      | string                                                                                                                 | no        | `logging-fluentbit-aggregator`                                                             | Set a name of `PodSecurityPolicy`, `SecurityContextConstraints` objects                                          |
| `podMonitor.scrapeInterval`   | string                                                                                                                 | no        | `30s`                                                                                      | Set metrics scrape interval                                                                                      |
| `podMonitor.scrapeTimeout`    | string                                                                                                                 | no        | `10s`                                                                                      | Set metrics scrape timeout                                                                                       |
| `tls`                         | [loggingservice/v11.FluentBitTLS](#fluentbit-tls)                                                                      | no        | `{}`                                                                                       | Configuration TLS for FluentBit Graylog Output                                                                   |
| `priorityClassName`           | string                                                                                                                 | no        | `-`                                                                                        | Pod priority. Priority indicates the importance of a Pod relative to other Pods and prevents them from evicting. |
| `output.loki.enabled` | boolean | Flag for enabling Loki output | no | false |
| `output.loki.host` | string | Loki host | no | `-` |
| `output.loki.tenant` | string | Loki tenant id | no | `-` |
| `output.loki.auth.token.name` | string | Authentication for Loki with token. Name of the secret where token is stored | no | `-` |
| `output.loki.auth.token.key` | string | Authentication for Loki with token. Name of key in the secret where token is stored | no | `-` |
| `output.loki.auth.user.name` | string | Basic authentication credentials for Loki. Name of the secret where username is stored | no | `-` |
| `output.loki.auth.user.key` | string | Basic authentication credentials for Loki. Name of key in the secret where username is stored | no | `-` |
| `output.loki.auth.password.name` | string | Basic authentication credentials for Loki. Name of the secret where password is stored | no | `-` |
| `output.loki.auth.password.key` | string | Basic authentication credentials for Loki. Name of key in the secret where password is stored | no | `-` |
| `output.loki.staticLabels` | string | Static labels that added as stream labels | no | `job=fluentbit` |
| `output.loki.labelsMapping` | string | Labels mappings that defines how to extract labels from each log record. Value should contain a JSON object | no | See example below |
| `output.loki.extraParams` | string | Additional configuration parameters for Loki output. See docs: [https://docs.fluentbit.io/manual/pipeline/outputs/loki#configuration-parameters](https://docs.fluentbit.io/manual/pipeline/outputs/loki#configuration-parameters) | no | See example below |
| `output.loki.tls.enabled` | boolean | Flag to enable TLS connection for Loki output | no | `false` |
| `output.loki.tls.ca.secretName` | string | Name of Secret with Loki CA certificate | no | `-` |
| `output.loki.tls.ca.secretKey` | string | Key (filename) in the Secret with Loki CA certificate | no | `-` |
| `output.loki.tls.cert.secretName` | string | Name of Secret with Loki certificate | no | `-` |
| `output.loki.tls.cert.secretKey` | string | Key (filename) in the Secret with Loki certificate | no | `-` |
| `output.loki.tls.key.secretName` | string | Name of Secret with key | no | `-` |
| `output.loki.tls.key.secretKey` | string | Key (filename) in the Secret with key | no | `-` |
| `output.loki.tls.verify` | boolean | Force certificate validation | no | `true` |
| `output.loki.tls.keyPasswd` | boolean | Optional password for private key file | no | `-` |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
fluentbit:
  aggregator:
    install: true
    dockerImage: fluent/fluent-bit:3.1.7
    replicas: 2

    tolerations:
    - key: node-role.kubernetes.io/master
      operator: Exists
    - operator: Exists
      effect: NoExecute
    - operator: Exists
      effect: NoSchedule

    graylogOutput: true
    graylogHost: graylog.logging.svc
    graylogPort: 12201
    graylogProtocol: tcp

    extraFields:
      foo_key: foo_value
      bar_key: bar_value
    customFilterConf: |-
      [FILTER]
        Name record_modifier
        Match *
        Record testField fluent-bit
    customOutputConf: |-
      [OUTPUT]
        Name null
        Match fluent.*
    customLuaScriptConf:
      "script1.lua": |-
        function()
          ...
        end
      "script2.lua": |-
        function()
          ...
        end

    multilineFirstLineRegexp: "/^(\\[\\d{4}\\-\\d{2}\\-\\d{2}|\\{\\\"|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/"
    multilineOtherLinesRegexp: "/^(?!\\[\\d{4}\\-\\d{2}\\-\\d{2}|\\{\\\"|\\u001b\\[.{1,5}m\\d{2}\\:\\d{2}\\:\\d{2}).*/"
    totalLimitSize: 1024M
    memBufLimit: 5M

    volume:
      bind: true
      storageClassName: cinder
      storageSize: 200Gi
```

Example of FluentBit HA configuration with Loki output enabled:

```yaml
fluentbit:
  install: true
  dockerImage: fluent/fluent-bit:3.1.7

  aggregator:
    install: true
    dockerImage: fluent/fluent-bit:3.1.7
    replicas: 2
    graylogOutput: false
    output:
      loki:
        enabled: true
        host: loki-write.loki.svc
        tenant: dev-cloud-1
        auth:
          token:
            name: loki-secret
            key: token
          user:
            name: loki-secret
            key: user
          password:
            name: loki-secret
            key: password
        staticLabels: job=fluentbit
        labelsMapping: |-
          {
              "container": "container",
              "pod": "pod",
              "namespace": "namespace",
              "stream": "stream",
              "level": "level",
              "hostname": "hostname",
              "nodename": "nodename",
              "request_id": "request_id",
              "tenant_id": "tenant_id",
              "addressTo": "addressTo",
              "originating_bi_id": "originating_bi_id",
              "spanId": "spanId"
          }
        tls:
          enabled: true
          ca:
            secretName: secret-ca
            secretKey: ca.crt
          cert:
            secretName: secret-cert
            secretKey: certificate.crt
          key:
            secretName: secret-key
            secretKey: privateKey.key
          verify: true
          keyPasswd: secretKeyPassword
        # See docs: https://docs.fluentbit.io/manual/pipeline/outputs/loki#configuration-parameters
        extraParams: |
            workers                2
            Retry_Limit            32
            storage.total_limit_size  5000M
            net.connect_timeout 20
```

[Back to TOC](#table-of-content)

### FluentBit TLS

The `fluentbit.tls` or `fluentbit.aggregator.tls` section contains parameters to configure TLS for
FluentBit Graylog Output.

All parameters described below should be specified under a section `fluentbit.tls` or `fluentbit.aggregator.tls`
as the following:

```yaml
fluentbit:
  tls:
    enable: true
    #...
```

or

```yaml
fluentbit:
  aggregator:
    tls:
      enable: true
      #...
```

<!-- markdownlint-disable line-length -->
| Parameter                         | Type    | Mandatory | Default value | Description                                                                                                                  |
| --------------------------------- | ------- | --------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `enabled`                         | boolean | no        | `false`       | Enable TLS for FluentBit                                                                                                     |
| `verify`                          | boolean | no        | `true`        | Enable certificate validation                                                                                                |
| `keyPasswd`                       | string  | no        | `-`           | Password for private key file                                                                                                |
| `ca.secretName`                   | string  | no        | `-`           | Name of Kubernetes Secret with CA certificate. Mutually exclusive with `generateCerts` section                               |
| `ca.secretKey`                    | string  | no        | `-`           | Key (filename) in the Secret with CA certificate                                                                             |
| `cert.secretName`                 | string  | no        | `-`           | Name of Kubernetes Secret with client certificate. Mutually exclusive with `generateCerts` section                           |
| `cert.secretKey`                  | string  | no        | `-`           | Key (filename) in the Secret with client certificate                                                                         |
| `key.secretName`                  | string  | no        | `-`           | Name of Kubernetes Secret with key for the client certificate. Mutually exclusive with `generateCerts` section               |
| `key.secretKey`                   | string  | no        | `-`           | Key (filename) in the Secret with key for the client certificate                                                             |
| `generateCerts.enabled`           | boolean | no        | `-`           | Enabling integration with `cert-manager` to generate certificates. Mutually exclusive with `ca`, `cert` and `key` parameters |
| `generateCerts.secretName`        | string  | no        | `-`           | Secret name with certificates that will generate by `cert-manager`                                                           |
| `generateCerts.clusterIssuerName` | string  | no        | `-`           | Issuer that will use to generate certificates                                                                                |
| `generateCerts.duration`          | string  | no        | `-`           | Set certificates validity period                                                                                             |
| `generateCerts.renewBefore`       | string  | no        | `-`           | Sets the number of days before the certificates expiration day for which they will be reissued                               |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
fluentbit:
  tls:
    enabled: true

    verify: true
    keyPasswd: secret

    # Certificates from Kubernetes Secrets
    ca:
      secretName: fluentbit-tls-assets-0
      secretKey: ca.crt
    cert:
      secretName: fluentbit-tls-assets-0
      secretKey: cert.crt
    key:
      secretName: fluentbit-tls-assets-0
      secretKey: key.crt

    # Integration with cert-manager
    generateCerts:
      enabled: true
      secretName: fluentbit-cert-manager-tls-assets-0
      clusterIssuerName: ""
      duration: 365
      renewBefore: 15
```

[Back to TOC](#table-of-content)

## FluentD

The `fluentd` section contains parameters to configure FluentD logging agent.

All parameters described below should be specified under a section `fluentd` as the following:

```yaml
fluentd:
  install: true
  #...
```

<!-- markdownlint-disable line-length -->
| Parameter                      | Type                                                                                                                               | Mandatory | Default value                                                                | Description                                                                                                                                                                                                      |
|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------| --------- |------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `install`                      | boolean                                                                                                                            | no        | `true`                                                                       | Flag for installation `logging-fluentd`                                                                                                                                                                          |
| `dockerImage`                  | string                                                                                                                             | no        | `-`                                                                          | Docker image of FluentD                                                                                                                                                                                          |
| `configmapReload.dockerImage`  | string                                                                                                                             | no        | `-`                                                                          | Docker image of configmap_reload for FluentD                                                                                                                                                          |
| `configmapReload.resources`    | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)             | no        | `{requests: {cpu: 10m, memory: 10Mi}, limits {cpu: 50m, memory: 50Mi}}`      | The resources describe to compute resource requests and limits for single Pods                                                                                                                                   |
| `ip_v6`                        | boolean                                                                                                                            | no        | `false`                                                                      | Flag for using IPv6 environment                                                                                                                                                                                  |
| `nodeSelectorKey`              | string                                                                                                                             | no        | `-`                                                                          | NodeSelector key, can be multiple by OR condition, separated by comma, usually `role`                                                                                                                            |
| `nodeSelectorValue`            | string                                                                                                                             | no        | `-`                                                                          | NodeSelector value, can be multiple by OR condition, separated by comma, usually `compute`                                                                                                                       |
| `tolerations`                  | [core/v1.Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#toleration-v1-core)                      | no        | `[]`                                                                         | List of tolerations applied to FluentD Pods                                                                                                                                                                      |
| `affinity`                           | [core/v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#podaffinityterm-v1-core)       | no        | `-`                                                                          | It specifies the pod\'s scheduling constraints                                                                                                                                                                                                                                                                          |
| `resources`                    | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core)             | no        | `{requests: {cpu: 100m, memory: 128Mi}, limits: {cpu: 500m, memory: 512Mi}}` | The resources describe to compute resource requests and limits for single Pods                                                                                                                                   |
| `graylogOutput`                | boolean                                                                                                                            | no        | `true`                                                                       | Flag for using Graylog output                                                                                                                                                                                    |
| `graylogHost`                  | string                                                                                                                             | no        | `-`                                                                          | Points to Graylog host                                                                                                                                                                                           |
| `graylogPort`                  | integer                                                                                                                            | no        | `12201`                                                                      | Graylog port                                                                                                                                                                                                     |
| `graylogProtocol`              | string                                                                                                                             | no        | `tcp`                                                                        | The Graylog protocol. Available values: `tcp`, `udp`. **Note** that the liveness probe for FluentD pods always returns success if the udp protocol is selected                                                   |
| `graylogBufferFlushInterval`   | string                                                                                                                             | no        | `5s`                                                                         | Interval of buffer flush                                                                                                                                                                                         |
| `esHost`                       | string                                                                                                                             | no        | `-`                                                                          | **Deprecated** Points to ElasticSearch host                                                                                                                                                                      |
| `esPort`                       | integer                                                                                                                            | no        | `-`                                                                          | **Deprecated** ElasticSearch port                                                                                                                                                                                |
| `esUsername`                   | string                                                                                                                             | no        | `-`                                                                          | **Deprecated** Username for ElasticSearch authentication                                                                                                                                                         |
| `esPassword`                   | string                                                                                                                             | no        | `-`                                                                          | **Deprecated** Password for ElasticSearch authentication                                                                                                                                                         |
| `extraFields`                  | map[string]string                                                                                                                  | no        | `-`                                                                          | Adds additional custom fields/labels to every log message by using filter based on record_transformer plugin. This parameter will override existing fields if their keys match those specified in `extraFields`. |
| `customInputConf`              | string                                                                                                                             | no        | `-`                                                                          | FluentD custom input configuration                                                                                                                                                                               |
| `customFilterConf`             | string                                                                                                                             | no        | `-`                                                                          | FluentD custom filter configuration                                                                                                                                                                              |
| `customOutputConf`             | string                                                                                                                             | no        | `-`                                                                          | FluentD custom output configuration                                                                                                                                                                              |
| `multilineFirstLineRegexp`     | string                                                                                                                             | no        | `/(^\\[\\d{4}\\-\\d{2}\\-\\d{2})\|(^\\{\")\|(^*0m\\d{2}\\:\\d{2}\\:\\d{2})/` | FluentD custom regexp for multiline filter                                                                                                                                                                       |
| `billCycleConf`                | boolean                                                                                                                            | no        | `false`                                                                      | FluentD filter for bil-cycle-logs stream                                                                                                                                                                         |
| `watchKubernetesMetadata`      | boolean                                                                                                                            | no        | `true`                                                                       | Set up a watch on pods on the API server for updates to metadata                                                                                                                                                 |
| `securityContextPrivileged`    | boolean                                                                                                                            | no        | `false`                                                                      | Allows specifying securityContext.privileged for FluentD container                                                                                                                                               |
| `additionalVolumes`            | [core/v1.PersistentVolumeSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#persistentvolumespec-v1-core)  | no        | `false`                                                                      | Additional volumes for FluentD                                                                                                                                                                                   |
| `additionalVolumeMounts`       | [core/v1.VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#volumemount-v1-core)                    | no        | `false`                                                                      | Allows volume-mounts for FluentD                                                                                                                                                                                 |
| `totalLimitSize`               | string                                                                                                                             | no        | `512MB`                                                                      | The size limitation of output buffer                                                                                                                                                                             |
| `fileStorage`                  | boolean                                                                                                                            | no        | `false`                                                                      | Flag for using file storage instead of memory                                                                                                                                                                    |
| `compress`                     | string                                                                                                                             | no        | `text`                                                                       | If `gzip` is set, Fluentd compresses data records before writing to buffer chunks.                                                                                                                               |
| `queueLimitLength`             | integer                                                                                                                            | no        | `-`                                                                          | **Deprecated** Max length of buffers queue                                                                                                                                                                       |
| `systemLogging`                | boolean                                                                                                                            | no        | `false`                                                                      | Enable system log                                                                                                                                                                                                |
| `systemLogType`                | string                                                                                                                             | no        | `varlogmessages`                                                             | Set type of system log. Available values `varlogmessages`, `varlogsyslog` and `systemd`                                                                                                                          |
| `systemAuditLogging`           | boolean                                                                                                                            | no        | `true`                                                                       | Enable input for system audit logs from `/var/log/audit/audit.log`.                                                                                                                                              |
| `kubeAuditLogging`             | boolean                                                                                                                            | no        | `true`                                                                       | Enable input for Kubernetes audit logs from `/var/log/kubernetes/kube-apiserver-audit.log` and `/var/log/kubernetes/audit.log`.                                                                                  |
| `kubeApiserverAuditLogging`    | boolean                                                                                                                            | no        | `true`                                                                       | Enable input for Kubernetes APIServer audit logs from `/var/log/kube-apiserver/audit.log` for Kubernetes and `/var/log/openshift-apiserver/audit.log` for OpenShift.                                             |
| `containerLogging`             | boolean                                                                                                                            | no        | `true`                                                                       | Enable input for container logs from `/var/logs/containers` for Docker or `/var/log/pods` for other engines.                                                                                                     |
| `securityResources.install`    | boolean                                                                                                                            | no        | `false`                                                                      | Enable creating security resources as `PodSecurityPolicy`, `SecurityContextConstraints`                                                                                                                          |
| `securityResources.name`       | string                                                                                                                             | no        | `logging-fluentd`                                                            | Set a name of `PodSecurityPolicy`, `SecurityContextConstraints` objects                                                                                                                                          |
| `podMonitor.scrapeInterval`    | string                                                                                                                             | no        | `30s`                                                                        | Set metrics scrape interval                                                                                                                                                                                      |
| `podMonitor.scrapeTimeout`     | string                                                                                                                             | no        | `10s`                                                                        | Set metrics scrape timeout                                                                                                                                                                                       |
| `tls`                          | [loggingservice/v11.FluentDTLS](#fluentd-tls)                                                                                      | no        | `{}`                                                                         | Configuration TLS for FluentD Graylog Output                                                                                                                                                                     |
| `excludePath`                  | []string                                                                                                                           | no        | `[]`                                                                         | Path to exclude logs. It can contain multiple values.                                                                                                                                                            |
| `annotations`                  | map                                                                                                                                | no        | `{}`                                                                         | Allows to specify list of additional annotations                                                                                                                                                                 |
| `labels`                       | map                                                                                                                                | no        | `{}`                                                                         | Allows to specify list of additional labels                                                                                                                                                                      |
| `priorityClassName`            | string                                                                                                                             | no        | `-`                                                                          | Pod priority. Priority indicates the importance of a Pod relative to other Pods and prevents them from evicting                                                                                                  |
| `cloudEventsReaderFormat`      | string                                                                                                                             | no        | `json`                                                                       | Allow to add filter to parse json of Kubernetes events logs from Cloud Events Reader. Possible value is `json`. If other value is set, no additional parsing configuration will be added in FluentBit/Fluentd    |
| `output.loki.enabled` | boolean | Flag for enabling Loki output | no | false |
| `output.loki.host` | string | Loki host | no | `-` |
| `output.loki.tenant` | string | Loki tenant id | no | `-` |
| `output.loki.auth.token.name` | string | Authentication for Loki with token. Name of the secret where token is stored | no | `-` |
| `output.loki.auth.token.key` | string | Authentication for Loki with token. Name of key in the secret where token is stored | no | `-` |
| `output.loki.auth.user.name` | string | Basic authentication credentials for Loki. Name of the secret where username is stored | no | `-` |
| `output.loki.auth.user.key` | string | Basic authentication credentials for Loki. Name of key in the secret where username is stored | no | `-` |
| `output.loki.auth.password.name` | string | Basic authentication credentials for Loki. Name of the secret where password is stored | no | `-` |
| `output.loki.auth.password.key` | string | Basic authentication credentials for Loki. Name of key in the secret where password is stored | no | `-` |
| `output.loki.staticLabels` | string | Static labels that added as stream labels | no | `{"job":"fluentd"}` |
| `output.loki.labelsMapping` | string | Labels mappings that defines how to extract labels from each log record | no | See example below |
| `output.loki.extraParams` | string | Additional configuration parameters for Loki output. Buffer can be configured here and other available parameters of Fluentd Loki output plugin. See all the parameters here: [https://github.com/grafana/loki/blob/main/clients/cmd/fluentd/lib/fluent/plugin/out_loki.rb](https://github.com/grafana/loki/blob/main/clients/cmd/fluentd/lib/fluent/plugin/out_loki.rb) | no | `-` |
| `output.loki.tls.enabled` | boolean | Flag to enable TLS connection for Loki output | no | `false` |
| `output.loki.tls.ca.secretName` | string | Name of Secret with Loki CA certificate | no | `-` |
| `output.loki.tls.ca.secretKey` | string | Key (filename) in the Secret with Loki CA certificate | no | `-` |
| `output.loki.tls.cert.secretName` | string | Name of Secret with Loki certificate | no | `-` |
| `output.loki.tls.cert.secretKey` | string | Key (filename) in the Secret with Loki certificate | no | `-` |
| `output.loki.tls.key.secretName` | string | Name of Secret with key | no | `-` |
| `output.loki.tls.key.secretKey` | string | Key (filename) in the Secret with key | no | `-` |
| `output.loki.tls.allCiphers` | boolean | Allows any ciphers to be used, may be insecure | no | `true` |
| `output.loki.tls.version` | boolean | Any of :TLSv1, :TLSv1_1, :TLSv1_2 | no | `-` |
| `output.loki.tls.noVerify` | boolean | Force certificate validation | no | `false` |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
fluentd:
  install: true
  dockerImage: qubership-fluentd:main
  ip_v6: false

  nodeSelectorKey: kubernetes.io/os
  nodeSelectorValue: linux
  excludePath:
    - "/var/log/pods/openshift-dns_dns-default*/dns/*.log"
  tolerations:
    - key: node-role.kubernetes.io/master
      operator: Exists
    - operator: Exists
      effect: NoExecute
    - operator: Exists
      effect: NoSchedule

  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  
  # Graylog input settings
  systemLogging: true
  systemLogType: varlogmessages
  systemAuditLogging: true
  kubeAuditLogging: true
  kubeApiserverAuditLogging: true
  containerLogging: true

  # Graylog output settings
  graylogOutput: true
  graylogHost: graylog.logging.svc
  graylogPort: 12201
  graylogProtocol: tcp

  # Custom FluentD configurations
  extraFields:
    foo_key: foo_value
    bar_key: bar_value
  customInputConf: |-
    <source>
      custom_input_configuration
    </source>
  customFilterConf: |-
    <filter raw.kubernetes.var.log.**>
      custom_filter_configuration
    </filter>
  customOutputConf: |-
    <store ignore_error>
      custom_output_configuration
    </store>

  multilineFirstLineRegexp: /(^\\[\\d{4}\\-\\d{2}\\-\\d{2})\|(^\\{\")\|(^*0m\\d{2}\\:\\d{2}\\:\\d{2})/
  billCycleConf: true

  watchKubernetesMetadata: true
  securityContextPrivileged: true
  totalLimitSize: 1GB

  # FluentD additional volumes
  additionalVolumes:
    - name: dockervolume
      hostPath:
        path: /var/lib/docker
        type: Directory
  additionalVolumeMounts:
    - name: dockervolume
      mountPath: /var/log/docker

  # FluentD TLS Graylog Output settings
  tls:
    ...
```

Example of FluentBit HA configuration with Loki output enabled:

```yaml
fluentd:
  install: true
  dockerImage: qubership-fluentd:main
  output:
    loki:
      enabled: true
      host: https://loki-write.loki.svc:3100
      tenant: dev-env-tenant-1
      auth:
        token:
          name: loki-secret
          key: token
        user:
          name: loki-secret
          key: user
        password:
          name: loki-secret
          key: password
      staticLabels: {"job":"fluentd"}
      labelsMapping: |-
        stream $.stream
        container $.container
        pod $.pod
        namespace $.namespace
        level $.level
        hostname $.hostname
        nodename $.kubernetes_host
        request_id $.request_id
        tenant_id $.tenant_id
        addressTo $.addressTo
        originating_bi_id $.originating_bi_id
        spanId $.spanId
      tls:
        enabled: true
        ca:
          secretName: secret-ca
          secretKey: ca.crt
        cert:
          secretName: secret-cert
          secretKey: certificate.crt
        key:
          secretName: secret-key
          secretKey: privateKey.key
        allCiphers: true
        version: ":TLSv1_2"
        noVerify: false
      # Buffer can be configured here and other available parameters of Fluentd Loki output plugin.
      # See all the parameters here: https://github.com/grafana/loki/blob/main/clients/cmd/fluentd/lib/fluent/plugin/out_loki.rb
      extraParams: |
        extract_kubernetes_labels false
        remove_keys []
        custom_headers header:value
```

[Back to TOC](#table-of-content)

### FluentD TLS

The `fluentd.tls` section contains parameters to configure TLS for FluentD Graylog Output.

All parameters described below should be specified under a section `fluentd.tls` as the following:

```yaml
fluentd:
  tls:
    enabled: true
    #...
```

<!-- markdownlint-disable line-length -->
| Parameter                         | Type    | Mandatory | Default value | Description                                                                                                                  |
| --------------------------------- | ------- | --------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `enabled`                         | boolean | no        | `false`       | EnablE TLS for FluentD output in Graylog                                                                                     |
| `noDefaultCA`                     | boolean | no        | `false`       | Prevents OpenSSL from using the systems CA store                                                                             |
| `version`                         | string  | no        | `:TLSv1_2`    | TLS version. Available values `:TLSv1`, `:TLSv1_1`, `:TLSv1_2`                                                               |
| `allCiphers`                      | boolean | no        | `true`        | Allows any ciphers to be used, may be insecure                                                                               |
| `rescueSslErrors`                 | boolean | no        | `false`       | Similar to rescue_network_errors in notifier.rb, allows SSL exceptions to be raised                                          |
| `noVerify`                        | boolean | no        | `false`       | Disable peer verification                                                                                                    |
| `ca.secretName`                   | string  | no        | `-`           | Name of Kubernetes Secret with CA certificate. Mutually exclusive with `generateCerts` section                               |
| `ca.secretKey`                    | string  | no        | `-`           | Key (filename) in the Secret with CA certificate                                                                             |
| `cert.secretName`                 | string  | no        | `-`           | Name of Kubernetes Secret with client certificate. Mutually exclusive with `generateCerts` section                           |
| `cert.secretKey`                  | string  | no        | `-`           | Key (filename) in the Secret with client certificate                                                                         |
| `key.secretName`                  | string  | no        | `-`           | Name of Kubernetes Secret with key for the client certificate. Mutually exclusive with `generateCerts` section               |
| `key.secretKey`                   | string  | no        | `-`           | Key (filename) in the Secret with key for the client certificate                                                             |
| `generateCerts.enabled`           | boolean | no        | `-`           | Enabling integration with `cert-manager` to generate certificates. Mutually exclusive with `ca`, `cert` and `key` parameters |
| `generateCerts.secretName`        | string  | no        | `-`           | Secret name with certificates that will generate by `cert-manager`                                                           |
| `generateCerts.clusterIssuerName` | string  | no        | `-`           | Issuer that will use to generate certificates                                                                                |
| `generateCerts.duration`          | string  | no        | `-`           | Set certificates validity period                                                                                             |
| `generateCerts.renewBefore`       | string  | no        | `-`           | Sets the number of days before the certificates expiration day for which they will be reissued                               |

<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
fluentd:
  tls:
    enabled: true

    noDefaultCA: false
    version: ":TLSv1_2"
    allCiphers: true
    rescueSslErrors: false
    noVerify: false

    # Certificates from Kubernetes Secrets
    ca:
      secretName: fluentd-tls-assets-0
      secretKey: ca.crt
    cert:
      secretName: fluentd-tls-assets-0
      secretKey: cert.crt
    key:
      secretName: fluentd-tls-assets-0
      secretKey: key.crt

    # Integration with cert-manager
    generateCerts:
      enabled: true
      secretName: fluentd-cert-manager-tls
      clusterIssuerName: ""
      duration: 365
      renewBefore: 15
```

[Back to TOC](#table-of-content)

## Cloud Events Reader

The `cloudEventsReader` section contains parameters to configure cloud-events-reader that collects and exposes
Kubernetes/OpenShift events.

All parameters described below should be specified under a section `cloudEventsReader` as the following:

```yaml
cloudEventsReader:
  install: true
  #...
```

<!-- markdownlint-disable line-length -->
| Parameter           | Type                                                                                                                   | Mandatory | Default value                                                                | Description                                                                                                      |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------- | --------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `install`           | boolean                                                                                                                | no        | `true`                                                                       | Flag for installation `cloud-events-reader`                                                                      |
| `dockerImage`       | string                                                                                                                 | no        | `-`                                                                          | Docker image of Cloud Events Reader                                                                              |
| `resources`         | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{requests: {cpu: 100m, memory: 128Mi}, limits: {cpu: 100m, memory: 128Mi}}` | The resources describe to compute resource requests and limits for single Pods                                   |
| `nodeSelectorKey`   | string                                                                                                                 | no        | `-`                                                                          | NodeSelector key, usually `role`                                                                                 |
| `nodeSelectorValue` | string                                                                                                                 | no        | `-`                                                                          | NodeSelector value, usually `compute`                                                                            |
| `affinity`                           | [core/v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#podaffinityterm-v1-core)       | no        | `-`                                                                          | It specifies the pod\'s scheduling constraints                                                                                                                                                                                                                                                                          |
| `annotations`       | map                                                                                                                    | no        | `{}`                                                                         | Allows to specify list of additional annotations                                                                 |
| `labels`            | map                                                                                                                    | no        | `{}`                                                                         | Allows to specify list of additional labels                                                                      |
| `priorityClassName` | string                                                                                                                 | no        | `-`                                                                          | Pod priority. Priority indicates the importance of a Pod relative to other Pods and prevents them from evicting. |
| `args`              | []string                                                                                                               | no        | `-`                                                                          | Command line arguments for Cloud Events Reader                                                                   |
<!-- markdownlint-enable line-length -->

More information about setting `args` described [here](user-guides/cloud-events.md).

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
cloudEventsReader:
  install: true
  dockerImage: k8s-events-reader:main
  resources:
    requests:
      cpu:
      memory:
    limits:
      cpu:
      memory:
  nodeSelectorKey: kubernetes.io/os
  nodeSelectorValue: linux
```

[Back to TOC](#table-of-content)

## Integration tests

The `integrationTests` section contains parameters to enable integration tests that can verify deployment of
Graylog, FluentBit or FluentD.

All parameters described below should be specified under a section `integrationTests` as the following:

```yaml
integrationTests:
  install: true
  #...
```

<!-- markdownlint-disable line-length -->
| Parameter                            | Type                                                                                                                   | Mandatory | Default value                                                                | Description                                                                                                                                                                                                                                                                                                             |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- | --------- | ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `install`                            | boolean                                                                                                                | no        | `false`                                                                      | Flag for run `integration-tests`                                                                                                                                                                                                                                                                                        |
| `image`                              | string                                                                                                                 | no        | `-`                                                                          | Docker image of `integration-tests`                                                                                                                                                                                                                                                                                     |
| `service.name`                       | string                                                                                                                 | no        | `logging-integration-tests-runner`                                           | The name of Logging integration tests service                                                                                                                                                                                                                                                                           |
| `tags`                               | string                                                                                                                 | no        | `smoke`                                                                      | The tags that select test cases to run                                                                                                                                                                                                                                                                                  |
| `externalGraylogServer`              | string                                                                                                                 | no        | `true`                                                                       | The kind of Graylog for testing                                                                                                                                                                                                                                                                                         |
| `graylogProtocol`                    | string                                                                                                                 | no        | `-`                                                                          | Graylog protocol                                                                                                                                                                                                                                                                                                        |
| `graylogHost`                        | string                                                                                                                 | no        | `-`                                                                          | The host name of Graylog                                                                                                                                                                                                                                                                                                |
| `graylogPort`                        | integer                                                                                                                | no        | `80`                                                                         | The Graylog HTTP port                                                                                                                                                                                                                                                                                                   |
| `affinity`                           | [core/v1.Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#podaffinityterm-v1-core)       | no        | `-`                                                                          | It specifies the pod\'s scheduling constraints                                                                                                                                                                                                                                                                          |
| `resources`                          | [core/v1.Resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | no        | `{requests: {cpu: 100m, memory: 128Mi}, limits: {cpu: 200m, memory: 256Mi}}` | The resources describe to compute resource requests and limits for Integration Tests Pod                                                                                                                                                                                                                                |
| `statusWriting.enabled`              | boolean                                                                                                                | no        | false                                                                        | Enable Store tests status to `LoggingService` custom resource                                                                                                                                                                                                                                                           |
| `statusWriting.isShortStatusMessage` | boolean                                                                                                                | no        | true                                                                         | The size of integration test status message                                                                                                                                                                                                                                                                             |
| `statusWriting.onlyIntegrationTests` | boolean                                                                                                                | no        | true                                                                         | Deploy only integration tests without any component (component was installed before).                                                                                                                                                                                                                                   |
| `statusWriting.customResourcePath`   | string                                                                                                                 | no        | `logging.qubership.org/v11/logging-operator/loggingservices/logging-service`        | Path to Custom Resource that should be used to write status of integration-tests execution. The value is a field from k8s entity selfLink without `apis` prefix and `namespace` part. The path should be composed according to the following template: `<group>/<apiversion>/<namespace>/<plural>/<customResourceName>` |
| `annotations`                        | map                                                                                                                    | no        | `{}`                                                                         | Allows to specify list of additional annotations                                                                                                                                                                                                                                                                        |
| `labels`                             | map                                                                                                                    | no        | `{}`                                                                         | Allows to specify list of additional labels                                                                                                                                                                                                                                                                             |
| `priorityClassName`                  | string                                                                                                                 | no        | `-`                                                                          | Pod priority. Priority indicates the importance of a Pod relative to other Pods and prevents them from evicting.                                                                                                                                                                                                        |
| `vmUser`                             | string                                                                                                                 | no        | `-`                                                                          | User for ssh login to VM.                                                                                                                                                                                                                                                                                               |
| `sshKey`                             | string                                                                                                                 | no        | `-`                                                                          | SSH key for ssh login to VM.                                                                                                                                                                                                                                                                                            |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It\'s just an example of a parameter\'s format, not a recommended parameter.

```yaml
integrationTests:
  install: true
  image: qubership-logging-integration-tests:main

  service:
    name: logging-integration-tests-runner
  tags: smoke
  externalGraylogServer: true
  graylogHost: 1.2.3.4
  graylogPort: 80
  vmUser: ubuntu
  sshKey: |
    -----BEGIN RSA PRIVATE KEY-----
    ................................
    -----END RSA PRIVATE KEY-----

  statusWriting:
    enabled: true
    isShortStatusMessage: false
    onlyIntegrationTests: false
    customResourcePath: logging.qubership.org/v11/logging-operator/loggingservices/logging-service
```

[Back to TOC](#table-of-content)

# Installation

This section describes how to install Logging and its components to the Kubernetes.

## Before you begin

* Make sure that selecting OpenSearch is working and has enough resources to handle a load from Graylog
  (in case of deploying Graylog in the Cloud)

[Back to TOC](#table-of-content)

## On-prem

TBD

## Amazon Web Services (AWS)

TBD

# Post Installation Steps

## Configuring URL whitelist

After successful deploy you can configure URL whitelist.
There are certain components in Graylog which will perform outgoing HTTP requests. Among those, are event notifications
and HTTP-based data adapters.
Allowing Graylog to interact with resources using arbitrary URLs may pose a security risk. HTTP requests are executed
from Graylog servers and might therefore be able to reach more sensitive systems than an external user would have
access to, including AWS EC2 metadata, which can contain keys and other secrets, Elasticsearch and others.
It is therefore advisable to restrict access by explicitly whitelisting URLs which are considered safe. HTTP requests
will be validated against the Whitelist and are prohibited if there is no Whitelist entry matching the URL.

The Whitelist configuration is located at `System/Configurations/URL Whitelist`. The Whitelist is enabled by default.

If the security implications mentioned above are of no concern, the Whitelist can be completely disabled.
When disabled, HTTP requests will not be restricted.

Whitelist entries of type `Exact match` contain a string which will be matched against a URL by direct comparison.
If the URL is equal to this string, it is considered to be whitelisted.

Whitelist entries of type `Regex` contain a regular expression. If a URL matches the regular expression, the URL is
considered to be whitelisted. Graylog uses the Java Pattern class to evaluate regular expressions.

[Back to TOC](#table-of-content)

# Upgrade

# Post Deploy Checks

There are some options to check after deployment that Logging is deployed and working correctly.

So this topic should cover the theme of how to check that Logging is working now.

[Back to TOC](#table-of-content)

## Jobs Post Deploy Check

## Smoke test

```bash
# verify installation
$oc get pods -n logging
NAME                                               READY     STATUS             RESTARTS   AGE
events-reader-64d6698bb8-tfp5l                     1/1       Running            0          1m
logging-fluentd-sh86l                              1/1       Running            0          1m
logging-service-operator-7b586d8767-lpwzl          1/1       Running            0          1m
```

## Integration tests

# Frequently asked questions

# Footnotes
