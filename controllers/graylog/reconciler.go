package graylog

import (
	"context"
	"errors"
	"regexp"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-logging-operator/controllers/graylog/utils"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GraylogReconciler struct {
	*util.ComponentReconciler
}

func NewGraylogReconciler(client client.Client, scheme *runtime.Scheme, updater util.StatusUpdater) GraylogReconciler {
	return GraylogReconciler{
		ComponentReconciler: &util.ComponentReconciler{
			Client:        client,
			Scheme:        scheme,
			Log:           util.Logger("graylog"),
			StatusUpdater: updater,
		},
	}
}

// Run reconciles Graylog custom resource.
// Creates new Statefulset, Service, ServiceAccount, ConfigMap if its don't exist.
func (r *GraylogReconciler) Run(ctx context.Context, cr *loggingService.LoggingService, clientSet kubernetes.Interface) error {
	if !r.StatusUpdater.IsStatusFailed(util.GraylogStatus) {
		r.StatusUpdater.UpdateStatus(util.GraylogStatus, util.InProgress, false, "Start reconcile of Graylog")
	}
	r.Log.Info("Start Graylog reconciliation")

	if cr.Spec.Graylog != nil && cr.Spec.Graylog.IsInstall() {
		connector, err := utils.CreateConnector(ctx, cr, configs, clientSet)
		if err != nil {
			return err
		}

		if err = r.handleServiceAccount(cr); err != nil {
			return err
		}
		if err = r.handleConfigMap(cr); err != nil {
			return err
		}
		if err = r.deleteDeployment(cr); err != nil {
			r.Log.Error(err, "Can not delete Deployment")
		}
		if cr.Spec.Graylog.MongoDBUpgrade != nil && r.checkGraylog5(cr) {
			if err = r.mongoUpgrade(cr); err != nil {
				r.Log.Error(err, "MongoDB upgrade failed. Try to continue the reconciliation anyway")
			}
		} else {
			if err = r.deleteUpgradeJobs(cr); err != nil {
				r.Log.Error(err, "Can not delete MongoDB upgrade jobs")
			}
		}
		if err = r.handleStatefulset(cr); err != nil {
			return err
		}
		if err = r.handleService(cr); err != nil {
			return err
		}

		if cr.Spec.Graylog.ContentDeployPolicy != "skip" {
			if err = r.configureGraylog(ctx, connector, cr, clientSet); err != nil {
				return err
			}
		}
		if err = r.watchSecret(cr); err != nil {
			r.Log.Error(err, "Error occurred while starting watch Secret")
		}

	} else {
		r.Log.Info("Uninstalling component if exists")
		r.uninstall(cr)
	}
	r.Log.Info("Component reconciled")
	r.StatusUpdater.RemoveStatus(util.GraylogStatus)
	return nil
}

// uninstall deletes all resources related to the component
func (r *GraylogReconciler) uninstall(cr *loggingService.LoggingService) {
	if err := r.deletePVC(util.GraylogClaimName, cr); err != nil {
		r.Log.Error(err, "Can not delete graylog PVC")
	}
	if err := r.deletePVC(util.MongoClaimName, cr); err != nil {
		r.Log.Error(err, "Can not delete mongo PVC")
	}
	if err := r.deleteStatefulset(cr); err != nil {
		r.Log.Error(err, "Can not delete Statefulset")
	}
	if err := r.deleteService(cr); err != nil {
		r.Log.Error(err, "Can not delete Service")
	}
	if err := r.deleteConfigMap(cr); err != nil {
		r.Log.Error(err, "Can not delete ConfigMap")
	}
	if err := r.deleteServiceAccount(cr); err != nil {
		r.Log.Error(err, "Can not delete ServiceAccount")
	}
	if err := r.deleteUpgradeJobs(cr); err != nil {
		r.Log.Error(err, "Can not delete MongoDB upgrade jobs")
	}
}

