The section provides information about Cloud Events Reader specific parameters. The parameters are set to it as
command line arguments. If you want to set it, you need to set it under section:

```yaml
cloudEventsReader:
  args:
    ...
```

Possible command line arguments described below:

<!-- markdownlint-disable line-length -->
| Argument    | Default value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | Description                                                                                                                                  |
|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `namespace` | `-`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Namespace to watch for events. The parameter can be used multiple times.<br>If parameter is not set events of all namespaces will be watched |
| `format`    | <details><summary>value</summary>{\"time\":\"{{.LastTimestamp.Format \"2006-01-02T15:04:05.999\"}}\",\"involvedObjectKind\":\"{{.InvolvedObject.Kind}}\",\"involvedObjectNamespace\":\"{{.InvolvedObject.Namespace}}\",\"involvedObjectName\":\"{{.InvolvedObject.Name}}\",\"involvedObjectUid\":\"{{.InvolvedObject.UID}}\",\"involvedObjectApiVersion\":\"{{.InvolvedObject.APIVersion}}\",\"involvedObjectResourceVersion\":\"{{.InvolvedObject.ResourceVersion}}\",\"reason\":\"{{.Reason}}\",\"type\":\"{{.Type}}\",\"message\":\"{{js .Message}}\",\"kind\":\"Event\"}</details> | Format to print Event. It should be valid Golang template of `text/template` package                                                         |
| `workers`   | `2`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Workers number for controller                                                                                                                |
<!-- markdownlint-enable line-length -->

Examples:

**Note:** It's just an example of a parameter's format, not a recommended parameter.

```yaml
cloudEventsReader:
  install: true
  args:
    - "-workers=2"
    - "-namespace=logging"
    - "-namespace=monitoring"
    - "-format={\"time\":\"{{.LastTimestamp.Format \"2006-01-02T15:04:05Z\"}}\""
```

or you can set `args` without quoting:

```yaml
cloudEventsReader:
  install: true
  args:
    - -namespace=logging
    - -namespace=monitoring
    - -format={"time":"{{.LastTimestamp.Format
      "2006-01-02T15:04:05.999"}}","involvedObjectKind":"{{.InvolvedObject.Kind}}","involvedObjectNamespace":"{{.InvolvedObject.Namespace}}","involvedObjectName":"{{.InvolvedObject.Name}}","involvedObjectUid":"{{.InvolvedObject.UID}}","involvedObjectApiVersion":"{{.InvolvedObject.APIVersion}}","involvedObjectResourceVersion":"{{.InvolvedObject.ResourceVersion}}","reason":"{{.Reason}}","type":"{{.Type}}","message":"{{js
      .Message}}","kind":"EventTEST"}
```

**Note**: If you want to set non-json format, do not forget to turn off filter in FluentBit/FluentD like this:

  ```yaml
    fluentd:
      cloudEventsReaderFormat: text
  ```

This is an example of Event (API version events.k8s.io/v1).

```json
{
   "message":"Created new replication controller \"postgres-backup-daemon-1\" for version 1",
   "kind":"KubernetesEvent",
   "log":{
      "firstTimestamp":"2018-11-02T02:26:00Z",
      "reason":"DeploymentCreated",
      "metadata":{
         "name":"postgres-backup-daemon.15632d8844761c5d",
         "namespace":"pg96-nighttest",
         "creationTimestamp":"2018-11-02T02:26:00Z",
         "uid":"a38df948-de46-11e8-9fdb-fa163e5f2c4f",
         "selfLink":"/api/v1/namespaces/pg96-nighttest/events/postgres-backup-daemon.15632d8844761c5d",
         "resourceVersion":"71608432"
      },
      "apiVersion":"v1",
      "involvedObject":{
         "namespace":"pg96-nighttest",
         "name":"postgres-backup-daemon",
         "uid":"a388d4aa-de46-11e8-9fdb-fa163e5f2c4f",
         "apiVersion":"v1",
         "kind":"DeploymentConfig",
         "resourceVersion":"71608172"
      },
      "lastTimestamp":"2018-11-02T02:26:00Z",
      "count":1,
      "source":{
         "component":"deploymentconfig-controller"
      },
      "type":"Normal"
   }
}
```

Formatted log output with default template of the example Event will be printed in stdout:

<!-- markdownlint-disable line-length -->
```text
{"time":"2018-11-02T02:26:00.999"","involvedObjectKind":"DeploymentConfig","involvedObjectNamespace":"pg96-nighttest","involvedObjectName":"postgres-backup-daemon","involvedObjectUid":"a388d4aa-de46-11e8-9fdb-fa163e5f2c4f","involvedObjectApiVersion":"v1","involvedObjectResourceVersion":"71608172","reason":"DeploymentCreated","type":"Normal","message":"Created new replication controller \"postgres-backup-daemon-1\" for version 1","kind":"KubernetesEvent"}
```
<!-- markdownlint-enable line-length -->
