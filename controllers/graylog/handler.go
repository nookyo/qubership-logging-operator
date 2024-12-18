package graylog

import (
	"errors"
	"time"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *GraylogReconciler) handleServiceAccount(cr *loggingService.LoggingService) error {
	m, err := graylogServiceAccount(cr)
	if err != nil {
		r.Log.Error(err, "Failed creating ServiceAccount manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
			e := &corev1.ServiceAccount{ObjectMeta: m.ObjectMeta}
			//Set parameters
			e.SetLabels(m.GetLabels())

			if err = r.UpdateResource(e); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (r *GraylogReconciler) handleConfigMap(cr *loggingService.LoggingService) error {
	if err := r.setCredentials(cr); err != nil {
		return err
	}
	m, err := graylogConfigMap(cr)
	if err != nil {
		r.Log.Error(err, "Failed creating ConfigMap manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
			r.Log.Info("ConfigMap already exists, update it")
			if err = r.UpdateResource(m); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (r *GraylogReconciler) handleMongoUpgradeJob(cr *loggingService.LoggingService, jobName, assetPath string) error {
	m, err := graylogMongoUpgradeJob(cr, assetPath)
	if err != nil {
		r.Log.Error(err, "Failed creating Job for MongoDB upgrade manifest")
		return err
	}

	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
			r.Log.Info("Job for MongoDB upgrade already exists, skip creating")
			return nil
		} else {
			return err
		}
	}

	// Delay to allow time for the job finished successfully
	time.Sleep(util.InitialDelay)

	podManager := util.NewPodManager(r.Client, cr.GetNamespace(), r.Log)
	timeout := util.GraylogMongoUpgradeJobTimeout
	succeeded, err := podManager.WaitForJobSucceeded(jobName, timeout)
	if err != nil {
		return err
	}

	if !succeeded {
		r.StatusUpdater.UpdateStatus(util.GraylogStatus, util.Failed, false, "Job failed")
		return errors.New("mongo upgrade job failed")
	}

	// Delete pods when the job is done
	_, err = podManager.DeletePods(util.GraylogMongoUpgradeLabels)
	if err != nil {
		return err
	}

	return nil
}

func (r *GraylogReconciler) handleStatefulset(cr *loggingService.LoggingService) error {
	m, err := graylogStatefulset(cr)
	if err != nil {
		r.Log.Error(err, "Failed creating Statefulset manifest")
		return err
	}
	for i, container := range m.Spec.Template.Spec.InitContainers {
		if container.Name == "download-plugins" {
			var graylogVersionEnv corev1.EnvVar
			if r.checkGraylog5(cr) {
				graylogVersionEnv = corev1.EnvVar{
					Name:  "GRAYLOG_VERSION",
					Value: "5",
				}
			} else {
				graylogVersionEnv = corev1.EnvVar{
					Name:  "GRAYLOG_VERSION",
					Value: "4",
				}
			}
			m.Spec.Template.Spec.InitContainers[i].Env = append(m.Spec.Template.Spec.InitContainers[i].Env, graylogVersionEnv)
		}
	}

	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
			e := &appsv1.StatefulSet{ObjectMeta: m.ObjectMeta}
			if err = r.GetResource(e); err != nil {
				return err
			}

			//Set parameters
			if e.Labels == nil && m.Labels != nil {
				e.SetLabels(m.Labels)
			} else {
				for k, v := range m.Spec.Template.Labels {
					e.Labels[k] = v
				}
			}
			e.Spec.Replicas = m.Spec.Replicas
			e.Spec.Selector = m.Spec.Selector
			e.Spec.Template.SetLabels(m.Spec.Template.GetLabels())
			e.Spec.Template.Spec.SecurityContext = m.Spec.Template.Spec.SecurityContext
			e.Spec.Template.Spec.Containers = m.Spec.Template.Spec.Containers
			e.Spec.Template.Spec.InitContainers = m.Spec.Template.Spec.InitContainers
			e.Spec.Template.Spec.Volumes = m.Spec.Template.Spec.Volumes
			e.Spec.Template.Spec.ServiceAccountName = m.Spec.Template.Spec.ServiceAccountName
			e.Spec.Template.Spec.NodeSelector = m.Spec.Template.Spec.NodeSelector

			if err = r.UpdateResource(e); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Delay to allow time for the deploy to be updated
	time.Sleep(util.InitialDelay)

	// Wait for Graylog running
	podManager := util.NewPodManager(r.Client, cr.GetNamespace(), r.Log)
	timeout := util.GraylogStartupTimeout
	if cr.Spec.Graylog.StartupTimeout != 0 {
		timeout = time.Duration(cr.Spec.Graylog.StartupTimeout) * time.Minute
	}
	started, err := podManager.WaitForStatefulsetUpdated(util.GraylogStatefulsetName, timeout)
	if err != nil {
		return err
	}

	if !started {
		r.StatusUpdater.UpdateStatus(util.GraylogStatus, util.Failed, false, "Graylog is not started")
		return errors.New("graylog is not started")
	}
	return nil
}

func (r *GraylogReconciler) handleService(cr *loggingService.LoggingService) error {
	m, err := graylogService(cr)
	if err != nil {
		r.Log.Error(err, "Failed creating Service manifest")
		return err
	}
	if err = r.CreateResource(cr, m); err != nil {
		if api_errors.IsAlreadyExists(err) {
			e := &corev1.Service{ObjectMeta: m.ObjectMeta}
			//Set parameters
			e.SetLabels(m.GetLabels())
			e.Spec.Ports = m.Spec.Ports
			e.Spec.Selector = m.Spec.Selector

			if err = r.UpdateResource(e); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}

	dnsName := util.GraylogComponentName + "." + cr.GetNamespace() + ".svc"
	if err = util.WaitForHostActive(dnsName, 9000, time.Minute); err != nil {
		return err
	}

	return nil
}

func (r *GraylogReconciler) deletePVC(name string, cr *loggingService.LoggingService) error {
	e := &corev1.PersistentVolumeClaim{
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

func (r *GraylogReconciler) deleteDeployment(cr *loggingService.LoggingService) error {
	e := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GraylogDeploymentName,
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
	podManager := util.NewPodManager(r.Client, cr.GetNamespace(), r.Log)
	_, err := podManager.DeletePods(util.GraylogLabels)
	if err != nil {
		return err
	}
	return nil
}

func (r *GraylogReconciler) deleteStatefulset(cr *loggingService.LoggingService) error {
	e := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GraylogStatefulsetName,
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

func (r *GraylogReconciler) deleteService(cr *loggingService.LoggingService) error {
	e := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GraylogComponentName,
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

func (r *GraylogReconciler) deleteConfigMap(cr *loggingService.LoggingService) error {
	e := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GraylogComponentName,
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

func (r *GraylogReconciler) deleteServiceAccount(cr *loggingService.LoggingService) error {
	e := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GraylogServiceAccountName,
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

func (r *GraylogReconciler) deleteJob(cr *loggingService.LoggingService, jobName string) error {
	e := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
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

func (r *GraylogReconciler) scaleDownStatefulset(cr *loggingService.LoggingService) error {
	e := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GraylogStatefulsetName,
			Namespace: cr.GetNamespace(),
		},
	}
	if err := r.GetResource(e); err != nil {
		if api_errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	replicas := new(int32)
	*replicas = 0
	e.Spec.Replicas = replicas

	if err := r.UpdateResource(e); err != nil {
		return err
	}
	// Delay to allow time for the statefulset to be updated
	time.Sleep(util.InitialDelay)

	// Wait for Graylog running
	podManager := util.NewPodManager(r.Client, cr.GetNamespace(), r.Log)
	timeout := util.GraylogStartupTimeout
	if cr.Spec.Graylog.StartupTimeout != 0 {
		timeout = time.Duration(cr.Spec.Graylog.StartupTimeout) * time.Minute
	}
	updated, err := podManager.WaitForStatefulsetUpdated(util.GraylogStatefulsetName, timeout)
	if err != nil {
		return err
	}

	if !updated {
		r.StatusUpdater.UpdateStatus(util.GraylogStatus, util.Failed, false, "Graylog has not scaled down")
		return errors.New("graylog has not scaled down")
	}
	return nil
}
