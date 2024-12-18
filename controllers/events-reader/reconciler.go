package events_reader

import (
	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EventsReaderReconciler struct {
	*util.ComponentReconciler
	ComponentList *[]util.Component
}

func NewEventsReaderReconciler(client client.Client, scheme *runtime.Scheme, updater util.StatusUpdater, pendingComponents *[]util.Component) EventsReaderReconciler {
	return EventsReaderReconciler{
		ComponentReconciler: &util.ComponentReconciler{
			Client:        client,
			Scheme:        scheme,
			Log:           util.Logger("events-reader"),
			StatusUpdater: updater,
		},
		ComponentList: pendingComponents,
	}
}

// Run reconciles EventsReader custom resource.
// Creates new Deployment, Service if its don't exist.
func (r *EventsReaderReconciler) Run(cr *loggingService.LoggingService) error {
	if !r.StatusUpdater.IsStatusFailed(util.EventsReaderStatus) {
		r.StatusUpdater.UpdateStatus(util.EventsReaderStatus, util.InProgress, false, "Start reconcile of Events Reader")
	}
	r.Log.Info("Start Events Reader reconciliation")

	if cr.Spec.CloudEventsReader != nil && cr.Spec.CloudEventsReader.IsInstall() {
		if err := r.handleDeployment(cr); err != nil {
			return err
		}
		if err := r.handleService(cr); err != nil {
			return err
		}

		*r.ComponentList = append(
			*r.ComponentList,
			util.Component{
				ComponentName: util.EventsReaderComponentName,
				StatusName:    util.EventsReaderStatus,
			},
		)
	} else {
		r.Log.Info("Uninstalling component if exists")
		r.uninstall(cr)
		r.StatusUpdater.RemoveStatus(util.EventsReaderStatus)
	}
	r.Log.Info("Component reconciled")
	return nil
}

// uninstall deletes all resources related to the component
func (r *EventsReaderReconciler) uninstall(cr *loggingService.LoggingService) {
	if err := r.deleteDeployment(cr); err != nil {
		r.Log.Error(err, "Can not delete Deployment")
	}
	if err := r.deleteService(cr); err != nil {
		r.Log.Error(err, "Can not delete Service")
	}
	if err := r.deleteServiceAccount(cr); err != nil {
		r.Log.Error(err, "Can not delete ServiceAccount")
	}
}
