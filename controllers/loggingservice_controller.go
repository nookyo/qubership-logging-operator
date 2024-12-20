package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	events_reader "github.com/Netcracker/qubership-logging-operator/controllers/events-reader"
	"github.com/Netcracker/qubership-logging-operator/controllers/fluentbit"
	fluentbit_forwarder_aggregator "github.com/Netcracker/qubership-logging-operator/controllers/fluentbit-forwarder-aggregator"
	"github.com/Netcracker/qubership-logging-operator/controllers/fluentd"
	"github.com/Netcracker/qubership-logging-operator/controllers/graylog"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	InitialTimeoutOnFailedReconcile = time.Second
	DefaultContainerRuntimeType     = "containerd"
)

type LoggingServiceReconciler struct {
	Config                   *rest.Config
	Scheme                   *runtime.Scheme
	TimeoutOnFailedReconcile time.Duration
	Client                   client.Client
	Log                      logr.Logger
	StatusUpdater            util.StatusUpdater
	DynamicParameters        util.DynamicParameters
}

// +kubebuilder:rbac:groups=logging.qubership.org,resources=loggingservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=logging.qubership.org,resources=loggingservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=logging.qubership.org,resources=loggingservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *LoggingServiceReconciler) Reconcile(context context.Context, request ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Start reconcile cycle of Logging Service")
	initialTime := time.Now()

	customResourceInstance := &loggingService.LoggingService{}
	err := r.Client.Get(context, request.NamespacedName, customResourceInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{Requeue: false}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	r.StatusUpdater = util.NewStatusUpdater(r.Client, customResourceInstance)
	clientSet := kubernetes.NewForConfigOrDie(r.Config)
	if !r.ReconcileLoggingServiceCluster(context, customResourceInstance, clientSet) {
		var reconcileTime = time.Since(initialTime)
		r.Log.V(util.Error).Info(fmt.Sprintf("Reconcile of Logging Service was failed with error in %s. Next reconcile cycle after %s",
			util.ToString(reconcileTime), r.TimeoutOnFailedReconcile.String()))
		r.StatusUpdater.UpdateStatus(util.LoggingServiceStatus, util.Failed, false, "Reconcile of Logging service failed")

		var result = reconcile.Result{RequeueAfter: r.TimeoutOnFailedReconcile}
		r.TimeoutOnFailedReconcile = r.TimeoutOnFailedReconcile * 2

		return result, nil
	}

	var reconcileTime = time.Since(initialTime)

	r.StatusUpdater.UpdateStatus(util.LoggingServiceStatus, util.Success, true, "Reconcile of Logging service succeeded")
	r.TimeoutOnFailedReconcile = InitialTimeoutOnFailedReconcile

	r.Log.Info(fmt.Sprintf("Reconcile a cycle of Logging Service successfully finished in %s", util.ToString(reconcileTime)))

	return ctrl.Result{}, nil
}

