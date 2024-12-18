package fluentbit

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

func (r *FluentbitReconciler) handleDaemonSet(cr *loggingService.LoggingService) error {
	m, err := fluentbitDaemonSet(cr, r.DynamicParameters)
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
			if err = r.UpdateResource(e); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (r *FluentbitReconciler) handleService(cr *loggingService.LoggingService) error {
	m, err := fluentbitService(cr, r.DynamicParameters)
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

func (r *FluentbitReconciler) handleConfigMap(cr *loggingService.LoggingService) error {
	cm, err := fluentbitConfigMap(cr, r.DynamicParameters)
	if err != nil {
		r.Log.Error(err, "Failed creating ConfigMap manifest")
		return err
	}

	_, err = r.CreateOrUpdate(cr, cm)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("Cannot create or update config map %s", cm.Name))
		return err
	}

	return nil
}

func (r *FluentbitReconciler) deleteDaemonSet(cr *loggingService.LoggingService) error {
	e := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.FluentbitComponentName,
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

func (r *FluentbitReconciler) deleteConfigMap(cr *loggingService.LoggingService) error {
	e := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.FluentbitComponentName,
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

func (r *FluentbitReconciler) deleteService(cr *loggingService.LoggingService) error {
	e := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.FluentbitComponentName,
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

func (r *FluentbitReconciler) Equal(source *corev1.ConfigMap, target *corev1.ConfigMap) bool {
	return cmp.Equal(source.Data, target.Data) && cmp.Equal(source.BinaryData, target.BinaryData)
}

func (r *FluentbitReconciler) CreateOrUpdate(cr *loggingService.LoggingService, configMap *corev1.ConfigMap) (created bool, err error) {
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

				return false, nil
			}

			r.Log.Info("The config map is not changed")
			return false, nil
		}

		return false, err
	}

	return true, nil
}
