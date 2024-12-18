package events_reader

import (
	"embed"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed  assets/*.yaml
var assets embed.FS

func eventsReaderDeployment(cr *loggingService.LoggingService) (*appsv1.Deployment, error) {
	deployment := appsv1.Deployment{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.EventsReaderDeployment), util.EventsReaderDeployment, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&deployment); err != nil {
		return nil, err
	}

	// Set required labels
	deployment.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(deployment.GetName(), deployment.GetNamespace())
	deployment.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.CloudEventsReader.DockerImage)
	deployment.Spec.Template.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(deployment.GetName(), deployment.GetNamespace())
	deployment.Spec.Template.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.CloudEventsReader.DockerImage)

	if cr.Spec.CloudEventsReader != nil {
		if cr.Spec.CloudEventsReader.Annotations != nil {
			deployment.SetAnnotations(cr.Spec.CloudEventsReader.Annotations)
			deployment.Spec.Template.SetAnnotations(cr.Spec.CloudEventsReader.Annotations)
		}
		if cr.Spec.CloudEventsReader.Labels != nil {
			for key, val := range cr.Spec.CloudEventsReader.Labels {
				deployment.Spec.Template.Labels[key] = val
				deployment.Labels[key] = val
			}
		}

		if len(strings.TrimSpace(cr.Spec.CloudEventsReader.PriorityClassName)) > 0 {
			deployment.Spec.Template.Spec.PriorityClassName = cr.Spec.CloudEventsReader.PriorityClassName
		}
	}

	return &deployment, nil
}

func eventsReaderService(cr *loggingService.LoggingService) (*corev1.Service, error) {
	service := corev1.Service{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.EventsReaderService), util.EventsReaderService, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&service); err != nil {
		return nil, err
	}

	// Set required labels
	service.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(service.GetName(), service.GetNamespace())
	service.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.CloudEventsReader.DockerImage)

	return &service, nil
}
