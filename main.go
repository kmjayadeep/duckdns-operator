package main

import (
	"context"
	"os"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func main() {
	logf.SetLogger(zap.New())

	var log = logf.Log.WithName("builder-examples")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Complete(&IngressDnsController{})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

// IngressDnsController is a simple ControllerManagedBy example implementation.
type IngressDnsController struct {
	client.Client
}

func (a *IngressDnsController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ing := &networkingv1.Ingress{}
	err := a.Get(ctx, req.NamespacedName, ing)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, rule := range ing.Spec.Rules {
		host := rule.Host
		if !strings.HasSuffix(host, "duckdns.org") {
			continue
		}
		logf.Log.Info("got duckdns domain", "host", host)
	}

	return reconcile.Result{}, nil
}

func (a *IngressDnsController) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}
