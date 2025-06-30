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
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/pehlicd/redis-operator/api/v1alpha1"
)

// reconcileSecret ensures the secret for the Redis instance exists.
func (r *RedisReconciler) reconcileSecret(ctx context.Context, redis *v1alpha1.Redis) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)
	secretName := redis.Spec.PasswordSecretName
	secret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: redis.Namespace}, secret); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating a new Secret", "Secret.Namespace", redis.Namespace, "Secret.Name", secretName)
			password, err := generateRandomPassword(16)
			if err != nil {
				return nil, fmt.Errorf("failed to generate random password: %w", err)
			}

			newSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: redis.Namespace},
				Type:       corev1.SecretTypeOpaque,
				StringData: map[string]string{"password": password},
			}
			if err := ctrl.SetControllerReference(redis, newSecret, r.Scheme); err != nil {
				return nil, err
			}
			if err = r.Create(ctx, newSecret); err != nil {
				return nil, err
			}
			r.Recorder.Event(redis, corev1.EventTypeNormal, "CreatedSecret", fmt.Sprintf("Created secret %s", secretName))
			return newSecret, nil
		}
		return nil, err
	}

	return secret, nil
}

// reconcileService ensures the service for the Redis instance is up-to-date.
func (r *RedisReconciler) reconcileService(ctx context.Context, redis *v1alpha1.Redis) (*corev1.Service, error) {
	logger := log.FromContext(ctx)
	desiredSvc := r.serviceForRedis(redis)
	foundSvc := &corev1.Service{}

	err := r.Get(ctx, types.NamespacedName{Name: desiredSvc.Name, Namespace: redis.Namespace}, foundSvc)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating a new Service", "Service.Namespace", desiredSvc.Namespace, "Service.Name", desiredSvc.Name)
			if err = r.Create(ctx, desiredSvc); err != nil {

				return nil, err
			}
			r.Recorder.Event(redis, corev1.EventTypeNormal, "CreatedService", fmt.Sprintf("Created service %s", desiredSvc.Name))
			return desiredSvc, nil
		}
		return nil, err
	}

	patch := client.MergeFrom(desiredSvc)
	needsUpdate := false
	if !reflect.DeepEqual(foundSvc.Spec.Type, desiredSvc.Spec.Type) {
		foundSvc.Spec.Type = desiredSvc.Spec.Type
		needsUpdate = true
	}
	for i, port := range foundSvc.Spec.Ports {
		if port.Port != desiredSvc.Spec.Ports[i].Port || port.Name != desiredSvc.Spec.Ports[i].Name {
			foundSvc.Spec.Ports[i] = desiredSvc.Spec.Ports[i]
			needsUpdate = true
		}
	}
	if !reflect.DeepEqual(foundSvc.Spec.Selector, desiredSvc.Spec.Selector) {
		foundSvc.Spec.Selector = desiredSvc.Spec.Selector
		needsUpdate = true
	}

	if needsUpdate {
		if err := r.Patch(ctx, foundSvc, patch); err != nil {
			return nil, err
		}
		r.Recorder.Event(redis, corev1.EventTypeNormal, "UpdatedService", fmt.Sprintf("Updated service %s", desiredSvc.Name))
	}

	return foundSvc, nil
}

// reconcileDeployment ensures the deployment for the Redis instance is up-to-date.
func (r *RedisReconciler) reconcileDeployment(ctx context.Context, redis *v1alpha1.Redis) (*appsv1.Deployment, error) {
	logger := log.FromContext(ctx)
	desiredDep := r.deploymentForRedis(redis)
	foundDep := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{Name: desiredDep.Name, Namespace: redis.Namespace}, foundDep)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating a new Deployment", "Deployment.Namespace", desiredDep.Namespace, "Deployment.Name", desiredDep.Name)
			if err = r.Create(ctx, desiredDep); err != nil {
				return nil, err
			}
			r.Recorder.Event(redis, corev1.EventTypeNormal, "CreatedDeployment", fmt.Sprintf("Created deployment %s", desiredDep.Name))
			return desiredDep, nil
		}
		return nil, err
	}

	needsUpdate := false
	patch := client.MergeFrom(foundDep.DeepCopy())
	if !replicasMatch(foundDep.Spec.Replicas, desiredDep.Spec.Replicas) {
		foundDep.Spec.Replicas = desiredDep.Spec.Replicas
		needsUpdate = true
		logger.Info("Replica count changed", "From", foundDep.Spec.Replicas, "To", desiredDep.Spec.Replicas)
	}

	if !templatesMatch(&foundDep.Spec.Template, &desiredDep.Spec.Template) {
		foundDep.Spec.Template = desiredDep.Spec.Template
		needsUpdate = true
		logger.Info("Pod template changed")
	}

	if needsUpdate {
		logger.Info("Updating Deployment", "Deployment.Name", foundDep.Name)
		if err = r.Patch(ctx, foundDep, patch); err != nil {
			return nil, err
		}
		r.Recorder.Event(redis, corev1.EventTypeNormal, "UpdatedDeployment", "Deployment spec updated.")
	}

	return foundDep, nil
}

