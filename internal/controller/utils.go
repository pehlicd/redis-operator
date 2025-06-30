package controller

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// RequeueDelay is the default delay for re-queuing the reconciliation loop.
	RequeueDelay = 10 * time.Second
)

func reconciled() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func requeueInstanceWithError(ctx context.Context, instanceName string, namespace string, err error) (ctrl.Result, error) {
	logf.FromContext(ctx).Error(err, "Re-queuing Redis instance due to error", "name", instanceName, "namespace", namespace, "after", RequeueDelay)
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: RequeueDelay,
	}, err
}
