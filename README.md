# qubership-logging-operator

The Logging Operator deploys in Kubernetes next components:

* Graylog
* FluentD
* FluentBit
* K8S Events Reader

## Concept

The main concept of operators is to extend Kubernetes API by creating custom resources
and controllers that watch this resource.

For more information, refer to
[https://kubernetes.io/docs/concepts/extend-kubernetes/operator/](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

This Operator is created using Operator SDK. For more information, refer to
[https://github.com/operator-framework/operator-sdk](https://github.com/operator-framework/operator-sdk).

## Features

The main features are as follows:

* Logging deployment via the template
* Kubernetes and OpenShift support
* FluentD deployment
* FluentBit deployment
* K8S Events Reader deployment
* Run Integration Tests

## Documents

All these documents provided to customer:

* [Architecture](/docs/architecture.md)
* [Installation](/docs/installation.md)
* [Maintenance](/docs/maintenance.md)
* [Troubleshooting](/docs/troubleshooting.md)
