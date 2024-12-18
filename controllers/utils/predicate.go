package utils

import (
	"fmt"
	"reflect"

	logging "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type SkipStatusUpdatePredicate struct {
	log logr.Logger
}

func NewPredicate(logger logr.Logger) SkipStatusUpdatePredicate {
	return SkipStatusUpdatePredicate{
		log: logger,
	}
}

func (predicate *SkipStatusUpdatePredicate) Create(event event.CreateEvent) bool {
	predicate.log.V(Debug).Info(fmt.Sprintf("Raised Create Event on %s", reflect.TypeOf(event.Object)))
	return true
}

func (predicate *SkipStatusUpdatePredicate) Delete(event event.DeleteEvent) bool {
	predicate.log.V(Debug).Info(fmt.Sprintf("Raised Delete Event on %s", reflect.TypeOf(event.Object)))
	return true
}

func (predicate *SkipStatusUpdatePredicate) Update(event event.UpdateEvent) bool {
	if !predicate.IsStatusUpdated(event.ObjectOld, event.ObjectNew) {
		predicate.log.V(Debug).Info(fmt.Sprintf("Raised Update Event on %s", reflect.TypeOf(event.ObjectOld)),
			"old", ToJSON(event.ObjectOld),
			"new", ToJSON(event.ObjectNew))
		return true
	} else {
		return false
	}
}

func (predicate *SkipStatusUpdatePredicate) Generic(event event.GenericEvent) bool {
	predicate.log.V(Debug).Info(fmt.Sprintf("Raised Generic Event on %s", reflect.TypeOf(event.Object)))
	return true
}

func (predicate *SkipStatusUpdatePredicate) IsStatusUpdated(oldInstance runtime.Object, newInstance runtime.Object) bool {
	oldLoggingService, oldOk := oldInstance.(*logging.LoggingService)
	newLoggingService, newOk := newInstance.(*logging.LoggingService)

	if oldOk && newOk {
		return reflect.DeepEqual(oldLoggingService.Spec, newLoggingService.Spec)
	} else {
		return false
	}
}
