package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	ResourceKey = "res"
	logger      = Logger("util")
)

func ToString(duration time.Duration) string {
	return (duration - (duration % time.Millisecond)).String()
}

func ToJSON(object interface{}) string {
	bytes, err := json.Marshal(object)
	if err != nil {
		logger.Error(err, "An unexpected problem in parsing object to json")
		return ""
	}

	return string(bytes)
}

func GetTimeNow() string {
	return time.Now().Format(time.RFC3339)
}

func GetFromResourceMap(resourceList core.ResourceList, key string) string {
	quantity := resourceList[core.ResourceName(key)]
	return quantity.String()
}

// K8sResource abstract k8s resource which can be reconciled
type K8sResource interface {
	runtime.Object
	k8smetav1.Object

	GetObjectMeta() k8smetav1.Object
}

type ComponentReconciler struct {
	Client        client.Client
	Scheme        *runtime.Scheme
	Log           logr.Logger
	StatusUpdater StatusUpdater
}

type DynamicParameters struct {
	ContainerRuntimeType string
}

func (r *ComponentReconciler) CreateResource(cr *loggingService.LoggingService, o K8sResource, setRefOptional ...bool) error {
	res := o.GetObjectKind().GroupVersionKind().Kind
	setRef := true
	if len(setRefOptional) > 0 {
		setRef = setRefOptional[0]
	}
	if setRef {
		if err := controllerutil.SetControllerReference(cr, o, r.Scheme); err != nil {
			if !(strings.Contains(err.Error(), "cluster-scoped resource must not have a namespace-scoped owner") ||
				strings.Contains(err.Error(), "cross-namespace owner references are disallowed")) {
				return err
			}
		}
	}
	if err := r.Client.Create(context.TODO(), o); err != nil {
		return err
	}
	r.Log.Info("Successful creating", ResourceKey, res)
	return nil
}

// GetResource tries to get resource inside namespace or on cluster level.
func (r *ComponentReconciler) GetResource(o K8sResource) error {
	objectKey := client.ObjectKeyFromObject(o)
	if err := r.Client.Get(context.TODO(), objectKey, o); err != nil {
		if errors.IsNotFound(err) {
			objectKey.Namespace = ""
			if err := r.Client.Get(context.TODO(), objectKey, o); err == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

func (r *ComponentReconciler) UpdateResource(o K8sResource) error {
	// Update object
	if err := r.Client.Update(context.TODO(), o); err != nil {
		return err
	}
	r.Log.Info("Successful updating", ResourceKey, o.GetObjectKind().GroupVersionKind().Kind)
	return nil
}

func (r *ComponentReconciler) DeleteResource(o K8sResource) error {
	err := r.Client.Delete(context.TODO(), o)
	if err != nil {
		return err
	}
	r.Log.Info("Successful deleting", ResourceKey, o.GetObjectKind().GroupVersionKind().Kind)
	return nil
}

// ResourceExists returns true if the given resource kind exists
// in the given api groupversion
func ResourceExists(dc discovery.DiscoveryInterface, apiGroupVersion, kind string) (bool, error) {
	_, apiLists, err := dc.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == apiGroupVersion {
			for _, r := range apiList.APIResources {
				if r.Kind == kind {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func GetTagFromImage(image string) string {
	partsOfImage := strings.Split(image, ":")
	return partsOfImage[len(partsOfImage)-1]
}

func GetInstanceLabel(name, namespace string) string {
	label := fmt.Sprintf("%s-%s", name, namespace)
	if len(label) >= 64 {
		return strings.Trim(label[:64], "-")
	}
	return strings.Trim(label, "-")
}

func GetAggregatorIds(num int) []int {
	slice := make([]int, num)
	for i := range slice {
		slice[i] = i
	}
	return slice
}
