package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

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

var (
	DUCKDNS_TOKEN = os.Getenv("DUCKDNS_TOKEN")
	DOMAINS       = strings.Split(os.Getenv("DOMAINS"), ",")
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
		prefix := strings.TrimSuffix(host, ".duckdns.org")

		if !slices.Contains(DOMAINS, prefix) {
			logf.Log.Info("skipping reonciliation for domain not listed in the filter")
			continue
		}

		url := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&ip=%s&verbose=true", prefix, DUCKDNS_TOKEN, getIp(ing))
		logf.Log.Info(url)
		r, err := http.Get(url)
		if err != nil {
			logf.Log.Error(err, "unable to update domain", "host", host)
			return reconcile.Result{}, err
		}
		if r.StatusCode != 200 {
			logf.Log.Info("unable to update domain", "host", host, "statusCode", r.StatusCode)
			return reconcile.Result{}, errors.New("unable to update domain")
		}
		res, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logf.Log.Error(err, "unable to parse response", "host", host)
			return reconcile.Result{}, err
		}
		logf.Log.Info("got response", "r", string(res))
		if string(res) != "OK" {
			logf.Log.Info("unable to update domain", "host", host, "response", string(res))
			return reconcile.Result{}, errors.New("unable to update domain")
		}
	}

	return reconcile.Result{RequeueAfter: 60 * time.Second}, nil
}

func (a *IngressDnsController) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}

func getIp(ing *networkingv1.Ingress) string {
	ingStatus := ing.Status.LoadBalancer.Ingress
	if len(ingStatus) == 0 {
		return ""
	}
	if ingStatus[0].IP == "" {
		return ingStatus[0].Hostname
	}
	return ingStatus[0].IP
}
