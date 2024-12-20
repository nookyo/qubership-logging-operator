package fluentbit

import (
	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FluentbitReconciler struct {
	*util.ComponentReconciler
	ComponentList     *[]util.Component
	DynamicParameters util.DynamicParameters
}

func NewFluentbitReconciler(client client.Client, scheme *runtime.Scheme, updater util.StatusUpdater, pendingComponents *[]util.Component, dynamicParameters util.DynamicParameters) FluentbitReconciler {
	return FluentbitReconciler{
		ComponentReconciler: &util.ComponentReconciler{
			Client:        client,
			Scheme:        scheme,
			Log:           util.Logger("fluentbit"),
			StatusUpdater: updater,
		},
		ComponentList:     pendingComponents,
		DynamicParameters: dynamicParameters,
	}
}

// Run reconciles fluentbit custom resource.
// Creates new DaemonSet, ConfigMap, Service if its don't exist.
func (r *FluentbitReconciler) Run(cr *loggingService.LoggingService) error {
	if !r.StatusUpdater.IsStatusFailed(util.FluentbitStatus) {
		r.StatusUpdater.UpdateStatus(util.FluentbitStatus, util.InProgress, false, "Start reconcile of Fluentbit")
	}
	r.Log.Info("Start Fluentbit reconciliation")

	if cr.Spec.Fluentbit != nil && cr.Spec.Fluentbit.IsInstall() && (cr.Spec.Fluentbit.Aggregator == nil || !cr.Spec.Fluentbit.Aggregator.Install) {
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
				ComponentName: util.FluentbitComponentName,
				StatusName:    util.FluentbitStatus,
			},
		)
	} else {
		r.Log.Info("Uninstalling component if exists")
		r.uninstall(cr)
		r.StatusUpdater.RemoveStatus(util.FluentbitStatus)
	}
	r.Log.Info("Component reconciled")
	return nil
}

// uninstall deletes all resources related to the component
func (r *FluentbitReconciler) uninstall(cr *loggingService.LoggingService) {
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
