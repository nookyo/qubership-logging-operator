package utils

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodManager struct {
	client    client.Client
	log       logr.Logger
	namespace string
}

func NewPodManager(client client.Client, namespace string, log logr.Logger) PodManager {
	return PodManager{
		client:    client,
		log:       log,
		namespace: namespace,
	}
}

func (manager *PodManager) DeletePods(labelSelectors map[string]string) (bool, error) {
	podList, err := manager.FindPods(labelSelectors)
	if err != nil {
		return false, err
	}

	for _, pod := range podList.Items {
		err = manager.client.Delete(context.TODO(), &pod)
		if err != nil {
			return false, err
		}
	}

	podNameList := manager.ToPodNameList(podList)
	done, err := manager.WaitForTerminatingPods(podNameList, labelSelectors)

	if done {
		manager.log.Info(fmt.Sprintf("Pods %q terminated successfully", podNameList))
	}

	return done, err
}

func (manager *PodManager) ToPodNameList(podList *core.PodList) []string {
	var podNameList []string

	for _, pod := range podList.Items {
		podNameList = append(podNameList, pod.Name)
	}

	return podNameList
}

func (manager *PodManager) FindPods(labelSelectors map[string]string) (podList *core.PodList, err error) {
	manager.log.V(Debug).Info(fmt.Sprintf("Will try to find Pod(s) with labels %q", labelSelectors))

	podList = &core.PodList{}
	listOps := &client.ListOptions{
		Namespace:     manager.namespace,
		LabelSelector: labels.SelectorFromSet(labelSelectors),
	}

	if err := manager.client.List(context.Background(), podList, listOps); err != nil {
		if errors.IsNotFound(err) {
			manager.log.V(Debug).Info("Pod doesn't exist yet.")
			return podList, nil
		}

		return podList, err
	}

	return podList, nil
}

func (manager *PodManager) IsDaemonSetExists(serviceName string) (bool, error) {
	daemonSet := &appsv1.DaemonSet{}

	err := manager.client.Get(context.TODO(), types.NamespacedName{
		Name: serviceName, Namespace: manager.namespace,
	}, daemonSet)

	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (manager *PodManager) IsDaemonSetAvailable(serviceName string) (bool, error) {
	daemonSet := &appsv1.DaemonSet{}

	err := manager.client.Get(context.TODO(), types.NamespacedName{
		Name: serviceName, Namespace: manager.namespace,
	}, daemonSet)
	if err != nil {
		return false, err
	}

	manager.log.V(Debug).Info(fmt.Sprintf("Service: %s. NumberAvailable: %d, DesiredNumberScheduled: %d, UpdatedNumberScheduled: %d.",
		serviceName, daemonSet.Status.NumberAvailable, daemonSet.Status.DesiredNumberScheduled, daemonSet.Status.UpdatedNumberScheduled))

	return daemonSet.Status.UpdatedNumberScheduled == daemonSet.Status.DesiredNumberScheduled &&
		daemonSet.Status.NumberAvailable == daemonSet.Status.DesiredNumberScheduled, nil
}

func (manager *PodManager) IsDeploymentExists(serviceName string) (bool, error) {
	daemonSet := &appsv1.Deployment{}

	err := manager.client.Get(context.TODO(), types.NamespacedName{
		Name: serviceName, Namespace: manager.namespace,
	}, daemonSet)

	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (manager *PodManager) IsDeploymentReplicasSynchronised(serviceName string) (bool, error) {
	deploy := &appsv1.Deployment{}

	err := manager.client.Get(context.TODO(), types.NamespacedName{
		Name: serviceName, Namespace: manager.namespace,
	}, deploy)
	if err != nil {
		return false, err
	}

	manager.log.V(Debug).Info(fmt.Sprintf("Service: %s. AvailableReplicas: %d, UpdatedReplicas: %d, Total replicas: %d",
		serviceName, deploy.Status.AvailableReplicas, deploy.Status.UpdatedReplicas, *deploy.Spec.Replicas))

	return deploy.Status.AvailableReplicas == *deploy.Spec.Replicas && deploy.Status.UpdatedReplicas == *deploy.Spec.Replicas, nil
}

func (manager *PodManager) IsStatefulsetReplicasSynchronised(serviceName string) (bool, error) {
	ss := &appsv1.StatefulSet{}

	err := manager.client.Get(context.TODO(), types.NamespacedName{
		Name: serviceName, Namespace: manager.namespace,
	}, ss)
	if err != nil {
		return false, err
	}

	manager.log.V(Debug).Info(fmt.Sprintf("Service: %s. AvailableReplicas: %d, ReadyReplicas: %d, Total replicas: %d.",
		serviceName, ss.Status.Replicas, ss.Status.ReadyReplicas, *ss.Spec.Replicas))

	return ss.Status.Replicas == *ss.Spec.Replicas && ss.Status.ReadyReplicas == *ss.Spec.Replicas, nil
}

func (manager *PodManager) IsJobSucceeded(jobName string) (bool, error) {
	job := &batchv1.Job{}

	err := manager.client.Get(context.TODO(), types.NamespacedName{
		Name: jobName, Namespace: manager.namespace,
	}, job)
	if err != nil {
		return false, err
	}

	manager.log.V(Debug).Info(fmt.Sprintf("Job: %s. RunningPods: %d, FailedPods: %d, SucceededPods: %d.",
		jobName, job.Status.Active, job.Status.Failed, job.Status.Succeeded))

	return job.Status.Active == 0 && job.Status.Succeeded == *job.Spec.Completions, nil
}
