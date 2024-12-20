This section describes the common problems while connecting to Graylog and the ways to troubleshoot them.

# Table of Content

* [Table of Content](#table-of-content)
* [Graylog](#graylog)
  * [Problems with Connection to Graylog](#problems-with-connection-to-graylog)
    * [Unable to Connect to Graylog via Browser](#unable-to-connect-to-graylog-via-browser)
    * [Unable to Read Log Messages](#unable-to-read-log-messages)
    * [Ingress/Route to Graylog cyclic redirect](#ingressroute-to-graylog-cyclic-redirect)
  * [Typical Issues](#typical-issues)
    * [HDD Full on Graylog VM](#hdd-full-on-graylog-vm)
    * [Graylog Container OOM Killed (out of RAM)](#graylog-container-oom-killed-out-of-ram)
    * [Low Graylog Performance](#low-graylog-performance)
    * [Graylog Not Processing Messages](#graylog-not-processing-messages)
    * [Index Oversized](#index-oversized)
    * [Negative number of Unprocessed Messages](#negative-number-of-unprocessed-messages)
    * [Incorrect timestamps in Graylog](#incorrect-timestamps-in-graylog)
    * [Information about OpenSearch nodes is unavailable](#information-about-opensearch-nodes-is-unavailable)
    * [Widgets do not show data with errors](#widgets-do-not-show-data-with-errors)
    * [Deflector exists as an index and is not an alias](#deflector-exists-as-an-index-and-is-not-an-alias)
  * [Performance tuning](#performance-tuning)
    * [Typical symptoms of performance issues and common words](#typical-symptoms-of-performance-issues-and-common-words)
    * [Common performance principles](#common-performance-principles)
  * [Extra tips and tricks](#extra-tips-and-tricks)
    * [/srv/docker/graylog/graylog/config/graylog.conf](#srvdockergrayloggraylogconfiggraylogconf)
    * [Crackdown for heavy loads](#crackdown-for-heavy-loads)
* [OpenSearch](#opensearch)
  * [Limit of total fields has been exceeded](#limit-of-total-fields-has-been-exceeded)
  * [Errors `no such index [.opendistro-ism-config]`](#errors-no-such-index-opendistro-ism-config)
  * [OpenSearch uses more than 32 GB of RAM](#opensearch-uses-more-than-32-gb-of-ram)
  * [Index read-only Warnings](#index-read-only-warnings)
* [FluentD](#fluentd)
  * [FluentD worker killed and restart with SIGKILL](#fluentd-worker-killed-and-restart-with-sigkill)
  * [FluentD generate a high DiskIO read load](#fluentd-generate-a-high-diskio-read-load)
  * [FluentD failed to flush buffer, data too big](#fluentd-failed-to-flush-buffer-data-too-big)
* [FluentBit](#fluentbit)
  * [Connection timeout to Graylog in FluentBit](#connection-timeout-to-graylog-in-fluentbit)
  * [FluentBit stuck and stopped sending logs to Graylog](#fluentbit-stuck-and-stopped-sending-logs-to-graylog)
* [ConfigMap Reloader](#configmap-reloader)
  * [Fluent container restarts after changing ConfigMap](#fluent-container-restarts-after-changing-configmap)

# Graylog

## Problems with Connection to Graylog

There are some problems that you may face while connecting to Graylog. These are explained below.

### Unable to Connect to Graylog via Browser

To identify the root cause:

Connect via SSH to the virtual machine with Graylog deployed.

To see the information about `STATUS` of Graylog Docker containers, execute the following command.

```bash
docker ps -f "name=graylog"
```

The normal output contains four running containers with `STATUS` field equal to "UP N days.".
The four running containers include `graylog_web_1`, `graylog_graylog_1`, `graylog_storage_1`, and `graylog_mongo_1`.

In case the output contains a lesser number of containers, or their status differs from the norm,
then try to restart the container with the problem.

If you are unable to connect to the virtual machine using SSH, check the network connection using the `ping` command.

[Back to TOC](#table-of-content)

### Unable to Read Log Messages

To check for errors, navigate to the **System > Overview** tab.

![graylog-system-overview](/docs/images/graylog/system-overview.png)

Navigate to the deployed FluentD (usually it is the "logging" project), and see the pods' health-check reports.

[Back to TOC](#table-of-content)

### Ingress/Route to Graylog cyclic redirect

Applicable for DR no-vIP schema only.

In this schema Logging service deploy procedure creates an external service in OpenShift.
By accessing this external service via OpenShift coordinates `graylog.logging.svc.cluster.local` other
applications can work with active Graylog instances.

Also, we created a Route for accessing of active Graylog Web UI. If OpenShift contains separate Load Balancers
with HTTPS certificates on them, this route will not work. It returns 302 (redirect) to itself,
getting an infinite loop.

To fix it manual actions are required. Route URL needs to be added into _os_sni_passthrough.map_ file on Load Balancers.

[Back to TOC](#table-of-content)

## Typical Issues

The typical issues that you may face are given below.

### HDD Full on Graylog VM

**Symptoms:**

* Graylog does not process any new messages.
* Search in logs shows various errors (for example, HTTP 500).
* OpenSearch is down, container constantly restarting.

**How to check:**

1. Log in to Graylog VM via SSH and execute `df -h`.
2. It shows you the information about the HDD utilization.

**How to fix:**

1. Log in to Graylog VM via SSH as root.
2. Execute the following commands:

   ```bash
   docker stop \
        graylog_web_1 \
        graylog_graylog_1 \
        graylog_storage_1 \
        graylog_mongo_1

   rm -rf /srv/docker/graylog/opensearch/nodes/

   docker start \
         graylog_mongo_1 \
         graylog_storage_1 \
         graylog_graylog_1 \
         graylog_web_1
   ```

3. If after cleaning up disk space you see 'index read-only' warnings in the Graylog Web UI,
   execute the following command on Graylog VM via SSH to unlock the index:

   ```bash
   curl -X PUT -u <username>:<password> -H "Content-Type: application/json" -d '{"index.blocks.read_only_allow_delete": null}' http://localhost:9200/_settings
   ```

   or command for usage in cloud Graylog:

   ```bash
   curl -X GET -u <username>:<password> -H "Content-Type: application/json" -d '{"index.blocks.read_only_allow_delete": null}' opensearch.opensearch-service:9200/_settings
   ```

   **Note**: All the existing logs are lost. To prevent it in the future, you need to adjust the indices
   rotation policy in Graylog according to the available HDD size. You can restore old logs from the backup.

   For checking blocked indices use the next command:

   ```bash
   curl -X GET -u <username>:<password> -H "Content-Type: application/json" opensearch.opensearch-graylog:9200/index_name
   ```

   If some index has `"read_only_allow_delete": "true"` it means that this index is blocked and new data
   can't be store in it. So you should unlock this index.

[Back to TOC](#table-of-content)

### Graylog Container OOM Killed (out of RAM)

**Symptoms:**

Graylog Web UI is not accessible or displays a 504 error.

**How to check:**

1. Login on Graylog VM via SSH and execute `docker ps`
2. It shows the container's status. If Graylog/opensearch containers are constantly restarting
   then it could be memory related issue

**How to fix:**

You can use any one of the following options to change the memory settings:

**Using Jenkins job:**

Run the redeploy of the Logging service procedure with the corrected `graylog_heap_size` and `es_heap_size` parameters.

**In manual mode:**

1. Log in to Graylog VM via SSH as root.
2. Execute the following command and remember the ID of the container with Graylog:

   ```bash
   docker inspect --format '{{.Id}}' graylog_graylog_1
   ```

3. Execute the following command and remember the ID of the container with OpenSearch:

   ```bash
   docker inspect --format '{{.Id}}' graylog_storage_1
   ```

4. Stop the Docker service using the following command:

   ```bash
   service docker stop
   ```

5. Change the memory parameter for the container with Graylog:
   In the `/var/lib/docker/containers/<container_id>/config.v2.json` file, find the `GRAYLOG_SERVER_JAVA_OPTS`
   parameter and correct its value.

   For example, it was 2GB:

   ```bash
   GRAYLOG_SERVER_JAVA_OPTS = -Xms2048m -Xmx2048m
   ```

   Corrected to 4GB:

   ```bash
   GRAYLOG_SERVER_JAVA_OPTS = -Xms4096m -Xmx4096m
   ```

6. Change the memory parameter for the container with OpenSearch:

   In the `/var/lib/docker/containers/<container_id>/config.v2.json` file, find the `ES_JAVA_OPTS`
   parameter and correct its value.

   By analogy with Graylog (step 5).

7. Start the docker and restart the containers:

   ```bash
   service docker start

   docker restart \
        graylog_web_1 \
        graylog_graylog_1 \
        graylog_storage_1 \
        graylog_mongo_1

   ```

[Back to TOC](#table-of-content)

### Low Graylog Performance

**Symptoms:**

1. Graylog Web UI is very slow
2. Graylog doesn't show any messages in search within the last 5-15 minutes
3. There is a notification "Journal utilization is too high" in the UI

**How to check:**

1. Log in to Graylog VM via SSH as root
2. Navigate to `top`
3. Check the resource consumption. You can also check resource consumption using System Monitoring if available.
   Define what kind of resource (CPU/RAM/HDD IOPS) is not enough according to the
   [documentation](/docs/installation.md#hwe-and-limits).

**How to fix:**

Add missing resources to the target VM.

See the section [Performance tuning](#performance-tuning) for additional information
about Graylog's tiny performance tuning aspects

Restart Graylog by executing the following commands:

```bash
docker restart \
    graylog_web_1 \
    graylog_graylog_1 \
    graylog_storage_1 \
    graylog_mongo_1
```

After restart go to `/system/inputs` in the Graylog UI and stop input messages by button `Stop input`.
This helps to prevent repeated Graylog flooding.

Go to detailed information about node by /system/nodes and button `Details`.

Wait for the input buffer to be freed. This will mean that Graylog has processed the messages.

Wait for journal utilization will reduce to values 0-5%. After that, you can run input.

[Back to TOC](#table-of-content)

### Graylog Not Processing Messages

**Symptoms:**

* New logs are not available for search
* Search does not work at all

**How to check:**

1. Navigate to `http://<graylog>/system/nodes`.
2. Check `The journal contains X unprocessed messages`.
3. If `X` is high (> 100000) and keeps growing, it is an issue.

**How to fix:**

Root cause: OpenSearch does not take payload.

Possible reasons and solutions:

* [HDD Full on Graylog VM](#hdd-full-on-graylog-vm)
* [Graylog container OOM killed (out of RAM)](#graylog-container-oom-killed-out-of-ram)
* [Low Graylog Performance](#low-graylog-performance)
* OpenSearch issue. Restarting the containers can help in this case. For more information, see [Low Graylog performance](#low-graylog-performance).

### Index Oversized

**Symptoms:**

* The HDD space utilization on the Logging VM is high. It exceeds the maximum possible utilization configured
  in the indices rotation policies.
* The size of one of the indices in OpenSearch is very big, more than what is configured
  in the `Max index size` parameter on the Index Set configuration.

You can check the indices size using the following command on Logging VM:

```bash
curl -X GET -u <username>:<password> -sk https://localhost/api/system/indexer/indices
```

**Root cause:**

Graylog indexer bug. It is a rare cause. A manual workaround can be applied if this issue occurs.

**How to fix:**

**Note**: Take a backup prior to deleting.

Delete an oversized index manually by executing the following command on the Logging VM:

```bash
curl -X DELETE -u <username>:<password> -H "X-Requested-By: graylog" https://localhost/api/system/indexer/indices/<index name>
```

[Back to TOC](#table-of-content)

### Negative number of Unprocessed Messages

If you have a negative number of unprocessed messages in the `Disk Journal` section it means that
you clean the journal directory but not completely.

**How to fix:**

Stop Graylog container:

```bash
docker stop graylog_graylog_1
```

Completely remove the directory:

```bash
{{ graylog_volume }}/graylog/data/journal/*
```

where `{{ graylog_volume }}` by default has the value `/srv/docker/graylog`, so to remove you need to execute a command:

```bash
rm -rf /srv/docker/graylog/graylog/data/journal/*
```

Start Graylog container:

```bash
docker start graylog_graylog_1
```

If you'd like to switch off the journal messages, you should also update `/srv/docker/graylog/graylog/config/graylog.conf`
and set parameter `message_journal_enabled=false`.

[Back to TOC](#table-of-content)

### Incorrect timestamps in Graylog

If you have different time values (time zones) in the `message`,  the `time`, and the `timestamp` fields,
need to check the timezone on nodes. The timezone must be set to UTC on each node.

Or you can change the timezone in the user settings in the Graylog to the timezone that is set on the nodes,
but this will not change the time inside the `message` field  (it will be equal UTC timezone).

[Back to TOC](#table-of-content)

### Information about OpenSearch nodes is unavailable

If you log in to Graylog UI, go to `System -> Nodes` and see that info about Elastic nodes is unavailable:

![Node info is unavailable](/docs/images/graylog/wrong-certificate-nodes-info.png)

Then, if you click on the node's name (`44a226cb/graylog-0` from the example above), you'll probably face an error like
this:

![Unavailable node details](/docs/images/graylog/wrong-certificate-details.png)

In this case, you should check that your Graylog's TLS certificate is not expired and contains valid alt names (e.g.
it must contain `graylog-service.logging.svc` if your Graylog is deployed into the `logging` namespace in the Cloud).

If you use a self-signed certificate,
[the article about certificate generation](user-guides/tls.md#self-signed-certificate-generation) can be useful for you.

[Back to TOC](#table-of-content)

### Widgets do not show data with errors

In case of problems with indexes in OpenSearch Graylog can show errors on the widgets.

For example with messages:

<!-- markdownlint-disable line-length -->
```bash
While retrieving data for this widget, the following error(s) occurred:

Unable to perform search query: Elasticsearch exception [
  type=illegal_argument_exception,
  reason=Text fields are not optimised for operations that require per-document field data like aggregations and sorting, so these operations are disabled by default. Please use a keyword field instead. Alternatively, set fielddata=true on [timestamp] in order to load field data by uninverting the inverted index. Note that this can use significant memory.
].
```
<!-- markdownlint-enable line-length -->

Also, in the Graylog logs you can see a similar error:

<!-- markdownlint-disable line-length -->
```bash
type=illegal_argument_exception,
reason=Text fields are not optimised for operations that require per-document field data like aggregations and sorting, so these operations are disabled by default. Please use a keyword field instead. Alternatively, set fielddata=true on [timestamp] in order to load field data by uninverting the inverted index. Note that this can use significant memory.
```
<!-- markdownlint-enable line-length -->

This error usually occurs when:

* Created custom OpenSearch index
* Created a Stream that routes messages in custom OpenSearch index

Created custom OpenSearch index may have fields declared with incorrect type or non-declared fields.
The second reason is most typical for custom indexes.

OpenSearch has a dynamic typing and a set of fields in the index. It means that OpenSearch
tries to automatically select a type for a new field if you didn't declare the field, and OpenSearch
receives a request to save data with this new field.

And selected type may not apply to Graylog. For example, Graylog can't use text fields to use them in sorting.

**Solution:**

Check the error and find which field has an incorrect type. For example, for the error above the problem field will be:

```bash
Alternatively, set fielddata=true on [timestamp] in ...
```

field with name `timestamp`.

Next, you have to check its type using requests to OpenSearch API. The following requests will help you:

* If you don't know index name or want to check the field type in all indexes:
  
    ```bash
    GET /_mapping/field/<field>
    ```

* If you know the index name:

    ```bash
    GET /<index_name>/_mapping/field/<field>
    ```

* If you want to check all index mapping:

    ```bash
    GET /_index_template/<index_name>
    ```

After that, you need to change your index mapping, declare the necessary field (if it wasn't declared)
and set the correct type. For example, if you are faced with an incorrect type to `timestamp` field you need to use
the `date` type for this field.

**How to avoid this issue:**

You have to remember about dynamic typing and declare all fields for custom OpenSearch indexes.

[Back to TOC](#table-of-content)

### Deflector exists as an index and is not an alias

Graylog uses a special OpenSearch alias to write and read logs always in the last index. This alias has
a postfix `_deflector` and it is managed by Graylog.

If Graylog detects that OpenSearch already has the index with a name:

```bash
<index_name>_deflector
```

it will raise the error in the UI (you can see it on the Overview page):

```bash
Deflector exists as an index and is not an alias
```

This problem may occur in two cases:

* Somebody manually created an index in OpenSearch with the name that Graylog wants to use as an alias
* During the update, you faced the following scenario:
  * Graylog is working and can receive logs
  * Agents active and send logs
  * Stream is already created, but mapped on non-existing Index
  * Index (that should store data from the Stream above) does not exist

In the last case, OpenSearch can receive a request to save data before Graylog creates the index and assigns
the deflector alias to it. You can understand and verify it by Graylog and OpenSearch logs.
For example:

* Graylog logs:

    ```bash
    [2023-10-26T12:49:12,327][WARN]Active write index for index set "v2_cis_inventory_change_log" (653a6047ab6c072bb306a2d5) doesn't exist yet
    ```

* OpenSearch logs:

    ```bash
    [2023-10-26T12:49:12,391][INFO ][o.o.c.m.MetadataCreateIndexService] [604eb8d3c4b3] [v2_cis_inventory_change_log_deflector] creating index, cause [auto(bulk api)], templates [v2_cis_inventory_change_log], shards [1]/[1]
    [2023-10-26T12:49:12,839][INFO ][o.o.c.m.MetadataMappingService] [604eb8d3c4b3] [v2_cis_inventory_change_log_deflector/3_kIpr9zQYunZMeZgumPVA] update_mapping [_doc]
    ```

**Solution:**

If you manually create the index with such a name, you have to remove it. And do not try to use such a name in the future.

If you are faced with such a problem during the update of the Logging VM it means that before the update
you must **disable all Graylog Inputs**.

To do it you need:

* Open Graylog UI
* Navigate to `System -> Inputs`
* Click on the button `Stop input` for each input

After upgrade will be successfully complete you can start all inputs.

**How to avoid this issue:**

You shouldn't create indexes with postfix `_deflector` and use it as an alias. It's a reserved alias by Graylog.

During updates that should be created Streams that use custom indexes, you must stop all Graylog Inputs.

[Back to TOC](#table-of-content)

## Performance tuning

### Typical symptoms of performance issues and common words

Graylog uses OpenSearch as backend storage for log data. Graylog itself acts as an incoming logs receiver and processor.
Graylog does not require many resources and in regular operations, it cannot be overloaded.
In most cases OpenSearch is a bottleneck - it cannot receive all logs from Graylog because of
a lack of resources.

OpenSearch is Disk speed greedy at first and RAM greedy at second.

If OpenSearch cannot handle all incoming log data - Graylog buffers grow, including disk journal.
Graylog began to utilize disk and CPU for serving journals which slowed down OpenSearch more and more.
As a result, the system falls into an unstable state.

The symptoms (from small overload to significant overload):

1. Low performance of search operations in Graylog
2. Graylog journal grows. Journal size 0-50k messages if fine. 50k-100k is worth. 500k+ is almost a disaster
3. Logs search does not show recent logs (because they are in Graylog's journal, not in OpenSearch)
4. Graylog UI slowness, random 500 and 503 errors
5. Graylog UI is down
6. Graylog VM CPU is fully utilized, VM became unresponsive even via ssh

[Back to TOC](#table-of-content)

### Common performance principles

* First of all, check the hardware resources of your Graylog instance according to the [table](/docs/installation.md#hwe-and-limits).
  The most important thing is disk speed and almost all performance issues can be solved by increasing it.
* Use `sysbench` to measure disk speed
* RAM and CPU are the second priority but it is also important
* Graylog does not require much RAM. 4-8 GB is enough. Better give more RAM to OpenSearch

[Back to TOC](#table-of-content)

## Extra tips and tricks

### /srv/docker/graylog/graylog/config/graylog.conf

* `processbuffer_processors`, `outputbuffer_processors` - set to CPU count / 2.
* `ring_size` - set to 131072 or to 262144 if you have 4+ RAM for Graylog. Higher values are not recommended

[Back to TOC](#table-of-content)

### Crackdown for heavy loads

* Remove the `Logs Routing` pipeline from Graylog. It will save the CPU, but logs routing to streams will be lost.
* Disable disk journal in Graylog to prevent disk concurrency between Graylog and OpenSearch.
* Disable collection of system and audit-system logs on the FluentD side

[Back to TOC](#table-of-content)

# OpenSearch

## Limit of total fields has been exceeded

**Symptoms:**

In OpenSearch's or Graylog's logs (or responses on API calls) generated one or some errors like:

```bash
Limit of total fields [1000] in index [test_index] has been exceeded
```

**Root cause:**

OpenSearch has a mechanism to prevent mapping explosions (too many dynamical fields):

* [https://www.elastic.co/guide/en/elasticsearch/reference/master/mapping.html#mapping-limit-settings](https://www.elastic.co/guide/en/elasticsearch/reference/master/mapping.html#mapping-limit-settings)
* [https://www.elastic.co/guide/en/elasticsearch/reference/master/mapping-settings-limit.html](https://www.elastic.co/guide/en/elasticsearch/reference/master/mapping-settings-limit.html)
* [https://opensearch.org/docs/latest/field-types/#mapping-limit-settings](https://opensearch.org/docs/latest/field-types/#mapping-limit-settings)

By default OpenSearch/ElasticSearch doesn't allow to save new fields in the index after reach the limit in **1000** fields.

**How to fix:**

Usually this issue occurs due the incorrect work of agent that should parse logs and send them to Graylog.

FluentBit or FluentD has a logic to parse new dynamical fields from the log's `message` and add these fields as metadata
in logs sending to Graylog.

There are some issues in FluentBit and FluentD that could lead to parse parts of `message` as `key=value` pairs.
For example, from the log:

```bash
[2024-09-30T04:59:40.498] [DEBUG] [request_id=1a04d001-37e6-418b-bc7f-4904d4dfc753] [tenant_id=-] [thread=main-8e36d] [class=mongo:storage.go:236] [traceId=0000000000000000176d565380a60f8b] [spanId=04546e4d3320dc9b] try to delete objects from certificates by filter map[$and:[map[meta.status:map[$ne:trusted]] map[$or:[map[meta.deactivatedAt:map[$lte:2024-08-31 04:59:40.498507159 +0000 UTC m=+6199354.549415617]] map[details.validTo:map[$lte:2024-08-31 04:59:40.498507159 +0000 UTC m=+6199354.549415617]]]]]] 
```

expect `key=value` from the `message` part

```bash
[request_id=1a04d001-37e6-418b-bc7f-4904d4dfc753] [tenant_id=-] [thread=main-8e36d] [class=mongo:storage.go:236] [traceId=0000000000000000176d565380a60f8b] [spanId=04546e4d3320dc9b]
```

was parsed the `key=value` pair:

```bash
_lte_2024-08-31_04_59_40_498507159__0000_UTC_m = +6199354.549415617
```

So if you faced with such issues we recommended to update to latest version of Logging and check again.

In the case, if you use external agent or send logs directly to Graylog, need to check your agent settings or
service/application that send these logs.

**Useful script to clean `trash` fields:**

All early saved `trash` `key=value` pairs will removed with indices when rotation strategy decided to remove index.

But if you already fixed the root cause of issue due FluentBit or FluentD (or other agent) generated a lot of dynamic fields
you can use the script to remove already saved `trash` fields in indices:

```painless
List fieldsToRemove = new ArrayList();
for (entry in ctx._source.keySet()) {
  if (entry.startsWith('ErrorEntry_')) {
    fieldsToRemove.add(entry);
  }
}
for (field in fieldsToRemove) {
  ctx._source.remove(field);
}
```

where `ErrorEntry_` it's a prefix of fields that should be remove and that you need to change.

This script can be execute using OpenSearch API:

**Warning!** Pay attention that it may require a lot of time, and CPU resources because you need to update all indices
and docs in indices.

```bash
curl -X POST -u <username>:<password> -H 'Content-Type: application/json' http://localhost:9200/<index_name>/_update_by_query -d '{
  "query": {
    "match_all": {}
  },
  "script": {
    "lang": "painless",
    "source": "List fieldsToRemove = new ArrayList();\nfor (entry in ctx._source.keySet()) {\n  if (entry.startsWith(\"ErrorEntry_\")) {\n    fieldsToRemove.add(entry);\n  }\n}\nfor (field in fieldsToRemove) {\n  ctx._source.remove(field);\n}"
  }
}'
```

where `<index_name>` can contain some indices separated by a comma, or "*", see
[https://opensearch.org/docs/latest/api-reference/document-apis/update-by-query/#path-parameters](https://opensearch.org/docs/latest/api-reference/document-apis/update-by-query/#path-parameters).

**Note:** If the index will be locked on write you can unlock it using the command:

```bash
curl -X PUT -u <username>:<password> -H 'Content-Type: application/json' -d '{"index.blocks.write": "false"}' http://localhost:9200/<index_name>/_settings
```

## Errors `no such index [.opendistro-ism-config]`

**Symptoms:**

In OpenSearch's logs generated one or some errors like:

```bash
[2024-10-11T11:47:21,697][ERROR][o.o.i.i.ManagedIndexCoordinator] [881c8d26fd21] get managed-index failed: [.opendistro-ism-config] IndexNotFoundException[no such index [.opendistro-ism-config]]
```

**Root cause:**

This error message has no effect on your OpenSearch.

It is generated in the `index-management` plugin for OpenSearch. Someone already asked the community
about this error in the GitHub issue:
[https://github.com/opensearch-project/index-management/issues/697](https://github.com/opensearch-project/index-management/issues/697)

The quote from the GitHub issue about this error log:

> This is actually intended behavior for the ISM plugin, and should not have any negative impact.
> However, we agree that logging this exception as an "ERROR" is inappropriate for this use case.
> This occurs because the plugin listener picks up a "ClusterChangedEvent" whenever an index is deleted,
> as we want the plugin to then check for any plugin-specific metadata related to that index.
> However, since the ISM plugin had not been used yet in your example (e.g., an ISM policy hasn't been created),
> the `.opendistro-ism-config` index had not been created.

The plugin's authors changed the level from ERROR to DEBUG since version 2.10.0.0:

* [https://github.com/opensearch-project/index-management/releases/tag/2.10.0.0](https://github.com/opensearch-project/index-management/releases/tag/2.10.0.0)
* [https://github.com/opensearch-project/index-management/pull/846](https://github.com/opensearch-project/index-management/pull/846)

**How to fix:**

There are four theoretical ways to avoid/remove this error:

* Add at least one ISM rule in OpenSearch, in this case, OpenSearch should create the system index
  `.opendistro-ism-config` which by default doesn't exist
* Disable this plugin

  ```yaml
  plugins.index_state_management.enabled: False
  ```

  Official documentation: [https://opensearch.org/docs/latest/im-plugin/ism/settings/](https://opensearch.org/docs/latest/im-plugin/ism/settings/)

* Upgrade OpenSearch to `>=2.10.x` version
* Ignore this error message

The option with upgrade OpenSearch to 2.x version now is not available, it's just a theoretical option.

## OpenSearch uses more than 32 GB of RAM

In case of a high load to Graylog and OpenSearch, you may want to allocate more than ~32 GB RAM
for OpenSearch. After that, you may notice that OpenSearch performance got even worse.

For example, if become often fails with OOM or processes less throughput than with a memory limit less than ~32 GB.

It occurs because after you set `-Xmx` for OpenSearch to more than ~32 GB it starts to use 64-bit
Ordinary Object Pointers (OOP) instead of 32-bit pointers. As a result, it decreases memory efficiency.

The official OpenSearch documentation tells us that in fact, it takes until around **40–50 GB** of allocated heap
before you have the same effective memory of a heap just under **32 GB** using compressed oops.

OpenSearch/ElasticSearch documentation
[Don’t Cross 32 GB!](https://www.elastic.co/guide/en/elasticsearch/guide/current/heap-sizing.html#compressed_oops).

**How to fix:**

Usually, if you want to increase the memory limit for OpenSearch it means that you don't want
to do a deeper analysis of the problems. And you want just adding resources in hopes it will help.

To fix this issue you first of all need to decrease the memory limit for OpenSearch to ~32 GB.

Second, you need to remember that on the Logging VM, there are other applications and processes like Graylog,
MongoDB and Nginx. All these applications run as docker containers and require some memory to run. Also, Java
applications can use more than set in `-Xmx`. This happens because of the way Java handles memory.

Deployment scripts of the Logging VM allow to specify limits for Graylog and OpenSearch.
So you have to specify a summary of limits for Graylog and OpenSearch less than the total VM memory size.
Also, you need to remember that you must leave 20-50% of free RAM on the VM.

Some examples (it's not recommendations, just examples):

* Graylog VM has 16 GB RAM, in this case, you can allocate:
  * Graylog - 4 GB
  * OpenSearch - 8 GB
* Graylog VM has 32 GB RAM, in this case, you can allocate:
  * Graylog - 8 GB
  * OpenSearch - 12-18 GB
* Graylog VM has 64 GB RAM, in this case, you can allocate:
  * Graylog - 20 GB
  * OpenSearch - 24-31 GB

After it and after you will fix OOM in Graylog or OpenSearch you can try to analyze which other
performance issues you have.

[Back to TOC](#table-of-content)

## Index read-only Warnings

**Symptoms:**

* A lot of `index X is read-only` warnings in `https://<graylog_url>/system/indices/failures`
* Graylog does not store logs in the OpenSearch. In or Out messages count is non-zero,
  but recent logs cannot be found on the **Search** page.

**Root cause:**

OpenSearch has a disk-allocator feature that marks indices as `read-only` if disk space utilization is too high.
By default, this feature is turned on and indices become `read-only` if disk space usage is **95%**.

Also, there are two threshold after reach which OpenSearch start signal about problems about high disk usage:

* Low - **80%**
* High - **90%**

**How to fix:**

Fist of all, if OpenSearch indices already marked as `read-only` you need to check disk space and cleanup data.

You can use the OpenSearch API to get list of indices and to remove old indices.
For example:

* To get list of indices:

  ```bash
  curl -X GET -u <username>:<password> http://localhost:9200/_cat/indices
  ```

* To remove indices by name, by some names (use colon `,` as separator) or regex:

  ```bash
  curl -X DELETE -u <username>:<password> http://localhost:9200/<index_name_or_regex>
  ```

  example:

  ```bash
  curl -X DELETE -u <username>:<password> http://localhost:9200/graylog_30,graylog_31
  ```

If OpenSearch is unavailable or you can't use it API usually OpenSearch save all it data in the host directories:

```bash
/srv/docker/graylog/opensearch/nodes
/srv/docker/graylog/opensearch/archives
/srv/docker/graylog/opensearch/snapshots
```

So you can clean data from these directories manually.

**Warning!** After remove some data from directories manually you **have to restart** OpenSearch container.

Next, you can run the command to remove `read-only` flag from indices:

```bash
curl -X PUT -u <username>:<password> -H "Content-Type: application/json" -d '{"index.blocks.read_only_allow_delete": null}' http://localhost:9200/_settings
```

It makes the OpenSearch indices writeable, but it helps only for some time.
Because indices were locked, the Graylog indices rotation is configured incorrectly and the disk is used for **95%** again.

Next, you should re-check your Index rotation settings in Graylog and change the to avoid this situation in the future.

**How to avoid this issue:**

You need to configure the indices rotation in Graylog to avoid high disk space utilization. Best practices are as follows:

* The total rotation size of all index sets should not be more than **85%** of the total HDD size
* If you want to use time or message count based on rotation, you **must** correctly calculate the required storage,
  but please keep in mind that rotation strategies can use unpredictable storage size on disk

You can disable the OpenSearch disk allocator feature by executing the following command on the Logging VM:

**Warning!** We strongly don't recommend use it for production environments!

```bash
curl -X PUT -u <username>:<password> -H "Content-Type: application/json" -d '{"persistent": {"cluster.routing.allocation.disk.threshold_enabled": "false"}}' http://localhost:9200/_cluster/settings
```

In this case, the indices are never locked in a read-only state. If the indices rotation configuration is incorrect,
the free space can be fully occupied. Ensure that the rotation configuration is correct.

[Back to TOC](#table-of-content)

# FluentD

## FluentD worker killed and restart with SIGKILL

**Symptoms:**

* In FluentD logs can be found the similar logs

    ```bash
    2024-05-14 10:14:23 +0000 [error]: Worker 1 exited unexpectedly with signal SIGKILL
    2024-05-14 10:14:25 +0000 [info]: #1 init workers logger path=nil rotate_age=nil rotate_size=nil
    ```

* After restart FluentD use a lot of read DiskIO operations and throughput
* The `dmesg` logs from Kubernetes node contains message about OOM for `ruby` process

**Root cause:**

FluentD inside the container run more than 1 process. Usually it run 3 ruby processes:

* 1 supervisor process that manage other FluentD processes
* 2 worker processes, for `#0` and `#1` workers

Worker `#1` collect, process, and send logs. It has a buffer that in almost all FluentD versions
was hardcoded in `1Gb`.

If FluentD has a memory limit `1Gi` so it just can't fit all content of the buffer. As result, the worker `#1`
can be killed by OOMKiller and restarted by supervisor process.

**How to fix:**

There are two solutions that exist to fix this issue:

* Either need increase the FluentD memory limit till `1500Mi` or to `2Gi`
* Or need to decrease the buffer size to value less than `1Gb`, for example to `512Mb`

    ```yaml
    <store ignore_error>
      @type gelf
      @log_level warn
      host "#{ENV['GRAYLOG_HOST']}"
      port "#{ENV['GRAYLOG_PORT']}"
      ...
      retry_wait 1s
      <buffer>
        total_limit_size 512Mb
      </buffer>
    </store>
    ```

[Back to TOC](#table-of-content)

## FluentD generate a high DiskIO read load

**Symptoms:**

* FluentD use a lot of read DiskIO operations and throughput
* In FluentD logs can be found the similar logs

    ```bash
    2024-05-14 10:14:23 +0000 [error]: Worker 1 exited unexpectedly with signal SIGKILL
    2024-05-14 10:14:25 +0000 [info]: #1 init workers logger path=nil rotate_age=nil rotate_size=nil
    ```

**How to fix:**

Most probably it's a problem described in [FluentD worker killed and restart with SIGKILL](#fluentd-worker-killed-and-restart-with-sigkill).
So please refer to it to check root cause and how to fix it.

[Back to TOC](#table-of-content)

## FluentD failed to flush buffer, data too big

**Symptoms:**

* In FluentD logs can be found the similar logs

    ```bash
    2024-02-13 11:46:42 +0000 [warn]: #1 failed to flush the buffer. error_class="ArgumentError" error="Data too big (514737 bytes), would create more than 128 chunks!" plugin_id="object:cb5c"
    2024-02-13 11:46:42 +0000 [warn]: #1 got unrecoverable error in primary and no secondary error_class=ArgumentError error="Data too big (514737 bytes), would create more than 128 chunks!"
    ```

**Root cause:**

The error can be reproduced if you send logs to Graylog using input with UDP protocol.
According to [official documentation](https://go2docs.graylog.org/5-0/getting_in_log_data/gelf.html?#GELFviaUDP)
related to GELF there are restrictions:

```bash
UDP datagrams are limited to a size of 65536 bytes. Some Graylog components are limited to processing up to 8192 bytes.
```

```bash
All chunks **MUST** arrive within 5 seconds or the server will discard all chunks that have arrived or
are in the process of arriving. A message **MUST NOT** consist of more than 128 chunks.
```

There was FluentD [issue](https://github.com/fluent/fluentd/issues/3651) on GitHub also.

To send logs to Graylog we use output plugin [`fluent-plugin-gelf-hs`](https://github.com/hotschedules/fluent-plugin-gelf-hs)
that uses Ruby module [`gelf-rb`](https://github.com/graylog-labs/gelf-rb).
The gelf-hs library creates the Notifier from the `gelf` using the "WAN" network type [here](
https://github.com/hotschedules/fluent-plugin-gelf-hs/blob/master/lib/fluent/plugin/out_gelf.rb#L52).
It means that the `max chunk size` should be set as `1420 (bytes)` [here](
https://github.com/graylog-labs/gelf-rb/blob/master/lib/gelf/notifier.rb#L57-L67).

According to GELF specification and validation, we should not separate data into more than 128 chunks.
It seems the max data size for GELF UDP that can be sent is:

```bash
1420 bytes * 128 chunks = 181760 bytes =~ 177 Kb
```

So if we correctly understand the ruby code the max data size for GELF UDP is **~177 Kb**.

**How to fix:**

We highly recommend to use TCP connection from FluentD/FluentBit to Graylog.

If you need to use UDP connection for some reasons, you can try to set smaller value in Graylog buffer section
in FluentD configuration. To do that you need to:

1. Scale logging-service-operator to 0 replicas (because it can rewrite your changes in Fluentd Configuration).

    ```bash
    kubectl scale -n <namespace> deployment logging-service-operator --replicas=0
    ```

2. Edit configmap `logging-fluentd`.

    ```bash
    kubectl edit cm logging-fluentd -n <namespace>
    ```

3. Find part of configuration `output-graylog.conf`, in section `<buffer>` set the value:

    ```xml
    <store ignore_error>
      @type gelf
      # other parameters
      <buffer>
        chunk_limit_size 176KB
      </buffer>
    </store>
    ```

**Note:** Read more about buffering parameters in
[official FluentD documentation](https://docs.fluentd.org/configuration/buffer-section#buffering-parameters).

[Back to TOC](#table-of-content)

# FluentBit

## Connection timeout to Graylog in FluentBit

**Symptoms:**

The following errors in FluentBit pod logs appears:

```bash
[2024/04/25 20:54:28] [error] [upstream] connection #-1 to tcp://unavailable:0 timed out after 10 seconds (connection timeout)
[2024/04/25 20:54:29] [ warn] [net] getaddrinfo(host='<graylog_url>', err=12): Timeout while contacting DNS servers
[2024/04/25 20:54:29] [error] [output:gelf:gelf.1] no upstream connections available
```

**How to fix:**

1. Check CPU consumption by FluentBit. Usually the error appears when FluentBit
faced limit of CPU. Increase limit by setting `fluentbit.resources.limits.cpu: "1"`.

2. Add the configuration of network and health checks to FluentBit. ConfigMap `logging-fluentbit` should contain
next parameters:

    ```yaml
      fluent-bit.conf: |
        [SERVICE]
            Flush         5
            HC_Errors_Count 5
            HC_Retry_Failure_Count 5
            HC_Period 5
      output-graylog.conf: |
        [OUTPUT]
            Name     gelf
            # configuration...
            net.connect_timeout 20s
            net.max_worker_connections 35
            net.dns.mode TCP
            net.dns.resolver LEGACY
    ```

3. After updating ConfigMap you should manually delete all FluentBit pods to apply changes.

[Back to TOC](#table-of-content)

## FluentBit stuck and stopped sending logs to Graylog

**Symptoms:**

FluentBit stuck and do not send any logs.

**How to fix:**

First of all, check that you upgraded to the latest version of Logging.
If you want to solve problem manually, follow steps below (it is temporary solution):

1. Make sure that you are connected to the Cloud. Scale `logging-service-operator` deployment to 0 replicas with the
   command:

   ```bash
   kubectl scale -n logging deployment logging-service-operator --replicas=1
   ```

2. Modify ConfigMap `logging-fluentbit`:

   ```bash
   kubectl edit -n logging cm logging-fluentbit
   ```

   1. Remove from `filter-log-parser.conf` the last lines:

      ```yaml
      [FILTER]
          Name          rewrite_tag
          Match         raw.*
          Rule          $log .*  parsed.$TAG false
          Emitter_Name          raw_parsed
          Emitter_Storage.type  filesystem
          Emitter_Mem_Buf_Limit 10M
      ```

   2. Change in `output-graylog.conf` the line from

      ```yaml
      Match   parsed.**
      ```

      to

      ```yaml
      Match_Regex (raw|parsed).**
      ```

3. The last step is to delete all pods `logging-fluentbit-*`. The pods will be restarted with the last configuration.

[Back to TOC](#table-of-content)

# ConfigMap Reloader

## Fluent container restarts after changing ConfigMap

**Symptoms:**

Fluent container restarts after manually updating configmap.

**How to fix:**

1. Check logs from the main Fluents' container. It contains information about error.
Example:

   ```bash
      2024-09-12 09:56:10 +0000 [error]: Worker 0 exited unexpectedly with status 1
      /fluentd/vendor/bundle/ruby/3.2.0/gems/fluentd-1.17.1/lib/fluent/config/basic_parser.rb:92:in `parse_error!': unmatched end tag at filter-add-hostname.conf line 6,12 (Fluent::ConfigParseError)
         hostnameTest "#{ENV['HOSTNAME']}"
       </record>
      ----------^
      </filter>
   ```

2. Found the problem file: `unmatched end tag at filter-add-hostname.conf line 6,12`
3. Fix configuration and wait for the next reload.

[Back to TOC](#table-of-content)
