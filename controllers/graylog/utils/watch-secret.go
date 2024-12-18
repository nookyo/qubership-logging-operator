package utils

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretEventWatcher struct {
	Clientset *kubernetes.Clientset
	Selector  string
	Log       logr.Logger
}

func (w *SecretEventWatcher) prepareWatchChannel(namespace string) (<-chan watch.Event, error) {

	listOptions := metav1.ListOptions{
		LabelSelector: util.GraylogSecretSelector,
	}

	watchSecret, err := w.Clientset.CoreV1().Secrets(namespace).Watch(context.TODO(), listOptions)
	if err != nil {
		w.Log.Error(err, "Cannot start watch channel for a secret with selector", "selector", listOptions)
		return nil, err
	}
	watchChan := watchSecret.ResultChan()
	return watchChan, err
}

func (w *SecretEventWatcher) Watch(cr *loggingService.LoggingService, client client.Client) {

	w.Log.Info("Start watching for Secret events.", "Label Selector", w.Selector)

	watchChan, err := w.prepareWatchChannel(cr.GetNamespace())
	if err != nil {
		w.Log.Error(err, "Cannot start watch channel. Exiting...")
	}
	for {
		event, ok := <-watchChan
		if !ok {
			w.Log.Info("Secret watch channel is closed")
			break
		}

		secret, ok := event.Object.(*corev1.Secret)
		if !ok {
			w.Log.Info("Cannot convert to Secret")
			break
		}
		if secret.GetName() != cr.Spec.Graylog.GraylogSecretName {
			break
		}
		w.Log.Info("New event for Secret", "type", event.Type, "secret", secret.GetName(), "namespace", secret.GetNamespace())

		switch event.Type {
		case watch.Modified:
			err = w.updateUser(secret, cr, client)
			if err != nil {
				w.Log.Error(err, "Error while working with Secret.")
			}
			err = w.updatePassword(secret, cr, client)
			if err != nil {
				w.Log.Error(err, "Error while working with Secret.")
			}
		default:
			w.Log.Info("Received event will be ignored...", "eventType", event.Type)
		}
	}
}

func (w *SecretEventWatcher) updateUser(secret *corev1.Secret, cr *loggingService.LoggingService, client client.Client) error {
	var usr string
	if secret.Data != nil && secret.Data["user"] != nil && string(secret.Data["user"]) != "" {
		usr = string(secret.Data["user"])
	} else {
		err := errors.New("Can not find user for Graylog in the secret " + cr.Spec.Graylog.GraylogSecretName + " in the namespace " + cr.GetNamespace())
		w.Log.Error(err, "Error occurred when getting user from secret")
		return err
	}
	if cr.Spec.Graylog.User == usr {
		w.Log.Info("User did not change")
		return nil
	} else {

		cr.Spec.Graylog.User = usr

		cm, err := w.Clientset.CoreV1().ConfigMaps(cr.GetNamespace()).Get(context.TODO(), util.GraylogComponentName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		config := cm.Data[util.GraylogConfigFileName]
		if config == "" {
			w.Log.Info("Config of Graylog is empty in configmap", "configmap", util.GraylogComponentName, "namespace", cr.GetNamespace())
			return nil
		}
		var lines []string
		scanner := bufio.NewScanner(strings.NewReader(config))
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err = scanner.Err(); err != nil {
			return err
		}
		var result strings.Builder
		var changed bool
		for i := range lines {
			if strings.HasPrefix(lines[i], util.GraylogUserField) {
				lines[i] = fmt.Sprintf("%s = %s", util.GraylogUserField, usr)
				changed = true
			}
			result.WriteString(lines[i])
			result.WriteString("\r\n")
		}
		if changed {
			cm.Data[util.GraylogConfigFileName] = result.String()
			updatedCm, err := w.Clientset.CoreV1().ConfigMaps(cr.GetNamespace()).Update(context.TODO(), cm, metav1.UpdateOptions{})
			if err != nil {
				w.Log.Error(err, "Update of config map failed", "configmap", updatedCm.GetName(), "namespace", updatedCm.GetNamespace())
				return err
			}
		}
	}

	return nil
}

func (w *SecretEventWatcher) updatePassword(secret *corev1.Secret, cr *loggingService.LoggingService, client client.Client) error {
	var pwd string
	if secret.Data != nil && secret.Data["password"] != nil && string(secret.Data["password"]) != "" {
		pwd = string(secret.Data["password"])
	} else {
		err := errors.New("Can not find password for Graylog in the secret " + cr.Spec.Graylog.GraylogSecretName + " in the namespace " + cr.GetNamespace())
		w.Log.Error(err, "Error occurred when getting password from secret")
		return err
	}
	if cr.Spec.Graylog.Password == pwd {
		w.Log.Info("Password did not change")
		return nil
	} else {

		cr.Spec.Graylog.Password = pwd

		cm, err := w.Clientset.CoreV1().ConfigMaps(cr.GetNamespace()).Get(context.TODO(), util.GraylogComponentName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		config := cm.Data[util.GraylogConfigFileName]
		if config == "" {
			w.Log.Info("Config of Graylog is empty in configmap", "configmap", util.GraylogComponentName, "namespace", cr.GetNamespace())
			return nil
		}
		var lines []string
		scanner := bufio.NewScanner(strings.NewReader(config))
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err = scanner.Err(); err != nil {
			return err
		}
		var result strings.Builder
		var changed bool
		for i := range lines {
			if strings.HasPrefix(lines[i], util.GraylogPasswordField) {
				h := sha256.New()
				h.Write([]byte(pwd))
				bs := h.Sum(nil)
				lines[i] = fmt.Sprintf("%s = %s", util.GraylogPasswordField, hex.EncodeToString(bs))
				changed = true
			}
			result.WriteString(lines[i])
			result.WriteString("\r\n")
		}
		if changed {
			cm.Data[util.GraylogConfigFileName] = result.String()
			updatedCm, err := w.Clientset.CoreV1().ConfigMaps(cr.GetNamespace()).Update(context.TODO(), cm, metav1.UpdateOptions{})
			if err != nil {
				w.Log.Error(err, "Update of config map failed", "configmap", updatedCm.GetName(), "namespace", updatedCm.GetNamespace())
				return err
			}
		}
	}

	return nil
}
