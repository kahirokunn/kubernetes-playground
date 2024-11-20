package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	runtimecatalog "sigs.k8s.io/cluster-api/exp/runtime/catalog"
	runtimehooksv1 "sigs.k8s.io/cluster-api/exp/runtime/hooks/api/v1alpha1"
	"sigs.k8s.io/cluster-api/exp/runtime/server"
)

var (
	// catalog contains all information about RuntimeHooks.
	catalog = runtimecatalog.New()

	// Flags.
	profilerAddress string
	webhookPort     int
	webhookCertDir  string
	logOptions      = logs.NewOptions()
)

func init() {
	// Adds to the catalog all the RuntimeHooks defined in cluster API.
	_ = runtimehooksv1.AddToCatalog(catalog)
}

// InitFlags initializes the flags.
func InitFlags(fs *pflag.FlagSet) {
	// Initialize logs flags using Kubernetes component-base machinery.
	logsv1.AddFlags(logOptions, fs)

	// Add test-extension specific flags
	fs.StringVar(&profilerAddress, "profiler-address", "",
		"Bind address to expose the pprof profiler (e.g. localhost:6060)")

	fs.IntVar(&webhookPort, "webhook-port", 9443,
		"Webhook Server port")

	fs.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs/",
		"Webhook cert dir.")
}

func main() {
	// Creates a logger to be used during the main func.
	setupLog := ctrl.Log.WithName("setup")

	// Initialize and parse command line flags.
	InitFlags(pflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Set log level 2 as default.
	if err := pflag.CommandLine.Set("v", "10"); err != nil {
		setupLog.Error(err, "Failed to set default log level")
		os.Exit(1)
	}
	pflag.Parse()
	// logOptionsをJSON形式で出力
	logOptionsJSON, err := json.MarshalIndent(logOptions, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling logOptions to JSON: %v\n", err)
		return
	}
	fmt.Printf("Log Options (JSON): %s\n", logOptionsJSON)

	// Validates logs flags using Kubernetes component-base machinery and applies them
	if err := logsv1.ValidateAndApply(logOptions, nil); err != nil {
		setupLog.Error(err, "Unable to start extension")
		os.Exit(1)
	}

	// Add the klog logger in the context.
	ctrl.SetLogger(klog.Background())

	// Initialize the golang profiler server, if required.
	if profilerAddress != "" {
		klog.Infof("Profiler listening for requests at %s", profilerAddress)
		go func() {
			klog.Info(http.ListenAndServe(profilerAddress, nil))
		}()
	}

	// Create a http server for serving runtime extensions
	webhookServer, err := server.New(server.Options{
		Catalog: catalog,
		Port:    webhookPort,
		CertDir: webhookCertDir,
	})
	if err != nil {
		setupLog.Error(err, "Error creating webhook server")
		os.Exit(1)
	}

	// Register extension handlers.
	if err := webhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        runtimehooksv1.BeforeClusterCreate,
		Name:        "before-cluster-create",
		HandlerFunc: DoBeforeClusterCreate,
	}); err != nil {
		setupLog.Error(err, "Error adding handler")
		os.Exit(1)
	}
	if err := webhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        runtimehooksv1.BeforeClusterUpgrade,
		Name:        "before-cluster-upgrade",
		HandlerFunc: DoBeforeClusterUpgrade,
	}); err != nil {
		setupLog.Error(err, "Error adding handler")
		os.Exit(1)
	}
	if err := webhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        runtimehooksv1.BeforeClusterDelete,
		Name:        "before-cluster-delete",
		HandlerFunc: DoBeforeClusterDelete,
	}); err != nil {
		setupLog.Error(err, "Error adding handler")
		os.Exit(1)
	}
	if err := webhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        runtimehooksv1.AfterClusterUpgrade,
		Name:        "after-cluster-upgrade",
		HandlerFunc: DoAfterClusterUpgrade,
	}); err != nil {
		setupLog.Error(err, "Error adding handler")
		os.Exit(1)
	}
	if err := webhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        runtimehooksv1.AfterControlPlaneUpgrade,
		Name:        "after-control-plane-upgrade",
		HandlerFunc: DoAfterControlPlaneUpgrade,
	}); err != nil {
		setupLog.Error(err, "Error adding handler")
		os.Exit(1)
	}
	if err := webhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        runtimehooksv1.AfterControlPlaneInitialized,
		Name:        "after-control-plane-initialized",
		HandlerFunc: DoAfterControlPlaneInitialized,
	}); err != nil {
		setupLog.Error(err, "Error adding handler")
		os.Exit(1)
	}

	// Setup a context listening for SIGINT.
	ctx := ctrl.SetupSignalHandler()

	// Start the https server.
	setupLog.Info("Starting Runtime Extension server")
	if err := webhookServer.Start(ctx); err != nil {
		setupLog.Error(err, "Error running webhook server")
		os.Exit(1)
	}
}

func DoBeforeClusterCreate(ctx context.Context, request *runtimehooksv1.BeforeClusterCreateRequest, response *runtimehooksv1.BeforeClusterCreateResponse) {
	// log := ctrl.LoggerFrom(ctx)
	// log.Info("BeforeClusterCreate is called")
	klog.Info("BeforeClusterCreate is called")
	// Your implementation
	response.SetStatus(runtimehooksv1.ResponseStatusFailure)
}

func DoBeforeClusterUpgrade(ctx context.Context, request *runtimehooksv1.BeforeClusterUpgradeRequest, response *runtimehooksv1.BeforeClusterUpgradeResponse) {
	// log := ctrl.LoggerFrom(ctx)
	// log.Info("BeforeClusterUpgrade is called")
	klog.Info("BeforeClusterUpgrade is called")
	// Your implementation
	response.SetStatus(runtimehooksv1.ResponseStatusFailure)
}

func DoBeforeClusterDelete(ctx context.Context, request *runtimehooksv1.BeforeClusterDeleteRequest, response *runtimehooksv1.BeforeClusterDeleteResponse) {
	// log := ctrl.LoggerFrom(ctx)
	// log.Info("DoBeforeClusterDelete is called")
	klog.Info("DoBeforeClusterDelete is called")
	// Your implementation
	response.SetStatus(runtimehooksv1.ResponseStatusSuccess)
}

func DoAfterClusterUpgrade(ctx context.Context, request *runtimehooksv1.AfterClusterUpgradeRequest, response *runtimehooksv1.AfterClusterUpgradeResponse) {
	// log := ctrl.LoggerFrom(ctx)
	// log.Info("AfterClusterUpgrade is called")
	klog.Info("AfterClusterUpgrade is called")
	// Your implementation
	response.SetStatus(runtimehooksv1.ResponseStatusFailure)
}

func DoAfterControlPlaneUpgrade(ctx context.Context, request *runtimehooksv1.AfterControlPlaneUpgradeRequest, response *runtimehooksv1.AfterControlPlaneUpgradeResponse) {
	// log := ctrl.LoggerFrom(ctx)
	// log.Info("AfterControlPlaneUpgrade is called")
	klog.Info("AfterControlPlaneUpgrade is called")
	// Your implementation
	response.SetStatus(runtimehooksv1.ResponseStatusFailure)
}

func DoAfterControlPlaneInitialized(ctx context.Context, request *runtimehooksv1.AfterControlPlaneInitializedRequest, response *runtimehooksv1.AfterControlPlaneInitializedResponse) {
	// log := ctrl.LoggerFrom(ctx)
	// log.Info("AfterControlPlaneInitialized is called")
	klog.Info("AfterControlPlaneInitialized is called")
	// Your implementation
	response.SetStatus(runtimehooksv1.ResponseStatusFailure)
}

