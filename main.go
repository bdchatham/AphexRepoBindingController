package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	triggersv1beta1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"github.com/bdchatham/AphexControllerRuntime/pkg/config"
	"github.com/bdchatham/AphexControllerRuntime/pkg/constants"
	"github.com/bdchatham/AphexControllerRuntime/pkg/metrics"

	repocontroller "github.com/bdchatham/AphexRepoBindingController/controller"
	"github.com/bdchatham/AphexRepoBindingController/controller/webhooks"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))
	utilruntime.Must(tektonv1.AddToScheme(scheme))
	utilruntime.Must(triggersv1beta1.AddToScheme(scheme))
}

func main() {
	os.Exit(run())
}

func run() int {
	var metricsAddr string
	var enableLeaderElection bool
	var disableLeaderElection bool
	var probeAddr string
	var enableWebhooks bool
	var webhookPort int
	var certDir string
	var approvedOrgs string
	var maxConcurrentReconciles int
	var developmentMode bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true, "Enable leader election for controller manager.")
	flag.BoolVar(&disableLeaderElection, "disable-leader-elect", false, "Disable leader election for development.")
	flag.BoolVar(&enableWebhooks, "enable-webhooks", false, "Enable admission webhooks.")
	flag.IntVar(&webhookPort, "webhook-port", 9443, "The port the webhook server binds to.")
	flag.StringVar(&certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "TLS certificates directory.")
	flag.StringVar(&approvedOrgs, "approved-orgs", "bdchatham", "Comma-separated approved GitHub organizations.")
	flag.IntVar(&maxConcurrentReconciles, "max-concurrent-reconciles", constants.DefaultMaxConcurrentReconciles, "Max concurrent reconciles.")
	flag.BoolVar(&developmentMode, "development", false, "Enable development mode with verbose logging.")

	opts := zap.Options{Development: developmentMode}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	if !developmentMode {
		if envDev := os.Getenv(constants.EnvDevelopmentMode); envDev == "true" || envDev == "1" {
			developmentMode = true
			opts.Development = true
		}
	}

	if !disableLeaderElection {
		if envLeader := os.Getenv(constants.EnvLeaderElectionEnable); envLeader != "" {
			enableLeaderElection = envLeader == "true" || envLeader == "1"
		}
	}
	if disableLeaderElection {
		enableLeaderElection = false
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	approvedOrgsList := parseApprovedOrgs(approvedOrgs)
	setupLog.Info("Approved organizations configured", "orgs", approvedOrgsList)

	configManager := config.NewManager(setupLog)
	if err := configManager.LoadFromEnvironment(); err != nil {
		setupLog.Error(err, "Failed to load configuration")
		return 1
	}
	controllerConfig := configManager.GetConfig()

	rateLimiter := workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
		controllerConfig.BaseDelay,
		controllerConfig.MaxDelay,
	)

	mgrOpts := ctrl.Options{
		Scheme:                  scheme,
		Metrics:                 server.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "repobinding-controller.aphex.io",
		GracefulShutdownTimeout: &controllerConfig.ShutdownTimeout,
	}

	if enableWebhooks {
		mgrOpts.WebhookServer = webhook.NewServer(webhook.Options{
			Port:    webhookPort,
			CertDir: certDir,
		})
		setupLog.Info("Webhooks enabled", "port", webhookPort, "certDir", certDir)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), mgrOpts)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return 1
	}

	controllerOpts := controller.Options{
		MaxConcurrentReconciles: maxConcurrentReconciles,
		RateLimiter:             rateLimiter,
	}

	templateCatalog := repocontroller.NewTemplateCatalog()
	setupLog.Info("Template catalog initialized", "templates", templateCatalog.List())

	if err = (&repocontroller.RepoBindingReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		Log:             ctrl.Log.WithName("controllers").WithName("RepoBinding"),
		TemplateCatalog: templateCatalog,
		Config:          controllerConfig,
	}).SetupWithManagerAndOptions(mgr, &controllerOpts); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RepoBinding")
		return 1
	}

	if enableWebhooks {
		repoBindingValidator := webhooks.NewRepoBindingValidator(
			ctrl.Log.WithName("webhooks").WithName("RepoBinding"),
			approvedOrgsList,
		)
		if err = repoBindingValidator.SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RepoBinding")
			return 1
		}
		setupLog.Info("Webhooks configured successfully")
	}

	healthState := metrics.GetHealthState()
	healthzChecker := func(_ *http.Request) error {
		if !healthState.IsHealthy() {
			return &healthCheckError{message: "no recent successful reconciliations"}
		}
		return nil
	}
	if err := mgr.AddHealthzCheck("healthz", healthzChecker); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return 1
	}
	if err := mgr.AddReadyzCheck("readyz", healthzChecker); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return 1
	}

	_ = metrics.GetCollector()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		setupLog.Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	setupLog.Info("Starting RepoBinding controller",
		"leaderElection", enableLeaderElection,
		"developmentMode", developmentMode,
	)

	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "Problem running manager")
		return 1
	}

	setupLog.Info("Manager stopped gracefully")
	return 0
}

type healthCheckError struct {
	message string
}

func (e *healthCheckError) Error() string {
	return e.message
}

func parseApprovedOrgs(orgs string) []string {
	if orgs == "" {
		return nil
	}
	parts := strings.Split(orgs, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
