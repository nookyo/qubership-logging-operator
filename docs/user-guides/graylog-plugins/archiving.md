This document provide information about Graylog Archiving plugin.

* [Overview](#overview)
* [Custom Archiving Plugin](#custom-archiving-plugin)
  * [Archiving process](#archiving-process)
  * [Restoring archived data](#restoring-archived-data)
  * [Configuration](#configuration)

# Overview

In some cases, you can start thinking about archiving logs from Graylog to any cold storage.

There are a lot of reasons to start thinking about it, like reducing the costs, store logs
on the cold storage for a long time (months, years) and so on.

We know two ways how you can configure archiving in Graylog:

* Pay for the Enterprise version where this feature exists
* Use our custom archiving plugin for Graylog

# Custom Archiving Plugin

Graylog uses OpenSearch (ElasticSearch in old versions) to store data after processing.
This plugin is installed in Graylog and uses the OpenSearch API to archive or restore data.

## Archiving process

How the archiving process works:

1. You configured the archiving (read below about configuration)
2. By scheduler or by request run the archiving process in the plugin
3. It connects to OpenSearch and runs snapshot creation

How it should work?

The plugin regularly should create snapshots (using OpenSearch) for data received, processed, and saved by Graylog.
These data will be saved in snapshots and will not as a part of Graylog's IndexSets.
So Graylog can rotate and remove logs from its IndexSet.

We can visualize it as:

```bash
----------------------------------------------> Time
    -----------> Data in OpenSearch Index (Hot/Warn data)
               ^
               |
               Remove data
         -------------------------------------> Data in OpenSearch snapshot (Cold data)
         ^     ^
         |     |
         The archiving plugin starts creating snapshots (it can include old data from already existing indices)
               |
               Data just continues to keep in snapshots
```

There are two options for how OpenSearch can save snapshots:

* on a local filesystem
* in S3-compatible storage (AWS S3, MinIO, and so on), but for OpenSearch 1.x needs to install the additional plugin
  (since OpenSearch 2.x it is included in OpenSearch vanilla image)

## Restoring archived data

This plugin can restore early archived data. It can restore previously created snapshots from local storage (verified)
and potentially from S3-compatible storage (unverified).

But there are some restrictions:

* Data/logs from snapshots aren't available in Graylog, you can't select these data without restore
* Restore is a manual process, there is no button in the Graylog Web interface to run restore
* Restore can be run only using API
* Because Graylog has a rotation strategy for Indices and can be cases when Graylog can use the same Index name
  (for example, if you are using snapshots to transfer data between Graylogs) all snapshots restore in Indices
  with the prefix `restored_`. For example, was `graylog_10` after restoring this Index will have
  the name `restored_graylog_10`.

Why the last restriction is important?

If we try to restore data in the Index with the original name it can:

* Match in already existing IndexSet and Graylog can try to remove it according to the rotation strategy
* Its original name may conflict with the already existing Index name

Also, it means that the restored Index won't show in the Graylog Web interface. It won't show because there is no
IndexSet and associated Stream for the restored Index.
Currently, the archiving plugin has no logic to add IndexSet and Stream for restored Indices in Graylog.
So after restoring you should manually add these objects.

## Configuration

The archiving plugin has no UI and can't be configured from the Graylog Web interface.
It has the configuration file, but better to use API to configure it.
