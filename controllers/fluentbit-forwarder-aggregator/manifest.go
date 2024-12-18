package fluentbit_forwarder_aggregator

import (
	"embed"
	"fmt"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed  assets/*.yaml
var assets embed.FS

//go:embed  aggregator.configmap/conf.d/*
var aggregatorConfigs embed.FS

//go:embed  forwarder.configmap/conf.d/*
var forwarderConfigs embed.FS

func forwarderConfigMap(cr *loggingService.LoggingService, dynamicParameters util.DynamicParameters) (*corev1.ConfigMap, error) {
	cr.Spec.ContainerRuntimeType = dynamicParameters.ContainerRuntimeType

	// Get Fluent-bit forwarder config from forwarder.configmap/conf.d files
	configMapData, err := util.DataFromDirectory(forwarderConfigs, util.ForwarderFluentbitConfigMapDirectory, cr.ToParams())

	if err != nil {
		return nil, err
	}

	defaultLabels := map[string]string{
		"k8s-app":                      "fluent-bit",
		"name":                         util.ForwarderFluentbitComponentName,
		"app.kubernetes.io/component":  "fluentbit",
		"app.kubernetes.io/part-of":    "logging",
		"app.kubernetes.io/managed-by": "logging-operator",
		"app.kubernetes.io/name":       util.ForwarderFluentbitComponentName,
		"app.kubernetes.io/instance":   util.ForwarderFluentbitComponentName + "-" + cr.GetNamespace(),
		"app.kubernetes.io/version":    util.GetTagFromImage(cr.Spec.Fluentbit.DockerImage),
	}

	// Set custom input from parameters
	if cr.Spec.Fluentbit.CustomInputConf != "" {
		configMapData["input-custom.conf"] = cr.Spec.Fluentbit.CustomInputConf
	}

	// Set custom scripts from parameters
	if cr.Spec.Fluentbit.CustomLuaScriptConf != nil {
		for scriptName, script := range cr.Spec.Fluentbit.CustomLuaScriptConf {
			configMapData[scriptName] = script
		}
	}

	// Set Configmap fields
	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ForwarderFluentbitComponentName,
			Namespace: cr.GetNamespace(),
			Labels:    defaultLabels,
		},
		Data: configMapData,
	}

	return &configMap, nil
}

func forwarderDaemonSet(cr *loggingService.LoggingService, dynamicParameters util.DynamicParameters) (*appsv1.DaemonSet, error) {
	if cr.Spec.Fluentbit != nil {
		ds := appsv1.DaemonSet{}
		cr.Spec.ContainerRuntimeType = dynamicParameters.ContainerRuntimeType
		fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.ForwarderFluentbitDaemonSet), util.ForwarderFluentbitDaemonSet, cr.ToParams())
		if err != nil {
			return nil, err
		}
		if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&ds); err != nil {
			return nil, err
		}

		if cr.Spec.Fluentbit != nil {
			if cr.Spec.Fluentbit.Annotations != nil {
				ds.SetAnnotations(cr.Spec.Fluentbit.Annotations)
				ds.Spec.Template.SetAnnotations(cr.Spec.Fluentbit.Annotations)
			}
			//Add required labels
			ds.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(ds.GetName(), ds.GetNamespace())
			ds.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentbit.Aggregator.DockerImage)
			ds.Spec.Template.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(ds.GetName(), ds.GetNamespace())
			ds.Spec.Template.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentbit.Aggregator.DockerImage)
			if cr.Spec.Fluentbit.Labels != nil {
				for key, val := range cr.Spec.Fluentbit.Labels {
					ds.Spec.Template.Labels[key] = val
					ds.Labels[key] = val
				}
			}
			if cr.Spec.Fluentbit.NodeSelectorKey != "" && cr.Spec.Fluentbit.NodeSelectorValue != "" {
				ds.Spec.Template.Spec.NodeSelector = map[string]string{cr.Spec.Fluentbit.NodeSelectorKey: cr.Spec.Fluentbit.NodeSelectorValue}
			}
			if len(strings.TrimSpace(cr.Spec.Fluentbit.PriorityClassName)) > 0 {
				ds.Spec.Template.Spec.PriorityClassName = cr.Spec.Fluentbit.PriorityClassName
			}
			if cr.Spec.Fluentbit.Tolerations != nil {
				ds.Spec.Template.Spec.Tolerations = cr.Spec.Fluentbit.Tolerations
			}
		}
		if err != nil {
			return nil, err
		}
		return &ds, nil
	} else {
		return nil, fmt.Errorf("fluentbit configuration in Logging Service %s is nil in the namespace %s", cr.GetName(), cr.GetNamespace())
	}
}

func forwarderService(cr *loggingService.LoggingService, dynamicParameters util.DynamicParameters) (*corev1.Service, error) {
	service := corev1.Service{}
	cr.Spec.ContainerRuntimeType = dynamicParameters.ContainerRuntimeType
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.ForwarderFluentbitService), util.ForwarderFluentbitService, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&service); err != nil {
		return nil, err
	}
	//Add required labels
	service.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(service.GetName(), service.GetNamespace())
	service.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentbit.Aggregator.DockerImage)
	return &service, nil
}

func aggregatorConfigMap(cr *loggingService.LoggingService, dynamicParameters util.DynamicParameters) (*corev1.ConfigMap, error) {
	// Get Fluent-bit forwarder config from forwarder.configmap/conf.d files
	configMapData, err := util.DataFromDirectory(aggregatorConfigs, util.AggregatorFluentbitConfigMapDirectory, cr.ToParams())

	if err != nil {
		return nil, err
	}

	if cr.Spec.Fluentbit.Aggregator.Output != nil && cr.Spec.Fluentbit.Aggregator.Output.Loki != nil &&
		cr.Spec.Fluentbit.Aggregator.Output.Loki.Enabled && cr.Spec.Fluentbit.Aggregator.Output.Loki.LabelsMapping != "" {
		configMapData["loki-labels.json"] = cr.Spec.Fluentbit.Aggregator.Output.Loki.LabelsMapping
	}

	defaultLabels := map[string]string{
		"k8s-app":                      "fluent-bit",
		"name":                         util.AggregatorFluentbitComponentName,
		"app.kubernetes.io/component":  "fluentbit",
		"app.kubernetes.io/part-of":    "logging",
		"app.kubernetes.io/managed-by": "logging-operator",
		"app.kubernetes.io/name":       util.AggregatorFluentbitComponentName,
		"app.kubernetes.io/instance":   util.AggregatorFluentbitComponentName + "-" + cr.GetNamespace(),
		"app.kubernetes.io/version":    util.GetTagFromImage(cr.Spec.Fluentbit.DockerImage),
	}

	// Set custom filters from parameters
	if cr.Spec.Fluentbit.Aggregator.CustomFilterConf != "" {
		configMapData["filter-custom.conf"] = cr.Spec.Fluentbit.Aggregator.CustomFilterConf
	}

	// Set custom output from parameters
	if cr.Spec.Fluentbit.Aggregator.CustomOutputConf != "" {
		configMapData["output-custom.conf"] = cr.Spec.Fluentbit.Aggregator.CustomOutputConf
	}

	// Set custom scripts from parameters
	if cr.Spec.Fluentbit.Aggregator.CustomLuaScriptConf != nil {
		for scriptName, script := range cr.Spec.Fluentbit.Aggregator.CustomLuaScriptConf {
			configMapData[scriptName] = script
		}
	}

	// Set Configmap fields
	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.AggregatorFluentbitComponentName,
			Namespace: cr.GetNamespace(),
			Labels:    defaultLabels,
		},
		Data: configMapData,
	}

	return &configMap, nil
}

func aggregatorStatefulSet(cr *loggingService.LoggingService) (*appsv1.StatefulSet, error) {
	statefulSet := appsv1.StatefulSet{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.AggregatorFluentbitStatefulSet), util.AggregatorFluentbitStatefulSet, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&statefulSet); err != nil {
		return nil, err
	}

	if cr.Spec.Fluentbit.Aggregator != nil {
		if cr.Spec.Fluentbit.Aggregator.Annotations != nil {
			statefulSet.SetAnnotations(cr.Spec.Fluentbit.Aggregator.Annotations)
			statefulSet.Spec.Template.SetAnnotations(cr.Spec.Fluentbit.Aggregator.Annotations)
		}
		//Add required labels
		statefulSet.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(statefulSet.GetName(), statefulSet.GetNamespace())
		statefulSet.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentbit.Aggregator.DockerImage)
		statefulSet.Spec.Template.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(statefulSet.GetName(), statefulSet.GetNamespace())
		statefulSet.Spec.Template.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentbit.Aggregator.DockerImage)
		if cr.Spec.Fluentbit.Aggregator.Labels != nil {
			for key, val := range cr.Spec.Fluentbit.Aggregator.Labels {
				statefulSet.Spec.Template.Labels[key] = val
				statefulSet.Labels[key] = val
			}
		}
		if cr.Spec.Fluentbit.Aggregator.NodeSelectorKey != "" && cr.Spec.Fluentbit.Aggregator.NodeSelectorValue != "" {
			statefulSet.Spec.Template.Spec.NodeSelector = map[string]string{cr.Spec.Fluentbit.Aggregator.NodeSelectorKey: cr.Spec.Fluentbit.Aggregator.NodeSelectorValue}
		}
		if len(strings.TrimSpace(cr.Spec.Fluentbit.Aggregator.PriorityClassName)) > 0 {
			statefulSet.Spec.Template.Spec.PriorityClassName = cr.Spec.Fluentbit.Aggregator.PriorityClassName
		}
		if cr.Spec.Fluentbit.Aggregator.Tolerations != nil {
			statefulSet.Spec.Template.Spec.Tolerations = cr.Spec.Fluentbit.Aggregator.Tolerations
		}
	}
	return &statefulSet, nil
}

func aggregatorService(cr *loggingService.LoggingService) (*corev1.Service, error) {
	service := corev1.Service{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.AggregatorFluentbitService), util.AggregatorFluentbitService, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&service); err != nil {
		return nil, err
	}
	//Add required labels
	service.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(service.GetName(), service.GetNamespace())
	service.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Fluentbit.Aggregator.DockerImage)
	return &service, nil
}