// replicasMatch compares the replica counts of two deployments.
func replicasMatch(foundReplicas *int32, desiredReplicas *int32) bool {
	if foundReplicas == nil && desiredReplicas == nil {
		return true
	}
	if foundReplicas == nil || desiredReplicas == nil {
		return false
	}
	return *foundReplicas == *desiredReplicas
}

// templatesMatch semantically compares two PodTemplateSpecs for the fields we manage.
func templatesMatch(found *corev1.PodTemplateSpec, desired *corev1.PodTemplateSpec) bool {
	if found.Spec.Containers == nil && len(found.Spec.Containers) < 1 {
		return false
	}

	foundContainer := &found.Spec.Containers[0]
	desiredContainer := &desired.Spec.Containers[0]

	if foundContainer.Image != desiredContainer.Image {
		return false
	}
	if !reflect.DeepEqual(foundContainer.Resources, desiredContainer.Resources) {
		return false
	}
	if foundContainer.LivenessProbe == nil || foundContainer.ReadinessProbe == nil {
		return false
	}
	if !reflect.DeepEqual(foundContainer.Env, desiredContainer.Env) {
		return false
	}
	return true
}

// updateStatus updates the status subresource of the Redis CR using a patch.
func (r *RedisReconciler) updateStatus(ctx context.Context, redis *v1alpha1.Redis, deployment *appsv1.Deployment) error {
	statusCopy := redis.DeepCopy()

	statusCopy.Status.PasswordSecretName = redis.Spec.PasswordSecretName
	desiredReplicas := *redis.Spec.Replicas

	// Add a nil check for the deployment to prevent panics early in the reconciliation.
	if deployment != nil && deployment.Status.AvailableReplicas == desiredReplicas {
		meta.SetStatusCondition(&statusCopy.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			Reason:             "DeploymentAvailable",
			Message:            "Redis deployment is fully available",
			ObservedGeneration: redis.Generation,
		})
	} else {
		meta.SetStatusCondition(&statusCopy.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionFalse,
			Reason:             "Reconciling",
			Message:            "Redis deployment is not yet fully available",
			ObservedGeneration: redis.Generation,
		})
	}

	patch := client.MergeFrom(redis)

	return r.Status().Patch(ctx, statusCopy, patch)
}

// serviceForRedis returns a Redis Service object.
func (r *RedisReconciler) serviceForRedis(redis *v1alpha1.Redis) *corev1.Service {
	labels := labelsForRedis(redis.Name)
	spec := redis.Spec.Service

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: spec.Name, Namespace: redis.Namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    []corev1.ServicePort{{Port: *spec.Port, TargetPort: intstr.FromInt32(*redis.Spec.Port), Name: "redis"}},
			Type:     corev1.ServiceType(spec.Type),
		},
	}

	_ = ctrl.SetControllerReference(redis, svc, r.Scheme)

	return svc
}

// deploymentForRedis returns a Redis Deployment object.
func (r *RedisReconciler) deploymentForRedis(redis *v1alpha1.Redis) *appsv1.Deployment {
	labels := labelsForRedis(redis.Name)

	envVars := []corev1.EnvVar{{
		Name: "REDIS_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: redis.Spec.PasswordSecretName},
				Key:                  "password",
			},
		},
	}}
	if redis.Spec.Env != nil {
		envVars = append(envVars, *redis.Spec.Env...)
	}

	// Define default Liveness Probe
	livenessProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(*redis.Spec.Port)},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      1,
		PeriodSeconds:       20,
		FailureThreshold:    3,
	}
	// If user provided a LivenessProbe in the spec, use that instead of the default.
	if redis.Spec.LivenessProbe != nil {
		livenessProbe = redis.Spec.LivenessProbe
	}

	// Define default Readiness Probe
	readinessProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"redis-cli", "ping"},
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		FailureThreshold:    3,
	}
	// If user provided a ReadinessProbe in the spec, use that instead of the default.
	if redis.Spec.ReadinessProbe != nil {
		readinessProbe = redis.Spec.ReadinessProbe
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.Name,
			Namespace: redis.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: redis.Spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:          redis.Spec.Image,
						Name:           "redis",
						Ports:          []corev1.ContainerPort{{ContainerPort: *redis.Spec.Port, Name: "redis"}},
						Env:            envVars,
						Resources:      redis.Spec.Resources,
						ReadinessProbe: readinessProbe,
						LivenessProbe:  livenessProbe,
					}},
				},
			},
		},
	}

	_ = ctrl.SetControllerReference(redis, dep, r.Scheme)

	return dep
}

func labelsForRedis(name string) map[string]string {
	return map[string]string{"app": "redis", "redis_cr": name}
}

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}
