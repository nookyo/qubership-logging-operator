package events_reader

import (
	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *EventsReaderReconciler) handleDeployment(cr *loggingService.LoggingService) error {
	m, err := eventsReaderDeployment(cr)
	if err != nil {
		r.Log.Error(err, "Failed creating Deployment manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if errors.IsAlreadyExists(err) {
			e := &appsv1.Deployment{ObjectMeta: m.ObjectMeta}
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
			e.Spec.Selector = m.Spec.Selector
			e.Spec.Template.SetLabels(m.Spec.Template.GetLabels())
			e.Spec.Template.Spec.SecurityContext = m.Spec.Template.Spec.SecurityContext
			e.Spec.Template.Spec.Containers = m.Spec.Template.Spec.Containers
			e.Spec.Template.Spec.ServiceAccountName = m.Spec.Template.Spec.ServiceAccountName
			e.Spec.Template.Spec.NodeSelector = m.Spec.Template.Spec.NodeSelector
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

func (r *EventsReaderReconciler) handleService(cr *loggingService.LoggingService) error {
	m, err := eventsReaderService(cr)
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

func (r *EventsReaderReconciler) deleteDeployment(cr *loggingService.LoggingService) error {
	e := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.EventsReaderComponentName,
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

func (r *EventsReaderReconciler) deleteService(cr *loggingService.LoggingService) error {
	e := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.EventsReaderComponentName,
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

func (r *EventsReaderReconciler) deleteServiceAccount(cr *loggingService.LoggingService) error {
	e := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.EventsReaderComponentName,
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
