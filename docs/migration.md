Fluentd -> Fluent bit migration guide

# Table of Content

* [Table of Content](#table-of-content)
* [Migration from FluentD to FluentBit](#migration-from-fluentd-to-fluentbit)
  * [Install case](#install-case)
  * [Upgrade case](#upgrade-case)
    * [HWE for FluentBit](#hwe-for-fluentbit)
    * [Upgrade procedure](#upgrade-procedure)
  * [High-availability FluentBit](#high-availability-fluentbit)
    * [Upgrade procedure for HA FluentBit](#upgrade-procedure-for-ha-fluentbit)
  * [HWE for HA FluentBit](#hwe-for-ha-fluentbit)

# Migration from FluentD to FluentBit

## Install case

To install Fluent-bit use parameter `fluentbit.install: true`.
All parameters for Fluent-bit and Fluentd are similar.
Logically configurations for Fluent-bit and Fluentd match.

The same fields/labels:

| Fluentd   | FluentBit |
| --------- | --------- |
| namespace | namespace |
| pod       | pod       |
| container | container |

There are several differences in fields-naming:

| Fluentd            | FluentBit                                                       |
| ------------------ | --------------------------------------------------------------- |
| container_id       | kubernetes_docker_id                                            |
| container_image    | kubernetes_container_image                                      |
| container_image_id | kubernetes_container_hash                                       |
| container_name     | kubernetes_container_name                                       |
| docker             | `-`                                                             |
| facility           | transmitter                                                     |
| host               | kubernetes_host                                                 |
| kubernetes         | `-`                                                             |
| labels-component   | `-`                                                             |
| labels-tier        | `-`                                                             |
| master_url         | `-`                                                             |
| namespace_id       | `-`                                                             |
| pod_id             | kubernetes_pod_id                                               |
| pod_ip             | source                                                          |
| protocol           | `-`                                                             |
| source             | hostname                                                        |
| time               | `-`                                                             |
| `-`                | application_id                                                  |
| `-`                | kubernetes_annotations_cni_projectcalico_org_containerID        |
| `-`                | kubernetes_annotations_cni_projectcalico_org_podIP              |
| `-`                | kubernetes_annotations_cni_projectcalico_org_podIPs             |
| `-`                | kubernetes_annotations_kubernetes_io_psp                        |
| `-`                | kubernetes_annotations_seccomp_security_alpha_kubernetes_io_pod |
| `-`                | kubernetes_docker_id                                            |
| `-`                | kubernetes_labels_app_kubernetes_io_component                   |
| `-`                | kubernetes_labels_app_kubernetes_io_instance                    |
| `-`                | kubernetes_labels_app_kubernetes_io_name                        |
| `-`                | kubernetes_labels_controller-revision-hash                      |
| `-`                | kubernetes_labels_pod-template-generation                       |

## Upgrade case

### HWE for FluentBit

FluentBit starts with the following resources:

```yaml
resources:
  requests:
    cpu: 50m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 512Mi
```

### Upgrade procedure

To upgrade from Fluentd to FluentBit use the parameters:

```yaml
fluentd:
  install: false
fluentbit:
  install: true
```

In this case `logging-operator` uninstalls all FluentD entities and installs FluentBit.

## High-availability FluentBit

There are some cases when you have high load of logs or you need to save all logs even if Graylog is not available for
a long time.

### Upgrade procedure for HA FluentBit

You can upgrade HA Fluentbit from Fluentd directly or from FluentBit. The procedures are the same.
You need to set the following parameters:

```yaml
fluentd:
  intall: false
fluentbit:
  intall: true
  aggregator:
    install: true
```

The other required or necessary parameters are listed in the
[Installation Guide](/docs/installation.md).

## HWE for HA FluentBit

The schema requires more resources for proper work with a high amount of data.
The default resources requirement is:

```yaml
fluentbit:
  resources:
    requests:
      cpu: 50m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 512Mi
  aggregator:
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 2Gi
```
