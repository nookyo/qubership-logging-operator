/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-logging-operator/controllers"
	"github.com/Netcracker/qubership-logging-operator/controllers/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"runtime"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	//+kubebuilder:scaffold:imports
)

var (
	scheme            = apiruntime.NewScheme()
	logger            = utils.Logger("cmd")
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(loggingService.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func printVersion() {
	logger.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	logger.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	var pprofAddr string
	var pprofEnabled bool

	flag.BoolVar(&pprofEnabled, "pprof-enable", true, "Enable pprof.")
	flag.StringVar(&pprofAddr, "pprof-address", ":9180", "The pprof address.")
	printVersion()
	logf.SetLogger(logger)
	namespace, found := os.LookupEnv("WATCH_NAMESPACE")
	if !found {
		namespace = "logging"
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		},
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
		},
	})
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.LoggingServiceReconciler{
		Client:                   mgr.GetClient(),
		Scheme:                   mgr.GetScheme(),
		Log:                      utils.Logger("controller-loggingservice"),
		Config:                   mgr.GetConfig(),
		TimeoutOnFailedReconcile: controllers.InitialTimeoutOnFailedReconcile,
		DynamicParameters:        utils.DynamicParameters{ContainerRuntimeType: ""},
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create controller", "controller", "LoggingService")
		os.Exit(1)
	}

	skipMetricsService, found := os.LookupEnv("SKIP_METRICS_SERVICE")
	if !(found && skipMetricsService == "true") {
		// Add to the below struct any other metrics ports you want to expose.
		servicePorts := []v1.ServicePort{
			{Port: metricsPort, Name: "http-metrics", Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		}
		serviceName := "logging-service-operator-metrics"
		label := map[string]string{
			"name":                         serviceName,
			"app.kubernetes.io/name":       serviceName,
			"app.kubernetes.io/instance":   utils.GetInstanceLabel(serviceName, namespace),
			"app.kubernetes.io/component":  "logging-operator",
			"app.kubernetes.io/part-of":    "logging",
			"app.kubernetes.io/managed-by": "logging-operator",
		}
		// Create Service object to expose the metrics port(s).
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "logging-service-operator-metrics",
				Namespace: namespace,
				Labels:    label,
			},
			Spec: v1.ServiceSpec{
				Ports:    servicePorts,
				Selector: label,
			},
		}
		if err != nil {
			logger.Info("Could not create metrics Service", "error", err.Error())
		}

		// CreatePodMonitors will automatically create the prometheus-operator PodMonitor resources
		// necessary to configure Prometheus to scrape metrics from this operator.
		services := []*v1.Service{service}
		scrapeInterval, _ := os.LookupEnv("SCRAPE_INTERVAL")
		scrapeTimeout, _ := os.LookupEnv("SCRAPE_TIMEOUT")
		err = utils.CreatePodMonitors(mgr.GetConfig(), namespace, services, scrapeInterval, scrapeTimeout)
		if err != nil {
			logger.Info("Could not create PodMonitor object", "error", err.Error())
		}
	} else {
		logger.Info("Skip a creating Service and PodMonitor which scrape metrics from this operator")
	}

	//+kubebuilder:scaffold:builder

	if pprofEnabled {
		go func() {
			logger.Info("start pprof HTTP server", "addr:", pprofAddr)
			err = http.ListenAndServe(pprofAddr, nil)
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error(err, "failed to start pprof HTTP server")
			}
			logger.Info("shutdown pprof HTTP server", "addr:", pprofAddr)
		}()
	}

	logger.Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "problem running manager")
		os.Exit(1)
	}
}