func (r *LoggingServiceReconciler) ReconcileLoggingServiceCluster(ctx context.Context, customResourceInstance *loggingService.LoggingService, clientSet kubernetes.Interface) bool {
	//If operator reconcile failed, then we keep failed status for all following reconcile cycles, before first success
	if !r.StatusUpdater.IsStatusFailed(util.LoggingServiceStatus) {
		r.StatusUpdater.UpdateStatus(util.LoggingServiceStatus, util.InProgress, false, "Logging Service reconcile cycle in progress")
	}

	r.StatusUpdater.RemoveTemporaryStatuses()

	r.updateDynamicParameters(customResourceInstance)

	var isDeployServiceSuccess = true
	var pendingComponents []util.Component

	graylogReconciler := graylog.NewGraylogReconciler(r.Client, r.Scheme, r.StatusUpdater)
	if err := graylogReconciler.Run(ctx, customResourceInstance, clientSet); err != nil {
		isDeployServiceSuccess = false
		r.Log.Error(err, "Deploy of Graylog is failed")
		graylogReconciler.StatusUpdater.UpdateStatus(util.GraylogStatus, util.Failed, false, fmt.Sprintf("Reason: %s", err.Error()))
	}

	fluentdReconciler := fluentd.NewFluentdReconciler(r.Client, r.Scheme, r.StatusUpdater, &pendingComponents, r.DynamicParameters)
	if err := fluentdReconciler.Run(customResourceInstance); err != nil {
		isDeployServiceSuccess = false
		r.Log.Error(err, "Deploy of Fluentd is failed")
		fluentdReconciler.StatusUpdater.UpdateStatus(util.FluentdStatus, util.Failed, false, fmt.Sprintf("Reason: %s", err.Error()))
	}

	fluentbitReconciler := fluentbit.NewFluentbitReconciler(r.Client, r.Scheme, r.StatusUpdater, &pendingComponents, r.DynamicParameters)
	if err := fluentbitReconciler.Run(customResourceInstance); err != nil {
		isDeployServiceSuccess = false
		r.Log.Error(err, "Deploy of Fluentbit is failed")
		fluentbitReconciler.StatusUpdater.UpdateStatus(util.FluentbitStatus, util.Failed, false, fmt.Sprintf("Reason: %s", err.Error()))
	}

	fluentsReconciler := fluentbit_forwarder_aggregator.NewHAFluentReconciler(r.Client, r.Scheme, r.StatusUpdater, &pendingComponents, r.DynamicParameters)
	if err := fluentsReconciler.Run(customResourceInstance); err != nil {
		isDeployServiceSuccess = false
		r.Log.Error(err, "Deploy of Fluentbit forwarder-aggregator is failed")
		fluentsReconciler.StatusUpdater.UpdateStatus(util.FluentbitStatus, util.Failed, false, fmt.Sprintf("Reason: %s", err.Error()))
	}

	eventsReaderReconciler := events_reader.NewEventsReaderReconciler(r.Client, r.Scheme, r.StatusUpdater, &pendingComponents)
	if err := eventsReaderReconciler.Run(customResourceInstance); err != nil {
		isDeployServiceSuccess = false
		r.Log.Error(err, "Deploy of Cloud Events Reader is failed")
		eventsReaderReconciler.StatusUpdater.UpdateStatus(util.EventsReaderStatus, util.Failed, false, fmt.Sprintf("Reason: %s", err.Error()))
	}

	statusReconciler := util.NewComponentsPendingReconciler(r.Client, r.Scheme, r.StatusUpdater, &pendingComponents)
	status, err := statusReconciler.Run(customResourceInstance)
	if err != nil {
		isDeployServiceSuccess = false
		r.Log.Error(err, "Failed waiting for component statuses")
		statusReconciler.StatusUpdater.UpdateStatus(util.ComponentPendingStatus, util.Failed, false, fmt.Sprintf("Reason: %s", err.Error()))
	} else if !status {
		isDeployServiceSuccess = false
	}

	return isDeployServiceSuccess
}

func (r *LoggingServiceReconciler) updateDynamicParameters(customResourceInstance *loggingService.LoggingService) {
	if customResourceInstance.Spec.ContainerRuntimeType != "" {
		r.DynamicParameters.ContainerRuntimeType = customResourceInstance.Spec.ContainerRuntimeType
		r.Log.Info(fmt.Sprintf("Container Runtime set in custom resource. '%s' is used.", customResourceInstance.Spec.ContainerRuntimeType))
	} else {
		if r.DynamicParameters.ContainerRuntimeType == "" {
			nodes := &corev1.NodeList{}
			if err := r.Client.List(context.TODO(), nodes); err != nil || len(nodes.Items) == 0 {
				r.Log.Error(err, "Could get nodes list")
				r.DynamicParameters.ContainerRuntimeType = DefaultContainerRuntimeType
				r.Log.Info(fmt.Sprintf("Container Runtime is not set in custom resource and can not be discovered. By default '%s' is used", DefaultContainerRuntimeType))
				return
			} else {
				nodeContainerRuntimes := make(map[string]int)
				for _, node := range nodes.Items {
					containerRuntimeVersion := node.Status.NodeInfo.ContainerRuntimeVersion
					if len(containerRuntimeVersion) > 0 {
						versions := strings.Split(containerRuntimeVersion, "://")
						customResourceInstance.Spec.ContainerRuntimeType = versions[0]
						nodeContainerRuntimes[versions[0]] = nodeContainerRuntimes[versions[0]] + 1
					}
				}
				if len(nodeContainerRuntimes) > 1 {
					r.Log.Info("There found more than one container runtime types", "container runtimes", nodeContainerRuntimes)
					r.DynamicParameters.ContainerRuntimeType = DefaultContainerRuntimeType
				} else if len(nodeContainerRuntimes) == 1 {
					for crv := range nodeContainerRuntimes {
						customResourceInstance.Spec.ContainerRuntimeType = crv
						r.DynamicParameters.ContainerRuntimeType = crv
					}
				}
			}
		}
		r.Log.Info(fmt.Sprintf("Container Runtime discovered. '%s' is used.", customResourceInstance.Spec.ContainerRuntimeType))
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoggingServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingService.LoggingService{}).
		WithEventFilter(ignoreDeletionPredicate()).
		Complete(r)
}

func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}
