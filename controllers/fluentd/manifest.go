package fluentd

import (
	"embed"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed  assets/*.yaml
var assets embed.FS

//go:embed  fluentd.configmap/conf.d/*
var configs embed.FS

func fluentdConfigMap(cr *loggingService.LoggingService, dynamicParameters util.DynamicParameters) (*corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}
	cr.Spec.ContainerRuntimeType = dynamicParameters.ContainerRuntimeType
	data, err := util.DataFromDirectory(configs, util.FluentdConfigMapDirectory, cr.ToParams())
	if err != nil {
		return nil, err
	}

	data["input-custom.conf"] = cr.Spec.Fluentd.CustomInputConf
	data["filter-custom.conf"] = cr.Spec.Fluentd.CustomFilterConf
	data["output-custom.conf"] = cr.Spec.Fluentd.CustomOutputConf

	//Set parameters
	configMap.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
	configMap.SetName(util.FluentdComponentName)
	configMap.SetNamespace(cr.GetNamespace())
	configMap.Data = data

	//Add required labels
	configMap.SetLabels(map[string]string{
		"name":                         util.FluentdComponentName,
		"app.kubernetes.io/name":       util.FluentdComponentName,
		"app.kubernetes.io/instance":   util.GetInstanceLabel(configMap.GetName(), configMap.GetNamespace()),
		"app.kubernetes.io/version":    util.GetTagFromImage(cr.Spec.Fluentd.DockerImage),
		"app.kubernetes.io/component":  "fluentd",
		"app.kubernetes.io/part-of":    "logging",
		"app.kubernetes.io/managed-by": "logging-operator",
	})
	return &configMap, nil
}

func fluentdDaemonSet(cr *loggingService.LoggingService, dynamicParameters util.DynamicParameters) (*appsv1.DaemonSet, error) {
	daemonSet := appsv1.DaemonSet{}
	cr.Spec.ContainerRuntimeType = dynamicParameters.ContainerRuntimeType
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.FluentdDaemonSet), util.FluentdDaemonSet, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&daemonSet); err != nil {
		return nil, err
	}

	if cr.Spec.Fluentd != nil {
		if cr.Spec.Fluentd.Annotations != nil {
			daemonSet.SetAnnotations(cr.Spec.Fluentd.Annotations)
			daemonSet.Spec.Template.SetAnnotations(cr.Spec.Fluentd.Annotations)
		}
		//Add required labels
		daemonSet.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(daemonSet.GetName(), daemonSet.GetNamespace())
		daemonSet.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentd.DockerImage)
		daemonSet.Spec.Template.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(daemonSet.GetName(), daemonSet.GetNamespace())
		daemonSet.Spec.Template.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentd.DockerImage)

		if cr.Spec.Fluentd.Labels != nil {
			for key, val := range cr.Spec.Fluentd.Labels {
				daemonSet.Spec.Template.Labels[key] = val
				daemonSet.Labels[key] = val
			}
		}
		if cr.Spec.Fluentd.NodeSelectorKey != "" && cr.Spec.Fluentd.NodeSelectorValue != "" {
			daemonSet.Spec.Template.Spec.NodeSelector = map[string]string{cr.Spec.Fluentd.NodeSelectorKey: cr.Spec.Fluentd.NodeSelectorValue}
		}
		if len(strings.TrimSpace(cr.Spec.Fluentd.PriorityClassName)) > 0 {
			daemonSet.Spec.Template.Spec.PriorityClassName = cr.Spec.Fluentd.PriorityClassName
		}
		if cr.Spec.Fluentd.Tolerations != nil {
			daemonSet.Spec.Template.Spec.Tolerations = cr.Spec.Fluentd.Tolerations
		}
		if cr.Spec.Fluentd.Affinity != nil {
			daemonSet.Spec.Template.Spec.Affinity = cr.Spec.Fluentd.Affinity
		}
		if cr.Spec.Fluentd.AdditionalVolumes != nil {
			daemonSet.Spec.Template.Spec.Volumes = append(daemonSet.Spec.Template.Spec.Volumes, cr.Spec.Fluentd.AdditionalVolumes...)
		}
		if cr.Spec.Fluentd.AdditionalVolumeMounts != nil {
			for it := range daemonSet.Spec.Template.Spec.Containers {
				c := &daemonSet.Spec.Template.Spec.Containers[it]
				if c.Name == "logging-fluentd" {
					c.VolumeMounts = append(c.VolumeMounts, cr.Spec.Fluentd.AdditionalVolumeMounts...)
				}
			}
		}
	}
	return &daemonSet, nil
}

func fluentdService(cr *loggingService.LoggingService, dynamicParameters util.DynamicParameters) (*corev1.Service, error) {
	service := corev1.Service{}
	cr.Spec.ContainerRuntimeType = dynamicParameters.ContainerRuntimeType
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.FluentdServiceTemplate), util.FluentdServiceTemplate, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&service); err != nil {
		return nil, err
	}
	//Add required labels
	service.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(service.GetName(), service.GetNamespace())
	service.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentd.DockerImage)
	return &service, nil
}
