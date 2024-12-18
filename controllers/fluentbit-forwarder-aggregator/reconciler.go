package fluentbit_forwarder_aggregator

import (
	"errors"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HAFluentReconciler struct {
	*util.ComponentReconciler
	ComponentList     *[]util.Component
	DynamicParameters util.DynamicParameters
}

func NewHAFluentReconciler(client client.Client, scheme *runtime.Scheme, updater util.StatusUpdater, pendingComponents *[]util.Component, dynamicParameters util.DynamicParameters) HAFluentReconciler {
	return HAFluentReconciler{
		ComponentReconciler: &util.ComponentReconciler{
			Client:        client,
			Scheme:        scheme,
			Log:           util.Logger("fluentbit-forwarder-aggregator"),
			StatusUpdater: updater,
		},
		ComponentList:     pendingComponents,
		DynamicParameters: dynamicParameters,
	}
}

// Run reconciles fluentbit-forwarder-aggregator custom resource.
// Creates new DaemonSet, ConfigMap, Service if its don't exist.
func (r *HAFluentReconciler) Run(cr *loggingService.LoggingService) error {
	if !r.StatusUpdater.IsStatusFailed(util.HAFluentStatus) {
		r.StatusUpdater.UpdateStatus(util.HAFluentStatus, util.InProgress, false, "Start reconcile of Fluentbit-Forwarder-Aggregator")
	}
	r.Log.Info("Start Fluentbit-Forwarder-Aggregator reconciliation")

	if cr.Spec.Fluentbit != nil && cr.Spec.Fluentbit.IsInstall() && cr.Spec.Fluentbit.Aggregator != nil && cr.Spec.Fluentbit.Aggregator.Install {
		if cr.Spec.Fluentbit.Aggregator.GraylogOutput && (cr.Spec.Fluentbit.Aggregator.GraylogHost == "" || cr.Spec.Fluentbit.Aggregator.GraylogPort == 0) {
			err := errors.New("configuration error: fluentbit.aggregator.graylogHost and fluentbit.aggregator.graylogPort are required with Graylog output")
			r.Log.Error(err, "configuration of fluentbit aggregator is incorrect")
			return err
		}
		if err := r.handleAggregatorConfigMap(cr); err != nil {
			r.Log.Error(err, "error occurred in handleAggregatorConfigMap")
			return err
		}
		if err := r.handleAggregatorStatefulSet(cr); err != nil {
			r.Log.Error(err, "error occurred in handleAggregatorStatefulSet")
			return err
		}
		if err := r.handleAggregatorService(cr); err != nil {
			r.Log.Error(err, "error occurred in handleAggregatorService")
			return err
		}

		if err := r.handleForwarderConfigMap(cr); err != nil {
			r.Log.Error(err, "error occurred in handleForwarderConfigMap")
			return err
		}
		if err := r.handleForwarderDaemonSet(cr); err != nil {
			r.Log.Error(err, "error occurred in handleForwarderDaemonSet")
			return err
		}
		if err := r.handleForwarderService(cr); err != nil {
			r.Log.Error(err, "error occurred in handleForwarderService")
			return err
		}

		*r.ComponentList = append(
			*r.ComponentList,
			util.Component{
				ComponentName: util.ForwarderFluentbitComponentName,
				StatusName:    util.HAFluentStatus,
			},
		)
	} else {
		r.Log.Info("Uninstalling component if exists")
		r.uninstall(cr)
		r.StatusUpdater.RemoveStatus(util.HAFluentStatus)
	}
	r.Log.Info("Component reconciled")
	return nil
}

// uninstall deletes all resources related to the component
func (r *HAFluentReconciler) uninstall(cr *loggingService.LoggingService) {
	if err := r.deleteDaemonSet(cr, util.ForwarderFluentbitComponentName); err != nil {
		r.Log.Error(err, "Can not delete Daemon Set")
	}
	if err := r.deleteConfigMap(cr, util.ForwarderFluentbitComponentName); err != nil {
		r.Log.Error(err, "Can not delete Config Map")
	}
	if err := r.deleteService(cr, util.ForwarderFluentbitComponentName); err != nil {
		r.Log.Error(err, "Can not delete Service")
	}
	if err := r.deleteStatefulSet(cr, util.AggregatorFluentbitComponentName); err != nil {
		r.Log.Error(err, "Can not delete Stateful Set")
	}
	if err := r.deleteConfigMap(cr, util.AggregatorFluentbitComponentName); err != nil {
		r.Log.Error(err, "Can not delete Config Map")
	}
	if err := r.deleteService(cr, util.AggregatorFluentbitComponentName); err != nil {
		r.Log.Error(err, "Can not delete Service")
	}
}
