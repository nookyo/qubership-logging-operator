package fluentd

import (
	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FluentdReconciler struct {
	*util.ComponentReconciler
	ComponentList     *[]util.Component
	DynamicParameters util.DynamicParameters
}

func NewFluentdReconciler(client client.Client, scheme *runtime.Scheme, updater util.StatusUpdater, pendingComponents *[]util.Component, dynamicParameters util.DynamicParameters) FluentdReconciler {
	return FluentdReconciler{
		ComponentReconciler: &util.ComponentReconciler{
			Client:        client,
			Scheme:        scheme,
			Log:           util.Logger("fluentd"),
			StatusUpdater: updater,
		},
		ComponentList:     pendingComponents,
		DynamicParameters: dynamicParameters,
	}
}

// Run reconciles fluentd custom resource.
// Creates new DaemonSet, ConfigMap, Service if its don't exist.
func (r *FluentdReconciler) Run(cr *loggingService.LoggingService) error {
	if !r.StatusUpdater.IsStatusFailed(util.FluentdStatus) {
		r.StatusUpdater.UpdateStatus(util.FluentdStatus, util.InProgress, false, "Start reconcile of Fluentd")
	}
	r.Log.Info("Start Fluentd reconciliation")

	if cr.Spec.Fluentd != nil && cr.Spec.Fluentd.IsInstall() {
		if err := r.handleConfigMap(cr); err != nil {
			return err
		}
		if err := r.handleDaemonSet(cr); err != nil {
			return err
		}
		if err := r.handleService(cr); err != nil {
			return err
		}
		*r.ComponentList = append(
			*r.ComponentList,
			util.Component{
				ComponentName: util.FluentdComponentName,
				StatusName:    util.FluentdStatus,
			},
		)
	} else {
		r.Log.Info("Uninstalling component if exists")
		r.uninstall(cr)
		r.StatusUpdater.RemoveStatus(util.FluentdStatus)
	}
	r.Log.Info("Component reconciled")
	return nil
}

// uninstall deletes all resources related to the component
func (r *FluentdReconciler) uninstall(cr *loggingService.LoggingService) {
	if err := r.deleteDaemonSet(cr); err != nil {
		r.Log.Error(err, "Can not delete DaemonSet")
	}
	if err := r.deleteConfigMap(cr); err != nil {
		r.Log.Error(err, "Can not delete ConfigMap")
	}
	if err := r.deleteService(cr); err != nil {
		r.Log.Error(err, "Can not delete Service")
	}
}
