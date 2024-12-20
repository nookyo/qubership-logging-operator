package utils

import (
	"context"
	"fmt"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monclientv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

var (
	ErrPodMonitorNotPresent = fmt.Errorf("no PodMonitor registered with the API")
	SelectorLabels          = map[string]string{"name": "logging-service-operator"}
)

// CreatePodMonitors creates PodMonitors objects based on an array of Pod objects.
// If CR PodMonitor is not registered in the Cluster it will not attempt at creating resources.
func CreatePodMonitors(config *rest.Config, ns string, services []*v1.Service, sInterval string, sTimeout string) error {
	// check if we can even create PodMonitors
	exists, err := hasPodMonitor(config)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPodMonitorNotPresent
	}

	mclient := monclientv1.NewForConfigOrDie(config)

	for _, s := range services {
		if s == nil {
			continue
		}
		sm := GeneratePodMonitor(s, sInterval, sTimeout)
		_, err = mclient.PodMonitors(ns).Create(context.TODO(), sm, metav1.CreateOptions{})
		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		}
	}
	return nil
}

// GeneratePodMonitor generates a prometheus-operator PodMonitor object
// based on the passed Service object.
func GeneratePodMonitor(s *v1.Service, sInterval string, sTimeout string) *promv1.PodMonitor {
	endpoints := populateEndpointsFromServicePorts(s, sInterval, sTimeout)

	return &promv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ObjectMeta.Name,
			Namespace: s.ObjectMeta.Namespace,
			Labels: map[string]string{
				"name":                         s.ObjectMeta.Name,
				"app.kubernetes.io/name":       s.ObjectMeta.Name,
				"app.kubernetes.io/instance":   GetInstanceLabel(s.ObjectMeta.Name, s.ObjectMeta.Namespace),
				"app.kubernetes.io/component":  "monitoring",
				"app.kubernetes.io/part-of":    "logging",
				"app.kubernetes.io/managed-by": "logging-operator",
			},
		},
		Spec: promv1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: SelectorLabels,
			},
			PodMetricsEndpoints: endpoints,
		},
	}
}

func populateEndpointsFromServicePorts(s *v1.Service, sInterval string, sTimeout string) []promv1.PodMetricsEndpoint {
	var endpoints []promv1.PodMetricsEndpoint
	for _, port := range s.Spec.Ports {
		endpoints = append(endpoints, promv1.PodMetricsEndpoint{Port: port.Name, Interval: promv1.Duration(sInterval), ScrapeTimeout: promv1.Duration(sTimeout)})
	}
	return endpoints
}

// hasPodMonitor checks if PodMonitor is registered in the cluster.
func hasPodMonitor(config *rest.Config) (bool, error) {
	dc := discovery.NewDiscoveryClientForConfigOrDie(config)
	apiVersion := "monitoring.coreos.com/v1"
	kind := "PodMonitor"

	return ResourceExists(dc, apiVersion, kind)
}