func (r *GraylogReconciler) configureGraylog(ctx context.Context, connector *utils.GraylogConnector, cr *loggingService.LoggingService, clientSet kubernetes.Interface) error {
	if cr.Spec.Graylog.AuthProxy.Install {
		if err := connector.ManageAuthHeaderConfig(cr); err != nil {
			return err
		}
	}

	if err := connector.ManageGrokPatterns(cr); err != nil {
		return err
	}

	if err := connector.ManageIndexSets(cr); err != nil {
		return err
	}

	if err := connector.ManageInputs(cr); err != nil {
		return err
	}

	if err := connector.ManageExtractors(cr, r.checkGraylog5(cr)); err != nil {
		return err
	}

	if err := connector.ManageStreams(cr); err != nil {
		return err
	}

	if err := connector.ManageProcessingRules(cr); err != nil {
		return err
	}

	if err := connector.ManagePipelines(cr); err != nil {
		return err
	}

	if err := connector.ManageDashboards(cr); err != nil {
		return err
	}

	if cr.Spec.Graylog.ContentPackPaths != "" {
		if err := connector.ManageContentPacks(cr); err != nil {
			return err
		}

		if err := connector.ManageOpensearchConfigs(cr); err != nil {
			return err
		}
	}

	if cr.Spec.Graylog.ContentPacks != nil {
		if err := connector.ManageContentPackTLS(ctx, cr, clientSet); err != nil {
			return err
		}

		if err := connector.ManageOpensearchConfigs(cr); err != nil {
			return err
		}
	}

	if err := connector.ManageArchivesDirectory(cr); err != nil {
		return err
	}

	if err := connector.ManageSavedSearches(cr); err != nil {
		return err
	}

	if err := connector.ManageUserAccounts(cr, r.checkGraylog5(cr)); err != nil {
		return err
	}

	return nil
}

func (r *GraylogReconciler) setCredentials(cr *loggingService.LoggingService) error {
	secret := &corev1.Secret{}
	if err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name: cr.Spec.Graylog.GraylogSecretName, Namespace: cr.GetNamespace(),
	}, secret); err != nil {
		return err
	}

	var usr string
	if secret.Data != nil && secret.Data["user"] != nil && len(secret.Data["user"]) > 0 {
		usr = string(secret.Data["user"])
	} else {
		err := errors.New("can not find user for Graylog in the secret " + cr.Spec.Graylog.GraylogSecretName + " in the namespace " + cr.GetNamespace())
		return err
	}
	cr.Spec.Graylog.User = usr

	var pwd string
	if secret.Data != nil && secret.Data["password"] != nil && len(secret.Data["password"]) > 0 {
		pwd = string(secret.Data["password"])
	} else {
		err := errors.New("can not find password for Graylog in the secret " + cr.Spec.Graylog.GraylogSecretName + " in the namespace " + cr.GetNamespace())
		return err
	}
	cr.Spec.Graylog.Password = pwd
	return nil
}

func (r *GraylogReconciler) watchSecret(cr *loggingService.LoggingService) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	w := utils.SecretEventWatcher{
		Log:       util.Logger("watch-secret"),
		Clientset: k8sClient,
	}
	go w.Watch(cr, r.Client)
	return nil
}

func (r *GraylogReconciler) mongoUpgrade(cr *loggingService.LoggingService) error {
	// Scale down the Graylog statefulset before starting upgrade jobs to avoid conflicts between MongoDB instances
	if err := r.scaleDownStatefulset(cr); err != nil {
		return err
	}
	// Run jobs in particular order
	for _, jobName := range util.GraylogMongoUpgradeOrderedJobs {
		assetPath := util.GraylogMongoUpgradeAssets[jobName]
		if err := r.handleMongoUpgradeJob(cr, jobName, assetPath); err != nil {
			return err
		}
	}
	return nil
}

func (r *GraylogReconciler) deleteUpgradeJobs(cr *loggingService.LoggingService) error {
	for jobName := range util.GraylogMongoUpgradeAssets {
		if err := r.deleteJob(cr, jobName); err != nil {
			return err
		}
	}
	return nil
}

func (r *GraylogReconciler) checkGraylog5(cr *loggingService.LoggingService) bool {
	// Parsing of the major version of the Graylog image
	re := regexp.MustCompile(`([0-9]+)\.([0-9]+)\.([0-9]+)`)
	majorVersion := strings.Split(re.FindString(cr.Spec.Graylog.DockerImage), ".")[0]
	return majorVersion == "5"
}
