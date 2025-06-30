/*
Copyright 2025.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	redisv1alpha1 "github.com/pehlicd/redis-operator/api/v1alpha1"
)

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	// FinalizerName is the name of the finalizer used for Redis resources.
	FinalizerName = "redis.yazio.com/finalizer"
)

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.yazio.com,resources=redis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.yazio.com,resources=redis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.yazio.com,resources=redis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	redis := &redisv1alpha1.Redis{}

	err := r.Get(ctx, req.NamespacedName, redis)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconciled()
		}
		log.Error(err, "unable to fetch Redis")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: RequeueDelay,
		}, err
	}

	// if status condition status is not set, it indicates that it is the first time operator gets this resource
	// so we need to set initial status conditions to let users know operator is working on this specific resource
	if len(redis.Status.Conditions) == 0 {
		meta.SetStatusCondition(&redis.Status.Conditions, metav1.Condition{
			Type:    "Available",
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Reconciling redis",
		})

		if err := r.Status().Update(ctx, redis); err != nil {
			log.Error(err, "unable to update Redis")
			return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
		}
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(redis, FinalizerName) {
		log.Info("Adding finalizer to Redis", "name", redis.Name)

		if err := r.Get(ctx, req.NamespacedName, redis); err != nil {
			log.Error(err, "unable to fetch Redis")
			return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
		}

		controllerutil.AddFinalizer(redis, FinalizerName)
		if err := r.Update(ctx, redis); err != nil {
			return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
		}
	}

	// Remove finalizer if deletion initiated
	if !redis.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Redis is being deleted, starting cleanup for the instance", "name", redis.Name)
		if controllerutil.ContainsFinalizer(redis, FinalizerName) {
			// Since every resource created with controller reference once redis cr is deleted,
			// other resource will be deleted by garbage collector automatically
			// so controller only needs to remove finalizer and throw event
			log.Info("Removing finalizer from Redis", "name", redis.Name)

			if err := r.Get(ctx, req.NamespacedName, redis); err != nil {
				log.Error(err, "unable to get Redis while removing finalizer", "name", redis.Name)
			}

			controllerutil.RemoveFinalizer(redis, FinalizerName)
			if err := r.Update(ctx, redis); err != nil {
				return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
			}

			log.Info("Finalizer removed from Redis", "name", redis.Name)
		}

		return reconciled()
	}

	// Reconcile Redis owned resources update/create if needed
	if _, err := r.reconcileSecret(ctx, redis); err != nil {
		return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
	}
	if _, err := r.reconcileService(ctx, redis); err != nil {
		return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
	}
	deployment, err := r.reconcileDeployment(ctx, redis)
	if err != nil {
		return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
	}

	if err = r.updateStatus(ctx, redis, deployment); err != nil {
		return requeueInstanceWithError(ctx, redis.Name, redis.Namespace, err)
	}

	return reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisv1alpha1.Redis{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Named("redis").
		Complete(r)
}
