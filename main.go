/*
Copyright 2023 mgrote.

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
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
	"github.com/mgrote/personal-iot/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(personaliotv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

//func getConfig(configFile string) (*personaliotv1alpha1.MQTTControllerConfig, error) {
//	setupLog.WithValues("config", configFile).Info("found config")
//	if configFile == "" {
//		return nil, errors.New("expected config file parameter")
//	}
//
//	mqttControllerConfig := &personaliotv1alpha1.MQTTControllerConfig{}
//	options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
//
//	cfgFile := ctrl.ConfigFile().OfKind(mqttControllerConfig).AtPath(configFile)
//	if err := cfgFile.InjectScheme(scheme); err != nil {
//		return nil, errors.Wrap(err, "unable to load config file")
//	}
//	if _, err := cfgFile.Complete(); err != nil {
//		return nil, errors.Wrap(err, "unable to load the config file")
//	}
//
//	return mqttControllerConfig, nil
//
//}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var configFile string
	flag.StringVar(&configFile, "config", "./config/manager/controller_manager_config.yaml",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var err error
	//mqttControllerConfig, err := getConfig(configFile)

	options := ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		MetricsBindAddress:     metricsAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "aasdfasdf",
		Port:                   9443,
	}
	//setupLog.WithValues("config", configFile).Info("found config")

	//ctrlConfig := personaliotv1alpha1.MQTTControllerConfig{}
	//if configFile != "" {
	//	options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
	//	if err != nil {
	//		setupLog.Error(err, "unable to load the config file")
	//		os.Exit(1)
	//	}
	//}
	//
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.PowerstripReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		//MQTTSubscriber: mqttiot.NewPahoMQTTSubscriber(mqttiot.ClientOpts(mqttControllerConfig.MQTTConfig)),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Powerstrip")
		os.Exit(1)
	}
	if err = (&controllers.PoweroutletReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		//MQTTSubscriber: mqttiot.NewPahoMQTTSubscriber(mqttiot.ClientOpts(mqttControllerConfig.MQTTConfig)),
		//MQTTPublisher:  mqttiot.NewPahoMQTTPublisher(mqttiot.ClientOpts(mqttControllerConfig.MQTTConfig)),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Poweroutlet")
		os.Exit(1)
	}
	if err = (&controllers.LocationReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Location")
		os.Exit(1)
	}
	if err = (&personaliotv1alpha1.Poweroutlet{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Poweroutlet")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
