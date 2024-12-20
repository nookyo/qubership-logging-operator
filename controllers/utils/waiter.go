package utils

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Interval              = time.Second * 10
	TerminatingPodTimeout = time.Minute * 2
)

func (manager *PodManager) WaitForJobSucceeded(jobName string, timeout time.Duration) (bool, error) {
	var initialTime = time.Now()
	var isSucceeded = false

	manager.log.Info(fmt.Sprintf("Wait for %s job is succeeded", jobName))
	err := wait.PollUntilContextTimeout(context.TODO(), Interval, timeout, true, func(ctx context.Context) (bool, error) {
		isJobSucceeded, err := manager.IsJobSucceeded(jobName)
		isSucceeded = isJobSucceeded
		if err != nil {
			return false, err
		}

		if isSucceeded {
			manager.log.Info(fmt.Sprintf("Job %s sucessfully finished in %s",
				jobName, time.Since(initialTime)))
			return true, err
		}

		return false, err
	})

	if err != nil {
		if wait.Interrupted(err) {
			return isSucceeded, nil
		} else {
			return isSucceeded, err
		}
	}

	return isSucceeded, nil
}

func (manager *PodManager) WaitForStatefulsetUpdated(serviceName string, timeout time.Duration) (bool, error) {
	var initialTime = time.Now()
	var isAvailable = false

	manager.log.Info(fmt.Sprintf("Wait for %s service is updated", serviceName))
	err := wait.PollUntilContextTimeout(context.TODO(), Interval, timeout, true, func(ctx context.Context) (bool, error) {
		isAvailableInternal, err := manager.IsStatefulsetReplicasSynchronised(serviceName)
		isAvailable = isAvailableInternal
		if err != nil {
			return false, err
		}

		if isAvailable {
			manager.log.Info(fmt.Sprintf("Service %s sucessfully updated in %s",
				serviceName, time.Since(initialTime)))
			return true, err
		}

		return false, err
	})

	if err != nil {
		if wait.Interrupted(err) {
			return isAvailable, nil
		} else {
			return isAvailable, err
		}
	}

	return isAvailable, nil
}

func WaitForHostActive(host string, port int, timeout time.Duration) error {
	url := host + ":" + strconv.Itoa(port)
	res := error(nil)
	for i := 1; i < 5; i++ {
		_, err := net.DialTimeout("tcp", url, timeout)
		if err != nil {
			logger.Info(fmt.Sprintf("Host %s is unreachable", url))
			time.Sleep(10 * time.Second)
			res = err
		} else {
			logger.Info(fmt.Sprintf("Host %s is active", url))
			return nil
		}
	}
	return res
}

func (manager *PodManager) WaitForTerminatingPods(terminatedPodNames []string, labelSelectors map[string]string) (bool, error) {
	if len(terminatedPodNames) > 0 {
		manager.log.Info(fmt.Sprintf("Wait terminating pods %q", terminatedPodNames))
		err := wait.PollUntilContextTimeout(context.TODO(), Interval, TerminatingPodTimeout, true, func(ctx context.Context) (bool, error) {
			podList, err := manager.FindPods(labelSelectors)
			if err != nil {
				return true, err
			}

			for _, deletedPodName := range terminatedPodNames {
				for _, pod := range podList.Items {
					if pod.Name == deletedPodName {
						manager.log.V(Debug).Info(fmt.Sprintf("Pod %s still not terminated", pod.Name))
						return false, nil
					}
				}
			}

			return true, nil
		})

		if err != nil {
			if wait.Interrupted(err) {
				return false, nil
			} else {
				return false, err
			}
		}
	}

	return true, nil
}

type Component struct {
	ComponentName string
	StatusName    string
}

type ComponentsPendingReconciler struct {
	*ComponentReconciler
	ComponentList *[]Component
}

func NewComponentsPendingReconciler(client client.Client, scheme *runtime.Scheme, updater StatusUpdater, pendingComponents *[]Component) ComponentsPendingReconciler {
	return ComponentsPendingReconciler{
		ComponentReconciler: &ComponentReconciler{
			Client:        client,
			Scheme:        scheme,
			Log:           Logger("components-pending"),
			StatusUpdater: updater,
		},
		ComponentList: pendingComponents,
	}
}

func (r *ComponentsPendingReconciler) Run(cr *loggingService.LoggingService) (bool, error) {
	if !r.StatusUpdater.IsStatusFailed(ComponentPendingStatus) {
		r.StatusUpdater.UpdateStatus(ComponentPendingStatus, InProgress, false, "Start reconcile of ComponentsPending")
	}
	r.Log.Info("Waiting for component statuses")

	isStatusSuccess := true

	if len(*r.ComponentList) != 0 {
		// Delay to allow time for the deployments and daemonsets to update
		time.Sleep(InitialDelay)

		podManager := NewPodManager(r.Client, cr.GetNamespace(), r.Log)
		start := time.Now()
		for {
			if time.Since(start) >= ComponentPendingTimeout {
				isStatusSuccess = false
				r.Log.Info("Timeout waiting for component statuses")
				for _, component := range *r.ComponentList {
					r.Log.Error(fmt.Errorf("%s is not started", component.ComponentName), fmt.Sprintf("Deploy of the %s is failed", component.ComponentName))
					r.StatusUpdater.UpdateStatus(component.StatusName, Failed, false, fmt.Sprintf("Reason: %s is not started", component.ComponentName))
				}
				break
			}

			for i := len(*r.ComponentList) - 1; i >= 0; i-- {
				component := (*r.ComponentList)[i]
				isAvailable := false
				var err error

				switch component.ComponentName {
				case FluentdComponentName, FluentbitComponentName, ForwarderFluentbitComponentName:
					isAvailable, err = podManager.IsDaemonSetAvailable(component.ComponentName)
				default:
					isAvailable, err = podManager.IsDeploymentReplicasSynchronised(component.ComponentName)
				}
				if err != nil {
					return false, err
				}
				if isAvailable {
					r.Log.Info(fmt.Sprintf("The %s component is started", component.ComponentName))
					r.StatusUpdater.RemoveStatus(component.StatusName)
					(*r.ComponentList)[i] = (*r.ComponentList)[len(*r.ComponentList)-1]
					*r.ComponentList = (*r.ComponentList)[:len(*r.ComponentList)-1]
				}
			}

			if len(*r.ComponentList) == 0 {
				break
			}

			r.Log.V(Debug).Info(fmt.Sprintf("Elapsed time: %s (timeout: %s). Not yet started components: %s. Pause %s.", time.Since(start), ComponentPendingTimeout, r.ToComponentNameList(r.ComponentList), Interval))
			time.Sleep(Interval)
		}
	}

	r.Log.Info("Component reconciled")
	r.StatusUpdater.RemoveStatus(ComponentPendingStatus)
	return isStatusSuccess, nil
}

func (r *ComponentsPendingReconciler) ToComponentNameList(componentList *[]Component) []string {
	var componentNameList []string

	for _, component := range *componentList {
		componentNameList = append(componentNameList, component.ComponentName)
	}

	return componentNameList
}
