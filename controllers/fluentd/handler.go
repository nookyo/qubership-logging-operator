package fluentd

import (
	"fmt"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *FluentdReconciler) handleConfigMap(cr *loggingService.LoggingService) error {
	m, err := fluentdConfigMap(cr, r.DynamicParameters)
	if err != nil {
		r.Log.Error(err, "Failed creating ConfigMap manifest")
		return err
	}

	_, err = r.updateConfigMap(cr, m)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("Cannot create or update config map %s", m.Name))
		return err
	}

	return nil
}

func (r *FluentdReconciler) handleDaemonSet(cr *loggingService.LoggingService) error {
	m, err := fluentdDaemonSet(cr, r.DynamicParameters)
	if err != nil {
		r.Log.Error(err, "Failed creating DaemonSet manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if errors.IsAlreadyExists(err) {
			e := &appsv1.DaemonSet{ObjectMeta: m.ObjectMeta}
			if err = r.GetResource(e); err != nil {
				return err
			}

			//Set parameters
			if e.Labels == nil && m.Labels != nil {
				e.SetLabels(m.Labels)
			} else {
				for k, v := range m.Labels {
					e.Labels[k] = v
				}
			}
			e.Spec.Template.SetLabels(m.Spec.Template.GetLabels())
			e.Spec.Template.Spec.Containers = m.Spec.Template.Spec.Containers
			e.Spec.Template.Spec.ServiceAccountName = m.Spec.Template.Spec.ServiceAccountName
			e.Spec.Template.Spec.NodeSelector = m.Spec.Template.Spec.NodeSelector
			e.Spec.Template.Spec.Volumes = m.Spec.Template.Spec.Volumes
			e.Spec.Template.Spec.Tolerations = m.Spec.Template.Spec.Tolerations
			e.Spec.Template.Spec.Affinity = m.Spec.Template.Spec.Affinity
			if err = r.UpdateResource(e); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (r *FluentdReconciler) handleService(cr *loggingService.LoggingService) error {
	m, err := fluentdService(cr, r.DynamicParameters)
	if err != nil {
		r.Log.Error(err, "Failed creating Service manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if errors.IsAlreadyExists(err) {
			e := &corev1.Service{ObjectMeta: m.ObjectMeta}
			if err = r.GetResource(e); err != nil {
				return err
			}

			//Set parameters
			e.SetLabels(m.GetLabels())
			e.Spec.Ports = m.Spec.Ports
			e.Spec.Selector = m.Spec.Selector

			if err = r.UpdateResource(e); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (r *FluentdReconciler) deleteDaemonSet(cr *loggingService.LoggingService) error {
	e := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.FluentdComponentName,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.DeleteResource(e); err != nil {
		return err
	}
	return nil
}

func (r *FluentdReconciler) deleteConfigMap(cr *loggingService.LoggingService) error {
	e := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.FluentdComponentName,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.DeleteResource(e); err != nil {
		return err
	}
	return nil
}

func (r *FluentdReconciler) deleteService(cr *loggingService.LoggingService) error {
	e := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.FluentdComponentName,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.DeleteResource(e); err != nil {
		return err
	}
	return nil
}

func (r *FluentdReconciler) Equal(source *corev1.ConfigMap, target *corev1.ConfigMap) bool {
	return cmp.Equal(source.Data, target.Data) &&
		cmp.Equal(source.BinaryData, target.BinaryData) &&
		cmp.Equal(source.GetLabels(), target.GetLabels())
}

func (r *FluentdReconciler) updateConfigMap(cr *loggingService.LoggingService, configMap *corev1.ConfigMap) (updated bool, err error) {
	if err = r.CreateResource(cr, configMap); err != nil {
		if errors.IsAlreadyExists(err) {
			existedConfigMap := &corev1.ConfigMap{ObjectMeta: configMap.ObjectMeta}
			if err = r.GetResource(existedConfigMap); err != nil {
				return false, err
			}

			if !r.Equal(existedConfigMap, configMap) {
				if err = r.UpdateResource(configMap); err != nil {
					return false, err
				}

				return true, nil
			}

			r.Log.Info("The config map is not changed")
			return false, nil
		}

		return false, err
	}

	return true, nil
}
