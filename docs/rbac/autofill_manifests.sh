#!/bin/bash

for arg in "$@"
do
  case $arg in
    -n=*|--namespace=*)
    NAMESPACE="${arg#*=}"
    ;;
    -d=*|--dir=*)
    PATH_TO_REPO="${arg#*=}"
    ;;
    -c=*|--cloud=*)
    CLOUD="${arg#*=}"
    ;;
  esac
done

if [ "$CLOUD" != "" ] && [ "$NAMESPACE" != "" ] && [ "$PATH_TO_REPO" != "" ]
  then
    if ls ${PATH_TO_REPO}/documentation/rbac/manifests > /dev/null
      then
        echo "Copy manifests to /tmp/${NAMESPACE}-manifest"
        cp -r ${PATH_TO_REPO}/documentation/rbac/manifests /tmp/${NAMESPACE}-manifest

        if [ $CLOUD == "openshift" ]
          then
            echo "Find and replace default namespace: logging-service to ${NAMESPACE} "
            find /tmp/${NAMESPACE}-manifest -type f -exec sed -i -e "s/namespace: logging-service/namespace: ${NAMESPACE}/g" {} \;
            echo "Find and replace all default names logging-service-<something> to ${NAMESPACE}-<something>"
            find /tmp/${NAMESPACE}-manifest -type f -exec sed -i -e "s/name: logging-service/name: ${NAMESPACE}/g" {} \;
        elif [ $CLOUD == "k8s" ] || [ $CLOUD == "kubernetes" ]
          then
            echo "Find and replace default namespace: logging-service to ${NAMESPACE} "
            find /tmp/${NAMESPACE}-manifest -type f -exec sed -i -e "s/namespace: logging-service/namespace: ${NAMESPACE}/g" {} \;
            echo "Find and replace all default names logging-service-<something> to ${NAMESPACE}-<something>"
            find /tmp/${NAMESPACE}-manifest -type f -exec sed -i -e "s/name: logging-service/name: ${NAMESPACE}/g" {} \;
        else
          echo "error, arguments are incorrectly filled"
        fi
    else
      echo "No such ${PATH_TO_REPO} directory"
    fi
elif [ "$arg" == "--help" ] || [ "$arg" == "-h" ]
    then
      echo "  Script to automate filling in manifests when deploying with limited rights
      -n=, --namespace= - indicate namespace where you to deploy logging-operator
      -d=, --dir=       - indicate dir with logging-operator repo
      -c=, --cloud=     - specify what orchestrator you use, openshift or kubernetes (k8s)

      Example:
      ./autofill_manifests.sh -c=k8s --dir=/tmp/logging-operator -n=test-test
      "
else
  echo "error, incorrect arguments"
fi
