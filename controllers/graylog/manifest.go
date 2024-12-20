package graylog

import (
	"embed"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed  assets/*.yaml
var assets embed.FS

//go:embed  config/*
var configs embed.FS

func graylogServiceAccount(cr *loggingService.LoggingService) (*corev1.ServiceAccount, error) {
	sa := corev1.ServiceAccount{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.GraylogServiceAccount), util.GraylogServiceAccount, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&sa); err != nil {
		return nil, err
	}
	//Add required labels
	sa.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(sa.GetName(), sa.GetNamespace())
	sa.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Graylog.DockerImage)
	return &sa, nil
}

func graylogConfigMap(cr *loggingService.LoggingService) (*corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}
	data, err := util.DataFromDirectory(configs, util.GraylogConfigMapDirectory, cr.ToParams())
	if err != nil {
		return nil, err
	}

	//Set parameters
	configMap.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
	configMap.SetName(util.GraylogComponentName)
	configMap.SetNamespace(cr.GetNamespace())
	configMap.Data = data
	//Add required labels
	configMap.SetLabels(map[string]string{
		"name":                         util.GraylogComponentName,
		"app.kubernetes.io/name":       util.GraylogComponentName,
		"app.kubernetes.io/instance":   util.GetInstanceLabel(configMap.GetName(), configMap.GetNamespace()),
		"app.kubernetes.io/version":    util.GetTagFromImage(cr.Spec.Graylog.DockerImage),
		"app.kubernetes.io/component":  "graylog",
		"app.kubernetes.io/part-of":    "logging",
		"app.kubernetes.io/managed-by": "logging-operator",
	})
	return &configMap, nil
}

func graylogMongoUpgradeJob(cr *loggingService.LoggingService, assetPath string) (*batchv1.Job, error) {
	job := batchv1.Job{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, assetPath), assetPath, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&job); err != nil {
		return nil, err
	}
	//Add required labels
	job.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(job.GetName(), job.GetNamespace())
	job.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(job.Spec.Template.Spec.Containers[0].Image)
	job.Spec.Template.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(job.GetName(), job.GetNamespace())
	job.Spec.Template.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(job.Spec.Template.Spec.Containers[0].Image)
	return &job, nil
}

func graylogStatefulset(cr *loggingService.LoggingService) (*appsv1.StatefulSet, error) {
	statefulset := appsv1.StatefulSet{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.GraylogStatefulset), util.GraylogStatefulset, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&statefulset); err != nil {
		return nil, err
	}

	if cr.Spec.Graylog != nil {
		if cr.Spec.Graylog.Annotations != nil {
			statefulset.SetAnnotations(cr.Spec.Graylog.Annotations)
			statefulset.Spec.Template.SetAnnotations(cr.Spec.Graylog.Annotations)
		}
		//Add required labels
		statefulset.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(statefulset.GetName(), statefulset.GetNamespace())
		statefulset.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Graylog.DockerImage)
		statefulset.Spec.Template.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(statefulset.GetName(), statefulset.GetNamespace())
		statefulset.Spec.Template.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Graylog.DockerImage)
		if cr.Spec.Graylog.Labels != nil {
			for key, val := range cr.Spec.Graylog.Labels {
				statefulset.Spec.Template.Labels[key] = val
				statefulset.Labels[key] = val
			}
		}

		if len(strings.TrimSpace(cr.Spec.Graylog.PriorityClassName)) > 0 {
			statefulset.Spec.Template.Spec.PriorityClassName = cr.Spec.Graylog.PriorityClassName
		}
	}
	return &statefulset, nil
}

func graylogService(cr *loggingService.LoggingService) (*corev1.Service, error) {
	service := corev1.Service{}
	fileContent, err := util.ParseTemplate(util.MustAssetReader(assets, util.GraylogService), util.GraylogService, cr.ToParams())
	if err != nil {
		return nil, err
	}
	if err = yaml.NewYAMLOrJSONDecoder(strings.NewReader(fileContent), util.BufferSize).Decode(&service); err != nil {
		return nil, err
	}
	//Add required labels
	service.Labels["app.kubernetes.io/instance"] = util.GetInstanceLabel(service.GetName(), service.GetNamespace())
	service.Labels["app.kubernetes.io/version"] = util.GetTagFromImage(cr.Spec.Graylog.DockerImage)
	return &service, nil
}
