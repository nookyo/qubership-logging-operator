This document describe how to configure Graylog in different cases.

# Table of Content

* [Table of Content](#table-of-content)
* [Using FluentD/FluentBit with compression](#using-fluentdfluentbit-with-compression)
  * [FluentBit](#fluentbit)
  * [FluentD](#fluentd)
* [Using Graylog](#using-graylog)
  * [Viewing Logs of a Specific Microservice](#viewing-logs-of-a-specific-microservice)
  * [Create Stream](#create-stream)
  * [Searching and Filtering Records](#searching-and-filtering-records)
  * [Saving Custom Search](#saving-custom-search)
  * [Logs Description](#logs-description)
    * [Log Fields](#log-fields)
    * [List of OOB streams](#list-of-oob-streams)
  * [Cleaning Logs](#cleaning-logs)
  * [Exporting Logs](#exporting-logs)
    * [Exporting via WEB-Interface](#exporting-via-web-interface)
  * [Authenticating and Managing Users](#authenticating-and-managing-users)
    * [Roles](#roles)
    * [Users](#users)
    * [LDAP integration](#ldap-integration)
  * [Audit Logs](#audit-logs)
  * [System Logs](#system-logs)
  * [Integration Logs](#integration-logs)
  * [Access Logs](#access-logs)
  * [Ingress Nginx Logs](#ingress-nginx-logs)
  * [Sending Logs to External Destination (SIEM)](#sending-logs-to-external-destination-siem)
  * [Logging Application Backuper](#logging-application-backuper)
    * [Running Backup Process](#running-backup-process)
    * [Running Restore Process](#running-restore-process)
    * [Changing Password](#changing-password)
  * [Graylog Plugins System](#graylog-plugins-system)
    * [Graylog Obfuscation Plugin](#graylog-obfuscation-plugin)
  * [Graylog Audit logs](#graylog-audit-logs)
    * [Login events](#login-events)
    * [Logout events](#logout-events)
    * [Change password events](#change-password-events)

# Using FluentD/FluentBit with compression

`logging-operator` uses GELF protocol to send logs over TCP protocol to Graylog.
When using TCP protocol, typically the network usage will be high and CPU usage will be low.
Whenever there is a need to reduce network usage, UDP protocol can be used.

**Note:** This will increase the CPU load.

## FluentBit

[FluentBit](https://docs.fluentbit.io/manual/pipeline/outputs/gelf) offers using GELF over UDP protocol
and it automatically compresses logs. Graylog already has a feature to automatically decompress
logs when it receives data over GELF UDP.

It is necessary to set `.Values.fluentbit.graylogProtocol` with the value `udp` to enable
compression and transferring the logs over UDP.

## FluentD

FluentD does not offer default mechanism to compress logs when GELF is used as output plugin.
FluentD has a [buffer](https://docs.fluentd.org/configuration/buffer-section) section which provides an ability
to compress data before writing data to buffer chunks.

It is necessary to set `.Values.fluentd.compress` with the value `gzip` to enable compression before
writing data to buffer chunks.

# Using Graylog

You can use Graylog to view the logs for services, search and filter logs, and so on as described in the sections below.

## Viewing Logs of a Specific Microservice

You can view logs of a specific microservice using the following methods:

* Using the **Search** tab
* Creating streams

A stream is a mechanism to filter messages by some criteria in real-time while they are processed.
For example, you can add the application name as filtering criteria and view logs of only the required microservice.

To view logs of a specific microservice use the **Search** tab:

1. Navigate to the **Search** tab.
2. Enter your search query. For example, “labels-app:your_microservice_name” or “labels-app:monitoring-collector”

For more information about Graylog search query language, refer to
[Writing Search Queries](https://go2docs.graylog.org/5-2/making_sense_of_your_log_data/writing_search_queries.html).

## Create Stream

To create a new stream:

1. Navigate to the **Streams** tab
2. Click **Create Stream**
3. Specify the required parameters
4. Click **Save**

    ![Create Stream](/docs/images/graylog/create-streame.png)

5. Click **Manage Rules** for your stream
6. For example, you can upload one message
7. Configure the required rules.
8. Click **Start Stream** and check the results

For more information about streams, refer to
[Streams](https://go2docs.graylog.org/5-2/making_sense_of_your_log_data/streams.html).

## Searching and Filtering Records

To create searching and filtering logs, navigate to the **Search** tab and enter your search criteria.
The search criteria can be built from any combination of fields in a log message according to the Graylog search query language.

For more information about Graylog search query language, refer to
[Writing Search Queries](https://go2docs.graylog.org/5-2/making_sense_of_your_log_data/writing_search_queries.html).

For example, to see all the messages from `logging-fluentd` pods that contain the word `info`, enter the following query:

```bash
container: logging-fluentd and message: info
```

Alternatively, you can create a stream that contains only the logs that satisfy your filtering conditions.

For more information about creating streams, refer to
[Streams](https://go2docs.graylog.org/5-2/making_sense_of_your_log_data/streams.html).

The search criteria parameters are as follows:

* `container_name` specifies the Docker container name.
* `host/hostname` specifies the name of the OpenShift compute node.
* `labels-app` specifies the microservice name.
* `message` specifies the message details.
* `namespace_name` specifies the project name.
* `pod_name` specifies the pod name.
* `time/timestamp` specifies the timestamp.

## Saving Custom Search

You can save search criteria as a named query for quick usage in the future.

After the search criteria are set, click **Save**, enter a name for the current query, and save it.
The saved search is available by the **Load** button for quick access.

![Saving custom search](/docs/images/graylog/saved-search.png)

Custom saved searches can be added automatically after the Graylog deployment through REST API.

The payload for creating search example:

```json
{
  "id": "5f33c6369de1b46a5aa085c3",
  "queries": [
    {
      "id": "0af939f7-0ad5-47a0-ab99-7878eb144e79",
      "timerange": {
        "type": "relative",
        "range": 300
      },
      "query": {
        "type": "elasticsearch",
        "query_string": "message:CEF AND message:(LOGIN LOGOUT FORCED_LOGOUT SESSION_TIMEOUT)"
      }
    }
  ]
}
```

Endpoint: POST `https://GRAYLOG_HOST/api/views/search`

The payload for creating view example:

```json
{
  "id": "5f33c6669de1b46a5aa085c4",
  "type": "SEARCH",
  "title": "User session history",
  "summary": "",
  "description": "",
  "search_id": "5f33c6369de1b46a5aa085c3"
}
```

Endpoint: POST `https://GRAYLOG_HOST/api/views`

Curl example:

```json
curl --location --request POST 'https://GRAYLOG_HOST/api/api/views/search' \
--header 'Content-Type: application/json' \
--header 'X-Requested-By: cli' \
--header 'Authorization: Basic YWRtaW46YWRtaW4=' \
-k \
--data '{
  "id": "5f33c6369de1b46a5aa085c3",
  "queries": [
    {
      "id": "0af939f7-0ad5-47a0-ab99-7878eb144e79",
      "timerange": {
        "type": "relative",
        "range": 300
      },
      "query": {
        "type": "elasticsearch",
        "query_string": "message:CEF AND message:(LOGIN LOGOUT FORCED_LOGOUT SESSION_TIMEOUT)"
      }
    }
  ]
}'
```

## Logs Description

This section provides detailed descriptions of logs collected in Graylog OOB.

### Log Fields

The list of log fields is as follows:

* `application_id` specifies the unique application ID in OpenShift. It is usually `project_name + POD_name`.
* `container_name` specifies the docker container name.
* `docker` specifies the docker container UUID.
* `facility` specifies the logs' collector type. It is usually `fluentd`.
* `host/hostname` specifies the hostname of the logs source. For example, OpenShift node name.
* `kubernetes` specifies the technical meta-information from Kubernetes.
* `labels-app` specifies the Kubernetes application name.
* `labels-deployment` specifies the Kubernetes deployment name.
* `labels-deploymentconfig` specifies the Kubernetes deployment config name.
* `labels-site` specifies the OpenShift site name. This is applicable only for DR schema.
* `level` specifies the logging level.
* `message` specifies the original log message.
* `namespace_id` specifies the OpenShift project UUID.
* `namespace_name` specifies the OpenShift project name.
* `pod_id` specifies the OpenShift POD UUID.
* `pod_name` specifies the OpenShift POD name.
* `source` specifies the name of fluentd POD, which sent the log to Graylog.
* `tag` specifies the internal fluentd tag of the log.
* `time` specifies the time of logs printing by the application.
* `timestamp` specifies the timestamp when the log comes to Graylog. It can change with time due to network lag.

All other fields are technical and can be ignored.

### List of OOB streams

The list of OOB streams is as follows:

* `Application logs` include `stdout` from docker containers running in Kubernetes (PODs)
* `System logs` include `/var/log/messages`, `/var/log/syslog` and `systemd` from Kubernetes nodes
  and other infrastructure VMs such as Logging VM, Monitoring VM, and so on.
* `Audit logs` include `/var/log/audit/audit.log` from Kubernetes nodes and other infrastructure VMs
  such as Logging VM, Monitoring VM, and so on; `/var/log/ocp-audit.log` from Kubernetes nodes;
  logs of Graylog docker containers; and application's audit logs.
* `Integration logs` include logs with marker `[logType=int]` or `[log_type=int]`.
* `Kubernetes events` include logs with key-value `kind=KubernetesEvent`. The stream includes Kubernetes events sent as
  logs to Graylog form cloud-events-reader.

## Cleaning Logs

Graylog uses Elasticsearch indices that are grouped into Graylog index sets to store log messages.
ach index set has its settings, which include index rotation and data retention.

Graylog contains the following index rotation policies:

* Message count
* Index size on disk (default)
* Index time (rotation goes after a specific time)

Graylog contains the following retention policies:

* Delete index (Default)
* Close index (Makes index unavailable for RW operations without deleting the logs. It is possible to restore a closed index.)
* Do nothing

By default,  Graylog is installed with the following retention politics:

| Log stream name | Retention policy           |
| --------------- | -------------------------- |
| All messages    | Index size on disk = 20 Gb |
| Audit logs      | Index size on disk = 5 Gb  |

To view and edit the policies:

1. Navigate to the **System->Indices** tab.
2. Click the required index set.
3. Click **Edit Index Set**.
4. Edit the required parameters.
5. Click **Save**.

The configuration for the rotation policy is shown in the following image.

![Configure Rotation Policy](/docs/images/graylog/configure-rotation-policy.png)

## Exporting Logs

The logs can be exported as described in the following sections.

### Exporting via WEB-Interface

You can export the results of a search request to a CSV format through the Web interface.

To export a search request:

1. Perform a search.
2. From the **...** button, click **Export to CSV**.

## Authenticating and Managing Users

The information on authentication and managing users is as follows:

### Roles

By default, there are next user roles available:

* Administrator - The Administrator has full rights on the system.
* Reader - The Reader has only ReadOnly rights.
* Operator - The Operator has ReadOnly rights for all streams on the system except of 'Audit logs'.
* AuditViewer - The AuditViewer has ReadOnly rights for all streams on the system.

![Default Roles](/docs/images/graylog/default-roles.png)

You can also create new user roles.

To create a new user role:

1. From the **System/Authentication** tab, click **Roles**.
2. Click **Add new role**.
3. Specify the required parameters.
4. Specify the R/W rights for each created stream by selecting the **Allow reading** or **Allow editing** option as appropriate.

A role is created as shown in the following image.

![Create Roles](/docs/images/graylog/create-roles.png)

### Users

By default, there are the following users:

* Admin with Administrator role.
* Operator with Reader and Operator role.
* AuditViewer with Reader and AuditViewer role.

The default users are shown in the following image.

![Default Users](/docs/images/graylog/default-users.png)

You can create a new user in the system.

To create a new user:

1. From the **System/Authentication** tab, click **Users**.
2. Click **Add new user**.
3. Specify the required parameters.
4. Click **Create User**.

A user is created as shown in the following image.

![Create Users](/docs/images/graylog/create-users.png)

For more information about Graylog search query language, refer to
[Export Results as CSV](https://go2docs.graylog.org/5-2/interacting_with_your_log_data/export_results_as_csv.html)

### LDAP integration

To configure LDAP/AD integration, refer to the official Graylog documentation:
[Authentication](https://go2docs.graylog.org/5-2/setting_up_graylog/permission_management.html?Highlight=ldap#Authentication)

**Note:**

* Graylog OOB has several system users, such as admin and operator. If these users are required to be in LDAP as well,
  they must be added manually.
* Graylog has a known issue in the LDAP authentication provider - extra calls to the LDAP server can take place.
  Refer to the official issue in the following link
  [https://github.com/Graylog2/graylog2-server/issues/2267](https://github.com/Graylog2/graylog2-server/issues/2267)

## Audit Logs

Sensitive or audit logs require special processing in terms of storing time and viewing permissions.
Such processing is delivered to OOB.

A special stream named `Audit logs` contains various audit logs. It has restricted permissions
(only an administrator can view it), and a separate index set.
The separate index set, the `Audit index set`, is responsible for the rotation of audit logs.
By default, the rotation size is 5GB. It can be manually changed after deployment according to the project's needs.

This stream has several OOB criteria, by which the platform audit logs are collected into the Audit stream.

OOB criteria cover the following audit logs:

* **/var/log/audit/audit.log** from VMs such as Kubernetes nodes, Logging VM, Monitoring VM, and so on.
* Graylog audit logs
* Kubernetes audit logs
* Audit logs such as PostgreSQL audit, and so on.

If you want to add custom audit logs into this stream, define the criteria by which Graylog can filter the audit logs.

It is recommended to write audit logs that combine `log_id` with the value `audit`.
Generally, more complicated and tricky criteria for audit logs can take place, depending on your application.
An example of using `log_id` with the value `audit` is given in the following image.

![Stream Rule](/docs/images/graylog/stream-rule.png)

## System Logs

Operation system logs from platform VMs such as Kubernetes nodes, Logging VM, Monitoring VM, and so on,
are collected into a separate stream, `System logs`, which is provided to OOB.

This stream contains `/var/log/messages` logs from the VMs.
The stream logs are stored in the same `Default index set` as any other typical logs.

## Integration Logs

The stream contains all logs with one of the following markers:

* `[logType=int]`
* `[log_type=int]`

## Access Logs

The stream contains all logs with one of the following markers:

* `[logType=access]`
* `[log_type=access]`

## Ingress Nginx Logs

The stream contains all logs from pods `ingress-nginx`.

## Sending Logs to External Destination (SIEM)

The following `output` protocols are supported:

* `GELF`
* `SYSLOG`

To configure logs' forwarding to an external system, perform the following steps:

1. Navigate to `System -> Outputs`
2. Select the required format from the drop-down list
3. Click **Launch new output**. The **Launch New Output** popup dialog is displayed
4. Specify the required parameters
5. Click **Save**

The configure logs parameters are as follows.

* Title
* Destination host
* Transport protocol (TCP/UDP)
* Destination port (make sure this port is opened for Graylog container)

## Logging Application Backuper

Logging-backuper allows you to make backups of all logging applications such as Graylog, OpenSearch, MongoDB, and so on,
and to restore if some exceptional case occurs.
Logging-backuper is installed on the Graylog host machine by the deploy-logging-backuper job.

### Running Backup Process

To run the backup process:

1. Navigate to `GRAYLOG_HOST` by ssh.
1. Execute the command, `curl -XPOST localhost:8080/backup`.

### Running Restore Process

To run the restore process:

1. Navigate to the Graylog folder in root home, `sudo -i && cd ~/graylog`.
2. Get the list of available backups by executing the command, `curl -XGET localhost:8080/listbackups`.
   For example, `20180809T102744 (date and time of backup)`.
3. Use one of the backup names and run `./restore-logging.sh #{backup_name_here}`. For example, `./restore-logging.sh 20180809T102744`.
4. Wait for the `restore successful` message.

### Changing Password

To change the password locally:

1. Navigate to the graylog folder in root home, `sudo -i` then `cd ~/graylog`
2. Use the `change-password.sh` script to change the password with the following parameters:

   * `--user (-u)` - To specify the user name
   * `--oldpass (-op)` - To specify the current password
   * `--newpass (-np)` - To specify the new password
   * `--confirmpass (-cp)` - To confirm the new password

3. Wait for the `[INFO] Password was successfully changed` message

To change the password remotely:

1. Copy scripts `change-password.sh` and `processing-change-password.sh` to your machine
2. Use `change-password.sh` script to change the password with the following parameters:

   * `--hosts (-h)` - To specify hosts
   * `--user (-u)` - To specify the user name
   * `--oldpass (-op)` - To specify the current password
   * `--newpass (-np)` - To specify the new password
   * `--confirmpass (-cp)` - To confirm the new password
   * `--sshuser` - To specify ssh login
   * `--sshkey` - To specify the path for ssh key of hosts

3. Wait for `[INFO] Password was successfully changed` message

**Note:** If the external destination is unreachable, then the logs are not forwarded to it until it is active.

## Graylog Plugins System

Graylog supports custom extensions using plugins.
To configure plugins, navigate to **System/Configurations > Configurations** as shown in the following image.

![Path to Configuration Page](/docs/images/graylog/configuration-page.png)

### Graylog Obfuscation Plugin

The Graylog Obfuscation Plugin is necessary for anonymization of sensitive data in the logs.
For more information, see [Graylog Obfuscation Plugin](graylog_obfuscation_plugin.md).

## Graylog Audit logs

The Graylog has audit logs on three security events. The information about that event is taken from access logs.
The graylog audit logs can be found in `Audit logs` stream for `graylog_graylog_1` container name.

The format of access logs:

```bash
[time][level][ip address][user][access url][user agent][status]
```

### Login events

The event on login consists of two log record, where in the first record describe the fact of login
and in second record describe who perform login to Graylog. On success login used 200 http code, on failed login
used 401 http code with the additional log record.

Examples:

```bash
[2020-09-16T08:48:30,488][INFO]Invalid credentials in session create request. Actor: "urn:graylog:user:admin"
```

```bash
[2020-09-16T07:46:13,069][DEBUG]0.0.0.0 - [-] "POST api/system/sessions" Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 200 -1
```

```bash
[2020-09-16T07:46:13,272][DEBUG]0.0.0.0 admin [-] "GET api/users/admin" Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 200 -1
```

### Logout events

Examples:

```bash
[2020-09-16T07:46:07,845][DEBUG]0.0.0.0 admin [-] "DELETE api/system/sessions/cef0692d-1154-4bb4-ba48-1f9abda83c1c" Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 204 -1
```

### Change password events

Examples:

```bash
[2020-09-16T08:37:53,746][DEBUG]0.0.0.0 admin [-] "PUT api/users/LogReader/password" Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 204 -1
```
