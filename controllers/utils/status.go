package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	InProgress = "In Progress"
	Success    = "Successful"
	Failed     = "Failed"
)

type StatusUpdater struct {
	resource *loggingService.LoggingService
	client   client.Client
	log      logr.Logger
}

func NewStatusUpdater(client client.Client, resource *loggingService.LoggingService) StatusUpdater {
	return StatusUpdater{
		client:   client,
		log:      Logger("status"),
		resource: resource,
	}
}

func (updater *StatusUpdater) GetCondition(newConditionReason string) (int, *loggingService.LoggingServiceCondition) {
	if len(newConditionReason) != 0 && len(updater.resource.Status.Conditions) != 0 {
		for i := range updater.resource.Status.Conditions {
			if updater.resource.Status.Conditions[i].Reason == newConditionReason {
				return i, &updater.resource.Status.Conditions[i]
			}
		}

		return -1, nil
	} else {
		return -1, nil
	}
}

func (updater *StatusUpdater) RemoveStatus(reason string) bool {
	if index, condition := updater.GetCondition(reason); condition != nil {
		updater.resource.Status.Conditions[index] = updater.resource.Status.Conditions[len(updater.resource.Status.Conditions)-1]
		updater.resource.Status.Conditions = updater.resource.Status.Conditions[:len(updater.resource.Status.Conditions)-1]

		if err := updater.patch(); err != nil {
			updater.log.Error(err, fmt.Sprintf("Failed to remove status for service %s", reason))
		}

		updater.log.V(Debug).Info(fmt.Sprintf("Status for service %s completed and successfully removed", reason))
		return true
	} else {
		updater.log.V(Warn).Info(fmt.Sprintf("Status for service %s is empty and cannot be completed", reason))
		return false
	}
}

func (updater *StatusUpdater) UpdateStatus(reason string, statusType string, status bool, message string) {
	newCondition := loggingService.LoggingServiceCondition{
		Type:               statusType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: meta.Now().String(),
	}

	index, oldCondition := updater.GetCondition(newCondition.Reason)
	isStatusChanged := false

	if oldCondition != nil {
		if newCondition.Type == Failed && IsStatusEqual(newCondition, oldCondition) {
			updater.log.V(Debug).Info(fmt.Sprintf("Status for Service %s the same and should not change", reason))
		} else {
			updater.resource.Status.Conditions[index] = newCondition
			isStatusChanged = true
		}
	} else {
		updater.resource.Status.Conditions = append(updater.resource.Status.Conditions, newCondition)
		isStatusChanged = true
	}

	if isStatusChanged {
		if err := updater.patch(); err != nil {
			updater.log.Error(err, fmt.Sprintf("Update the status for service %s failed", reason))
			return
		}

		updater.log.V(Debug).Info(fmt.Sprintf("Status for service %s successfully changed to %s", reason, statusType))
	}
}

func (updater *StatusUpdater) patch() error {

	resourceBuf, err := json.Marshal(updater.resource)
	if err != nil {
		updater.log.Error(err, "failed to marshal object")
		return err
	}
	object := &unstructured.Unstructured{Object: map[string]interface{}{}}
	if err = json.Unmarshal(resourceBuf, object); err != nil {
		updater.log.Error(err, "failed to unmarshal object")
		return err
	}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"status": object.Object["status"],
	})
	if err != nil {
		updater.log.Error(err, "failed to marshal object")
		return err
	}

	patch := client.RawPatch(types.MergePatchType, mergePatch)

	if err := updater.client.Status().Patch(context.TODO(), updater.resource, patch); err != nil {
		return err
	}
	return nil
}

func IsStatusEqual(source loggingService.LoggingServiceCondition, target *loggingService.LoggingServiceCondition) bool {
	// Exclude LastTransitionTime from comparison
	source.LastTransitionTime = target.LastTransitionTime

	return reflect.DeepEqual(&source, target)
}

func (updater *StatusUpdater) RemoveTemporaryStatuses() {
	isStatusChanged := false

	for i := len(updater.resource.Status.Conditions) - 1; i >= 0; i-- {
		if updater.resource.Status.Conditions[i].Reason == LoggingServiceStatus {
			continue
		}

		if updater.resource.Status.Conditions[i].Type == Failed {
			if updater.resource.Status.Conditions[i].Reason == GraylogStatus && updater.resource.Spec.Graylog.IsInstall() {
				continue
			} else if updater.resource.Status.Conditions[i].Reason == FluentdStatus && updater.resource.Spec.Fluentd.IsInstall() {
				continue
			} else if updater.resource.Status.Conditions[i].Reason == FluentbitStatus && updater.resource.Spec.Fluentbit.IsInstall() {
				continue
			} else if updater.resource.Status.Conditions[i].Reason == EventsReaderStatus && updater.resource.Spec.CloudEventsReader.IsInstall() {
				continue
			} else if updater.resource.Status.Conditions[i].Reason == MonitoringAgentStatus && updater.resource.Spec.MonitoringAgentLoggingPlugin.IsInstall() {
				continue
			} else if updater.resource.Status.Conditions[i].Reason == ComponentPendingStatus {
				continue
			}
		}

		updater.resource.Status.Conditions = updater.resource.Status.Conditions[:len(updater.resource.Status.Conditions)-1]
		isStatusChanged = true
	}

	if isStatusChanged {
		if err := updater.patch(); err != nil {
			updater.log.Error(err, "Remove all statuses failed")
		}

		updater.log.V(Debug).Info("Status for service  successfully changed to")
	}
}

func (updater *StatusUpdater) IsStatusFailed(conditionReason string) bool {
	_, condition := updater.GetCondition(conditionReason)
	return condition != nil && condition.Type == Failed
}
