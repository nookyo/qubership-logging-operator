This document describes which privileges the `logging-operator` requires and how to install the `logging-operator`
with restricted privileges.

# Table of Content

* [Table of Content](#table-of-content)
* [Overview](#overview)
* [Explanation of using permissions](#explanation-of-using-permissions)
  * [Logging operator permissions](#logging-operator-permissions)
  * [FluentD permissions](#fluentd-permissions)
  * [FluentBit permissions](#fluentbit-permissions)
  * [Cloud event reader permissions](#cloud-event-reader-permissions)
  * [Graylog permissions](#graylog-permissions)
* [Deploy with restricted privileges](#deploy-with-restricted-privileges)
  * [Before you begin](#before-you-begin)
  * [All cluster-wide objects](#all-cluster-wide-objects)
  * [Custom Resource Definitions](#custom-resource-definitions)
  * [Cloud event reader cluster objects](#cloud-event-reader-cluster-objects)
  * [FluentD cluster objects](#fluentd-cluster-objects)
  * [FluentBit cluster objects](#fluentbit-cluster-objects)
    * [FluentBit DaemonSet cluster objects](#fluentbit-daemonset-cluster-objects)
    * [FluentBit StatefulSet cluster objects](#fluentbit-statefulset-cluster-objects)
  * [Graylog cluster objects](#graylog-cluster-objects)

# Overview

In general, the application requires a set of `Roles` to have access to resources inside the namespace
where it is deployed and a set of `ClusterRoles` to have limited access for cluster-scoped resources
or resources in other namespaces, e.g. to discover PODs metadata.

The logging operator can be deployed with full access to the cluster (it means that the logging operator creates
all necessary `ClusterRoles` and `CustomResourceDefinitions` by itself) and in the restricted mode
when the logging operator doesn't have permission to create or edit cluster-scoped resources
and resources in other namespaces except its custom resources.

The restricted deployment mode means that all required resources for access to cluster-scoped resources and custom
resource definitions must be created manually by a deployment engineer. The mode can be activated via the deployment
property:

```yaml
createClusterAdminEntities: [true|false]
```

**Warning!** If you want just to get an instruction on how to create all necessary resources to install
the logging operator in non-privileged mode, please, follow the
[Deploy with restricted privileges](#deploy-with-restricted-privileges) section.

In restricted privileges mode the logging operator doesn't have permission to create resources outside
the namespace where it is deployed (except its custom resources) and has limited permissions to read, list or watch
resources outside its namespace and cluster-scoped resources.

# Explanation of using permissions

This section describes per logging component why these components require cluster-wide permissions, and how to avoid
using these permissions (of course, if possible).

## Logging operator permissions

The logging operator requires `ClusterRole` and `ClusterRoleBinding` also to access Nodes information
to set container runtime type from nodes to read and process logs correctly.

Permissions in logging-operator `ClusterRole`:

```yaml
- apiGroups:
    - ""
  resources:
    - nodes
  verbs:
    - get
    - list
    - watch
```

## FluentD permissions

The FluentD requires a `ClusterRole` with permissions to get/list/watch permissions to collect metadata of PODs
not enrich logs with metadata about their origin:

```yaml
- apiGroups:
    - ""
  resources:
    - pods
    - namespaces
    - events
    - endpoints
  verbs:
    - get
    - list
    - watch
```

Default `ClusterRole` with the name `view` can be used to cover these permissions.

Also, FluentD uses `hostPath` and requires access to host directories on Kubernetes Nodes and runs under `root (0)` user.
Depending on using Cloud type and version FluentD can require different configurations to use `hostPath`:

* Kubernetes >= 1.25 - FluentD can use `privileged` PodSecurityStandards (PSS), no need to create any additional objects
* Kubernetes < 1.25 - Need to use `PodSecurityPolicy` with name `logging-fluentd`
* OpenShift 4.x - Need to use `SecurityContextConstraints` with name `logging-fluentd`

**Note:** It's not possible to run FluentD without access to `hostPath`! In the current deployment schema,
it's not possible to avoid using `privileged` PSS or specific PSP or SCC.

## FluentBit permissions

The FluentBit requires a `ClusterRole` with permissions to get/list/watch permissions to collect metadata of PODs
not enrich logs with metadata about their origin:

```yaml
- apiGroups:
    - ""
  resources:
    - pods
    - namespaces
    - events
    - endpoints
  verbs:
    - get
    - list
    - watch
```

Default `ClusterRole` with the name `view` can be used to cover these permissions.

Also, FluentBit uses `hostPath` and requires access to host directories on Kubernetes Nodes and runs under `root (0)` user.
Depending on using Cloud type and version FluentBit can require different configurations to use `hostPath`:

* Kubernetes >= 1.25 - FluentBit can use `privileged` PodSecurityStandards (PSS), no need to create any additional objects
* Kubernetes < 1.25 - Need to use `PodSecurityPolicy` with name `logging-fluentbit`
* OpenShift 4.x - Need to use `SecurityContextConstraints` with name `logging-fluentbit`

**Note:** It's not possible to run FluentBit without access to `hostPath`! In the current deployment schema,
it's not possible to avoid using `privileged` PSS or specific PSP or SCC.

## Cloud event reader permissions

The cloud event reader requires a `ClusterRole` with permissions to get/list/watch permissions to collect
Kubernetes events:

```yaml
- apiGroups:
    - ""
  resources:
    - pods
    - events
  verbs:
    - get
    - list
    - watch
```

Default `ClusterRole` with the name `view` can be used to cover these permissions.

## Graylog permissions

The Graylog requires a `ClusterRole` with permissions to PSP or SCC:

* Kubernetes < 1.25 - Need to use `PodSecurityPolicy` with name `logging-graylog`
* OpenShift 4.x - Need to use `SecurityContextConstraints` with name `logging-graylog`

# Deploy with restricted privileges

The logging operator can be deployed with restricted permissions.

It means during deploy won't try to create cluster-wide objects and won't require permissions for cluster-wide objects
like `ClusterRole`, `ClusterRoleBinding`, `CustomResourceDefinitions` and so on.

To do a successful installation the set of cluster-scoped resources should be created manually
**before** deploy under a user with enough permissions to do it.

## Before you begin

* Installed and configured `kubectl`
* ServiceAccount with permissions to manage `ClusterRole`, `ClusterRoleBinding`, `CustomResourceDefinitions` in the
  Cloud to manual creation cluster-wide objects
* ServiceAccount with full access to the namespace in which `logging-operator` will be deployed
  including access to API groups `qubership.org` that will use Helm

Example of ServiceAccount with full access to namespace for Helm:

```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: logging-service
  name: logging-full-access
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
```

## All cluster-wide objects

The full step-by-step guide on how to create necessary resources is described below.

Steps to execute:

* Download and extract roles. If you see this document in GitLab you can use the button:
  `Download -> Download this directory -> Zip`
* Modify template files in folder `manifests`, and change the namespace from default to your one.
* Open a terminal in the folder with extracted files
* Run command to create `CustomResourceDefinitions`, `ClusterRoles` and `ClusterRoleBindings`:

    ```bash
    kubectl apply -f ./manifests --recursive=true
    ```

* Set parameter for deploy to skip CRD-s creation `--skip-crds`
* Set parameter for skip creation of cluster-wide RBAC entities
  
    ```yaml
    createClusterAdminEntities: false
    ```

## Custom Resource Definitions

The `logging-operator` and its downstream component require a `CustomResourceDefinition` resource to control
an application.

If you want to deploy into **Kubernetes v1.16+** in the restricted rights mode
the following `CRD` version v1 resources should be created manually before deploy:

* [loggingservices.logging.qubership.org](/docs/crds/logging.qubership.org_loggingservices.yaml)

To create the specified resources you can use the command (from a terminal opened in the root `logging-operator` folder):

```bash
kubectl create -f docs/crds/
```

## Cloud event reader cluster objects

The resources with cluster scope access should be created when deployed with restricted privileges:

* `ClusterRoleBinding` with name [\<NAMESPACE>-cloud-events-reader-cluster-reader](/docs/rbac/manifests/cloud-event-reader/clusterrole-binding.yaml)

To create the specified resources you can use the command (from a terminal opened in the document folder):

```bash
kubectl apply -f manifests/cloud-event-reader/ --recursive=true
```

## FluentD cluster objects

The resources with cluster scope access should be created when deployed with restricted privileges:

* Kubernetes:
  * `ClusterRole` with name [\<NAMESPACE>-logging-fluentd-cluster-role](/docs/rbac/manifests/fluentd/kubernetes/clusterrole.yaml)
  * `ClusterRoleBinding` with name [\<NAMESPACE>-fluentd-cluster-reader](/docs/rbac/manifests/fluentd/kubernetes/clusterrole-binding.yaml)
  * `PodSecurityPolicy` with name [\<NAMESPACE>-logging-fluentd](/docs/rbac/manifests/fluentd/kubernetes/podsecuritypolicy.yaml)
* OpenShift:
  * `ClusterRoleBinding` with name [\<NAMESPACE>-fluentd-cluster-reader](/docs/rbac/manifests/fluentd/openshift/clusterrole-binding.yaml)
  * `SecurityContextConstraints` with name [\<NAMESPACE>-logging-fluentd](/docs/rbac/manifests/fluentd/openshift/securitycontextconstraints.yaml)

To create the specified resources you can use the command (from a terminal opened in the document folder):

```bash
kubectl apply -f manifests/fluentd/<cloud_type> --recursive=true
```

where `<cloud_type>`:

* `kubernetes` in case of deployment in Kubernetes
* `openshift` in case of deployment in OpenShift

## FluentBit cluster objects

### FluentBit DaemonSet cluster objects

The resources with cluster scope access should be created when deployed with restricted privileges:

* Kubernetes:
  * `ClusterRole` with name [\<NAMESPACE>-logging-fluentbit-cluster-role](/docs/rbac/manifests/fluentbit-daemonset/kubernetes/clusterrole.yaml)
  * `ClusterRoleBinding` with name [\<NAMESPACE>-fluentbit-cluster-reader](/docs/rbac/manifests/fluentbit-daemonset/kubernetes/clusterrole-binding.yaml)
  * `PodSecurityPolicy` with name [\<NAMESPACE>-logging-fluentbit](/docs/rbac/manifests/fluentbit-daemonset/kubernetes/podsecuritypolicy.yaml)
* OpenShift:
  * `ClusterRoleBinding` with name [\<NAMESPACE>-fluentbit-cluster-reader](/docs/rbac/manifests/fluentbit-daemonset/openshift/clusterrole-binding.yaml)
  * `SecurityContextConstraints` with name [\<NAMESPACE>-logging-fluentbit](/docs/rbac/manifests/fluentbit-daemonset/openshift/securitycontextconstraints.yaml)

To create the specified resources you can use the command (from a terminal opened in the document folder):

```bash
kubectl apply -f manifests/fluentbit-daemonset/kubernetes --recursive=true
```

### FluentBit StatefulSet cluster objects

The resources with cluster scope access should be created when deployed with restricted privileges:

* Kubernetes:
  * `ClusterRole` with name [\<NAMESPACE>-logging-fluentbit-aggregator-cluster-role](/docs/rbac/manifests/fluentbit-statefulset/kubernetes/clusterrole.yaml)
  * `ClusterRoleBinding` with name [\<NAMESPACE>-logging-fluentbit-aggregator-cluster-reader](/docs/rbac/manifests/fluentbit-statefulset/kubernetes/clusterrole-binding.yaml)
  * `PodSecurityPolicy` with name [\<NAMESPACE>-logging-fluentbit-aggregator](/docs/rbac/manifests/fluentbit-statefulset/kubernetes/podsecuritypolicy.yaml)
* OpenShift:
  * `ClusterRoleBinding` with name [\<NAMESPACE>-logging-fluentbit-aggregator-cluster-reader](/docs/rbac/manifests/fluentbit-statefulset/openshift/clusterrole-binding.yaml)
  * `SecurityContextConstraints` with name [\<NAMESPACE>-logging-fluentbit-aggregator](/docs/rbac/manifests/fluentbit-statefulset/openshift/securitycontextconstraints.yaml)

To create the specified resources you can use the command (from a terminal opened in the document folder):

```bash
kubectl apply -f manifests/fluentbit-statefulset/kubernetes --recursive=true
```

## Graylog cluster objects

The resources with cluster scope access should be created when deployed with restricted privileges:

* Kubernetes:
  * `ClusterRole` with name [\<NAMESPACE>-logging-graylog-cluster-role](/docs/rbac/manifests/graylog/kubernetes/clusterrole.yaml)
  * `ClusterRoleBinding` with name [\<NAMESPACE>-graylog-cluster-role](/docs/rbac/manifests/graylog/kubernetes/clusterrole-binding.yaml)
  * `PodSecurityPolicy` with name [\<NAMESPACE>-logging-graylog](/docs/rbac/manifests/graylog/kubernetes/podsecuritypolicy.yaml)
* OpenShift:
  * `SecurityContextConstraints` with name [\<NAMESPACE>-logging-graylog](/docs/rbac/manifests/graylog/openshift/securitycontextconstraints.yaml)

To create the specified resources you can use the command (from a terminal opened in the document folder):

```bash
kubectl apply -f manifests/graylog/kubernetes --recursive=true
```
