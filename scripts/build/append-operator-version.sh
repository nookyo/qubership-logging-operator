#!/usr/bin/env bash

OPERATOR_CRD_DIR="charts/qubership-logging-operator/crds"
GROUP_NAME="logging.qubership.org"
CRD_GROUP="loggingservices.${GROUP_NAME}"
OPERATOR_ANNOTATION="logging-operator.${GROUP_NAME}/version"

COMMON_LABELS="  labels:\n    app.kubernetes.io/component: qubership-logging-operator\n    app.kubernetes.io/part-of: logging"

# Add annotation with the version
if [[ "$OSTYPE" == "darwin"* ]]; then
  find "${OPERATOR_CRD_DIR}" -name '*.yaml' -exec sed -i '' -e "/^    controller-gen.kubebuilder.io.version.*/a\\    ${OPERATOR_ANNOTATION}: ${VERSION}" {} +
else
  find "${OPERATOR_CRD_DIR}" -name '*.yaml' -exec sed -i "/^    controller-gen.kubebuilder.io.version.*/a\\    ${OPERATOR_ANNOTATION}: ${VERSION}" {} +
fi

# Add default labels
if [[ "$OSTYPE" == "darwin"* ]]; then
  find "${OPERATOR_CRD_DIR}" -name '*.yaml' -exec sed -i '' -e "/^  name: ${CRD_GROUP}/i\\${COMMON_LABELS}" {} +
else
  find "${OPERATOR_CRD_DIR}" -name '*.yaml' -exec sed -i "/^  name: ${CRD_GROUP}/i\\${COMMON_LABELS}" {} +
fi
