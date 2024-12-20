This document describes different limits, restrictions, recommendations and best practices for logs.

# Table of Content

* [Table of Content](#table-of-content)
* [Limits](#limits)
  * [Limits for Graylog](#limits-for-graylog)
  * [Limits for OpenSearch/ElasticSearch instance](#limits-for-opensearchelasticsearch-instance)
  * [Limits for OpenSearch/ElasticSearch index](#limits-for-opensearchelasticsearch-index)
  * [Limits for OpenSearch/ElasticSearch fields](#limits-for-opensearchelasticsearch-fields)
  * [Limits for CRI](#limits-for-cri)
* [Restrictions](#restrictions)
  * [Supported log levels](#supported-log-levels)
  * [Default log level](#default-log-level)
  * [Content of fields/labels for text log format](#content-of-fieldslabels-for-text-log-format)

# Limits

This section describes limits related to using a centralized log server.

## Limits for Graylog

Graylog GELF UDP has the following limits:

<!-- markdownlint-disable line-length -->
| Type                      | Limits                       |
| ------------------------- | ---------------------------- |
| Character encoding system | `UTF-8`                      |
| Max GELF UDP chunks       | `128 chunks`                 |
| Max GELF UDP package size | `1420 bytes`                 |
| Max GELF UDP total size   | `181760 bytes` (~ `177 KiB`) |
<!-- markdownlint-enable line-length -->

## Limits for OpenSearch/ElasticSearch instance

<!-- markdownlint-disable line-length -->
| Type                          | Limits        |
| ----------------------------- | ------------- |
| Max shards count per IndexSet | `1000 shards` |
<!-- markdownlint-enable line-length -->

## Limits for OpenSearch/ElasticSearch index

<!-- markdownlint-disable line-length -->
| Type                                   | Limits                                  |
| -------------------------------------- | --------------------------------------- |
| Max document size                      | `2 GiB` (better keep less than 100 MiB) |
| Max fields in the Stream               | `1000 fields`                           |
| Max words/terms in the `message` field | `65536 words/terms`                     |
| Max length of the dynamical fields     | `256 symbols`                           |
| Max word/term size                     | `32 KiB` (8192 symbols for UTF-8)       |
<!-- markdownlint-enable line-length -->

## Limits for OpenSearch/ElasticSearch fields

<!-- markdownlint-disable line-length -->
| Field type   | Limits                                                                                                                     |
| ------------ | -------------------------------------------------------------------------------------------------------------------------- |
| `INT64`      | `-2^63~2^63-1`                                                                                                             |
| `FLOAT`      | `+/-3.40282e+038`                                                                                                          |
| `DOUBLE`     | `+/-1.79769e+308`                                                                                                          |
| `LITERAL`    | A field of this type can be up to `65535 bytes` in length                                                                  |
| `TEXT`       | A field of this type can be up to `65536 words` in length                                                                  |
| `SHORT_TEXT` | A field of this type can be up to `100 bytes` in length. If a field exceeds `100 bytes` in length, the field is truncated. |
<!-- markdownlint-enable line-length -->

## Limits for CRI

<!-- markdownlint-disable line-length -->
| Type                          | Limits    |
| ----------------------------- | --------- |
| Container max log line length | `16 KiB`  |
| Container max log file size   | `5 files` |
| Container max log files       | `10 MiB`  |
<!-- markdownlint-enable line-length -->

# Restrictions

This section describes different restrictions related to log parsing and using a centralized log server.

## Supported log levels

FluentBit and FluentD by itself can parse any log and log levels in the logs. However, we are using Graylog
as a centralized logging server. Output plugins used to send logs in Graylog restrict the list of supported
log levels.

FluentD supports the following log levels (case insensitive):

<!-- markdownlint-disable line-length -->
| Log levels in logs     | Applicable levels     | Log level in Graylog |
| ---------------------- | --------------------- | -------------------- |
| `0` or start from `e`  | `emerg`, `emergency`  | `Unknown`            |
| `1` or start from `a`  | `alert`               | `Unknown`            |
| `2` or start from`c`   | `crit`, `critical`    | `Fatal`              |
| `3` or start from `er` | `error`               | `Error`              |
| `4` or start from `w`  | `warn`, `warning`     | `Warning`            |
| `5` or start from `n`  | `notice`              | `Informational`      |
| `6` or start from `i`  | `info`, `information` | `Informational`      |
| `7` or start from `d`  | `debug`               | `Debug`              |
<!-- markdownlint-enable line-length -->

The source code using the Gelf output plugin:
[https://github.com/hotschedules/fluent-plugin-gelf-hs/blob/master/lib/fluent/gelf_util.rb#L23-L51](https://github.com/hotschedules/fluent-plugin-gelf-hs/blob/master/lib/fluent/gelf_util.rb#L23-L51).

FluentBit supports the following log levels (case insensitive):

<!-- markdownlint-disable line-length -->
| Log levels in logs | Log level in Graylog |
| ------------------ | -------------------- |
| `0` or `emerg`     | `Emergency`          |
| `1` or `alert`     | `Alert`              |
| `2` or `crit`      | `Critical`           |
| `3` or `err`       | `Error`              |
| `4` or `warning`   | `Warning`            |
| `5` or `notice`    | `Notice`             |
| `6` or `info`      | `Informational`      |
| `7` or `debug`     | `Debug`              |
<!-- markdownlint-enable line-length -->

[https://github.com/fluent/fluent-bit/blob/master/src/flb_pack_gelf.c#L563-L592](https://github.com/fluent/fluent-bit/blob/master/src/flb_pack_gelf.c#L563-L592)

## Default log level

In case, if Graylog doesn't receive the log level in the message it will set the `Informational` as a default.

## Content of fields/labels for text log format

The field/label in logs can contain almost all symbols. But there are two restrictions:

* The max length of dynamic fields can be only `256` symbols (Graylog and OpenSearch limits)
* In the case of using text log format doesn't allow to use `[` and `]` symbols inside the `[..text...]` (Fluent restriction)

Example of correct logs:

```bash
[2024-03-24T16:11:35.120][ERROR][thread=pool-2-thread-5] <message>
```

Here `thread` field/label has a finite and countable number of values and has no specific symbols.

Example of **incorrect** logs:

```bash
[2024-03-24T16:11:35.120][ERROR][thread=[CPS] worker thread] <message>
```

Here the `thread` field/label has the incorrect value that **won't parse** and the whole log line **will lost**.

Example of **incorrect** logs:

```bash
[2024-03-24T16:11:35.120][ERROR][finishedAt=2022-08-31T13:46:30.845+03:00[Europe/Moscow]] <message>
```

Here the `finishedAt` field/label has the incorrect value that **won't parse** and the whole log line **will lost**.
It doesn't matter where used `[` or `]` inside the `[]`, it is still prohibited.

Example of **incorrect** logs:

```bash
[2024-03-24T16:11:35.120][ERROR][payload={"customerId":"123","paymentParts":[{"amount":{"currency":{"name":"Emirati Dirham","id":"123","currencyCode":"AED"},"value":957.60},"name":"Upfront Payment Part #2103","orderItemIds":["123","234","345","456","567"]}]] <message>
```

Here `payload` field/label has two problems:

* use restricted symbols `[` and `]` inside the `[]` because contains a JSON
* potentially this JSON payload can have a length of more than `256` symbols

This log line **will be lost**.

Example of **incorrect** logs:

```bash
[2024-03-24T16:11:35.120][ERROR][XNIO-1 task-1] <message>
```

Here the third field/label `[XNIO-1 task-1]` is not a `key=value` pair. It won't affect parsing but this part of
the log line **won't parse as a field/label** and can be used only in full-text search.
