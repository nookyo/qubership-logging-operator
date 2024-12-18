package fluentbit_forwarder_aggregator

import (
	"errors"
	"fmt"
	"time"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *HAFluentReconciler) handleForwarderConfigMap(cr *loggingService.LoggingService) error {
	m, err := forwarderConfigMap(cr, r.DynamicParameters)
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

func (r *HAFluentReconciler) handleForwarderDaemonSet(cr *loggingService.LoggingService) error {
	m, err := forwarderDaemonSet(cr, r.DynamicParameters)
	if err != nil {
		r.Log.Error(err, "Failed creating DaemonSet manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
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

func (r *HAFluentReconciler) handleForwarderService(cr *loggingService.LoggingService) error {
	m, err := forwarderService(cr, r.DynamicParameters)
	if err != nil {
		r.Log.Error(err, "Failed creating Service manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
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

func (r *HAFluentReconciler) Equal(source *corev1.ConfigMap, target *corev1.ConfigMap) bool {
	return cmp.Equal(source.Data, target.Data) &&
		cmp.Equal(source.BinaryData, target.BinaryData) &&
		cmp.Equal(source.GetLabels(), target.GetLabels())
}

func (r *HAFluentReconciler) CreateOrUpdate(cr *loggingService.LoggingService, configMap *corev1.ConfigMap) (created bool, updated bool, err error) {
	if err = r.CreateResource(cr, configMap); err != nil {
		if api_errors.IsAlreadyExists(err) {
			existedConfigMap := &corev1.ConfigMap{ObjectMeta: configMap.ObjectMeta}
			if err = r.GetResource(existedConfigMap); err != nil {
				return false, false, err
			}

			if !r.Equal(existedConfigMap, configMap) {
				if err = r.UpdateResource(configMap); err != nil {
					return false, false, err
				}

				return false, true, nil
			}

			r.Log.Info("The config map is not changed")
			return false, false, nil
		}

		return false, false, err
	}

	return true, false, nil
}

func (r *HAFluentReconciler) handleAggregatorConfigMap(cr *loggingService.LoggingService) error {
	m, err := aggregatorConfigMap(cr, r.DynamicParameters)
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

func (r *HAFluentReconciler) handleAggregatorStatefulSet(cr *loggingService.LoggingService) error {
	ss, err := aggregatorStatefulSet(cr)
	if err != nil {
		r.Log.Error(err, "Failed creating Stateful Set manifest")
		return err
	}

	if err = r.CreateResource(cr, ss); err != nil {
		if api_errors.IsAlreadyExists(err) {
			e := &appsv1.StatefulSet{ObjectMeta: ss.ObjectMeta}
			if err = r.GetResource(e); err != nil {
				return err
			}

			//Set parameters
			if e.Labels == nil && ss.Labels != nil {
				e.SetLabels(ss.Labels)
			} else {
				for k, v := range ss.Labels {
					e.Labels[k] = v
				}
			}
			e.Spec.Template.SetLabels(ss.Spec.Template.GetLabels())
			e.Spec.Template.Spec.Containers = ss.Spec.Template.Spec.Containers
			e.Spec.Template.Spec.ServiceAccountName = ss.Spec.Template.Spec.ServiceAccountName
			e.Spec.Template.Spec.NodeSelector = ss.Spec.Template.Spec.NodeSelector
			e.Spec.Template.Spec.Volumes = ss.Spec.Template.Spec.Volumes
			e.Spec.Template.Spec.Tolerations = ss.Spec.Template.Spec.Tolerations
			if err = r.UpdateResource(e); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// Delay to allow time for the deploy to be updated
	time.Sleep(util.InitialDelay)

	// Wait for Aggregator running
	podManager := util.NewPodManager(r.Client, cr.GetNamespace(), r.Log)
	timeout := util.FluentbitAggregatorPendingTimeout
	if cr.Spec.Fluentbit.Aggregator.StartupTimeout != 0 {
		timeout = time.Duration(cr.Spec.Fluentbit.Aggregator.StartupTimeout) * time.Minute
	}
	started, err := podManager.WaitForStatefulsetUpdated(util.AggregatorFluentbitComponentName, timeout)
	if err != nil {
		return err
	}

	if !started {
		r.StatusUpdater.UpdateStatus(util.HAFluentStatus, util.Failed, false, "Fluent bit aggregator is not started")
		return errors.New("fluent bit aggregator is not started")
	}
	return nil
}

func (r *HAFluentReconciler) handleAggregatorService(cr *loggingService.LoggingService) error {
	m, err := aggregatorService(cr)
	if err != nil {
		r.Log.Error(err, "Failed creating Service manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
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

func (r *HAFluentReconciler) deleteDaemonSet(cr *loggingService.LoggingService, name string) error {
	e := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if api_errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.DeleteResource(e); err != nil {
		return err
	}
	return nil
}

func (r *HAFluentReconciler) deleteStatefulSet(cr *loggingService.LoggingService, name string) error {
	e := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if api_errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.DeleteResource(e); err != nil {
		return err
	}
	return nil
}

func (r *HAFluentReconciler) deleteConfigMap(cr *loggingService.LoggingService, name string) error {
	e := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if api_errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.DeleteResource(e); err != nil {
		return err
	}
	return nil
}

func (r *HAFluentReconciler) deleteService(cr *loggingService.LoggingService, name string) error {
	e := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if api_errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.DeleteResource(e); err != nil {
		return err
	}
	return nil
}

func (r *HAFluentReconciler) updateConfigMap(cr *loggingService.LoggingService, configMap *corev1.ConfigMap) (updated bool, err error) {
	if err = r.CreateResource(cr, configMap); err != nil {
		if api_errors.IsAlreadyExists(err) {
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
