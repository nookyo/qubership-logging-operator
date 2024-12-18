This document describes how to manually create, update or delete Logging Custom Resource Definitions (CRDs).

# Table Of Contents

* [Table Of Contents](#table-of-contents)
* [When is it needed?](#when-is-it-needed)
* [Before you begin](#before-you-begin)
* [How to manage CRDs](#how-to-manage-crds)
  * [Create](#create)
  * [Upgrade](#upgrade)
  * [Remove](#remove)

# When is it needed?

Almost all Qubership applications and microservices are integrated with Logging. This integration means
that almost all microservices during deploy can create `LoggingService` objects from the API
`logging.qubership.org`.

Such objects allow for Logging to understand from which microservices and how to collect metrics.
Similar objects allow providing alerts, recording rules and Grafana dashboards.

Also, it means that before deploying applications and microservices in Kubernetes all Custom Resource Definitions
(CRDs) must already created in Kubernetes. Otherwise, the deployment will fail.

The deployment process (helm chart) of Logging will automatically install or update all required CRDs.
But there are environments where Logging can't or doesn't make sense to deploy. In such cases, there is
an ability to create or update CRDs manually.

In the list of Logging artifacts, you can find an archive with the name

```bash
logging-operator-<version>-crds.zip
```

that contains all CRDs related to AMS operator that are required for Logging.
How to use this archive and CRDs inside you can read below.

<!-- #GFCFilterMarkerStart# -->
[Back to TOC](#table-of-contents)
<!-- #GFCFilterMarkerEnd# -->

# Before you begin

* You should have cluster-wide permissions enough to operate with CRDs (cluster admin is not required)
* You should configure a context for your `kubectl` and make sure the connection configured to correct Kubernetes

# How to manage CRDs

This section describes different cases of manual manipulation with CRDs.

## Create

To create CRDs for Logging you need to execute the command:

```bash
kubectl create -f path/to/crds/directory/
```

**Warning!** Never use the `kubectl apply` command! This command will generate the annotation
`kubectl.kubernetes.io/last-applied-configuration` that will contain the whole content of CRD.
It means that the object size will be increasing two times and it may lead to problems with storing
this CRD in Etcd.

```bash
mkdir /tmp/crds/
unzip -d /tmp/crds/ logging-operator-<version>-crds.zip
kubectl create -f /tmp/crds/*
```

<!-- #GFCFilterMarkerStart# -->
[Back to TOC](#table-of-contents)
<!-- #GFCFilterMarkerEnd# -->

## Upgrade

To upgrade CRDs for Logging you need to execute the command:

```bash
kubectl replace -f path/to/crds/directory/
```

**Warning!** Never use the `kubectl apply` command! This command will generate the annotation
`kubectl.kubernetes.io/last-applied-configuration` that will contain the whole content of CRD.
It means that the object size will be increasing two times and it may lead to problems with storing
this CRD in Etcd.

For example, if you will use the archive with CRDs providing Logging:

```bash
mkdir /tmp/crds/
unzip -d /tmp/crds/ logging-operator-<version>-crds.zip
kubectl replace -f /tmp/crds/*
```

<!-- #GFCFilterMarkerStart# -->
[Back to TOC](#table-of-contents)
<!-- #GFCFilterMarkerEnd# -->

## Remove

**Warning!** This step removes **all CRDs** for Logging and deleting these CRDs causes the deletion of
**all resources** of their type in **all namespaces**.
It means that all resources like `LoggingService` from applications are removed.

To remove CRDs and all Custom Resources (CRs) `LoggingService` from all namespaces for Logging
you need to execute the command:

```bash
# ams-operator CRD
kubectl delete crd loggingservices.logging.qubership.org
```

<!-- #GFCFilterMarkerStart# -->
[Back to TOC](#table-of-contents)
<!-- #GFCFilterMarkerEnd# -->
