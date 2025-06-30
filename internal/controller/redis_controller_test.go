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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisv1alpha1 "github.com/pehlicd/redis-operator/api/v1alpha1"
)

var _ = Describe("Redis Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			resourceNamespace = "default"

			timeout  = time.Second * 20
			interval = time.Millisecond * 250
		)

		var (
			redisLookupKey      types.NamespacedName
			deploymentLookupKey types.NamespacedName
			serviceLookupKey    types.NamespacedName
			secretLookupKey     types.NamespacedName
		)

		ctx := context.Background()

		redis := &redisv1alpha1.Redis{}

		BeforeEach(func() {
			// Define keys for the test resources.
			redisLookupKey = types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}
			deploymentLookupKey = types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}
			serviceLookupKey = types.NamespacedName{Name: "redis-service", Namespace: resourceNamespace}
			secretLookupKey = types.NamespacedName{Name: "redis-password", Namespace: resourceNamespace}

			By("creating the custom resource for the Kind Redis")
			err := k8sClient.Get(ctx, redisLookupKey, redis)
			if err != nil && errors.IsNotFound(err) {
				replicas := int32(1)
				port := int32(6379)
				resource := &redisv1alpha1.Redis{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: resourceNamespace,
					},
					Spec: redisv1alpha1.RedisSpec{
						Replicas:           &replicas,
						Image:              "bitnami/redis",
						Port:               &port,
						PasswordSecretName: "redis-password",
						Service: redisv1alpha1.Service{
							Name: "redis-service",
							Type: string(corev1.ServiceTypeClusterIP),
							Port: &port,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"redis-cli", "ping"},
								},
							},
							InitialDelaySeconds: 5,
							TimeoutSeconds:      1,
							PeriodSeconds:       10,
							FailureThreshold:    3,
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(port)},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      1,
							PeriodSeconds:       20,
							FailureThreshold:    3,
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, deploymentLookupKey, &appsv1.Deployment{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			resource := &redisv1alpha1.Redis{}
			err := k8sClient.Get(ctx, redisLookupKey, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Redis")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RedisReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: &record.FakeRecorder{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: redisLookupKey,
			})
			Expect(err).NotTo(HaveOccurred())

			// Wait for the reconciliation to complete and get the updated resource
			Eventually(func() bool {
				err := k8sClient.Get(ctx, redisLookupKey, redis)
				return err == nil
			}, timeout, interval).Should(BeTrue(), "Reconciliation should complete and update the status")

			// Check secret if it was created
			By("Checking if the Secret is created")
			secret := &corev1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, secret)
				return err == nil
			}, timeout, interval).Should(BeTrue(), "Secret should be created")
			Expect(secret.Data).To(HaveKey("password"))
			Expect(secret.Data["password"]).NotTo(BeEmpty())

			// Check service if it was created
			By("Checking if the Service is created")
			service := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceLookupKey, service)
				return err == nil
			}, timeout, interval).Should(BeTrue(), "Service should be created")
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(6379)))
			Expect(service.Spec.Selector).To(HaveKeyWithValue("app", "redis"))

			// Check deployment if it was created
			By("Checking if the Deployment is created")
			deployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, deploymentLookupKey, deployment)
				return err == nil
			}, timeout, interval).Should(BeTrue(), "Deployment should be created")
			Expect(deployment.Spec.Replicas).To(Equal(redis.Spec.Replicas))
			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(redis.Spec.Image))
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(int32(6379)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Env).To(ContainElement(
				corev1.EnvVar{
					Name: "REDIS_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretLookupKey.Name,
							},
							Key: "password",
						},
					},
				},
			))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources).To(Equal(redis.Spec.Resources))
			Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(Equal(
				&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"redis-cli", "ping"},
						},
					},
					InitialDelaySeconds: 5,
					TimeoutSeconds:      1,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					FailureThreshold:    3,
				},
			))
			Expect(deployment.Spec.Template.Spec.Containers[0].LivenessProbe).To(Equal(
				&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(int32(6379))},
					},
					InitialDelaySeconds: 15,
					TimeoutSeconds:      1,
					PeriodSeconds:       20,
					SuccessThreshold:    1,
					FailureThreshold:    3,
				},
			))
			Expect(deployment.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app", "redis"))

			// Check if the finalizer is added
			By("Checking if the finalizer is added to the Redis resource")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, redisLookupKey, redis)
				if err != nil {
					return false
				}
				return controllerutil.ContainsFinalizer(redis, FinalizerName)
			}, timeout, interval).Should(BeTrue(), "Finalizer should be added to the Redis resource")

			// Check if the status is updated
			By("Checking if the status is updated in the Redis resource")
			Eventually(func() error {
				err := k8sClient.Get(ctx, redisLookupKey, redis)
				if err != nil {
					return err
				}
				return nil
			}, timeout, interval).Should(Succeed(), "Status should be updated with the secret name")
			Expect(redis.Status.PasswordSecretName).To(Equal(secretLookupKey.Name))

			By("Checking replica count update in the Deployment")
			newReplicas := int32(3)
			redis.Spec.Replicas = &newReplicas
			Expect(k8sClient.Update(ctx, redis)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: redisLookupKey,
			})
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() int32 {
				deployment := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, deploymentLookupKey, deployment)
				if err != nil {
					return 0
				}
				if deployment.Spec.Replicas != nil {
					return *deployment.Spec.Replicas
				}
				return 0
			}, timeout, interval).Should(Equal(newReplicas), "Deployment replicas should be updated to 3")

			By("Checking if new environment variable is added to the Deployment")
			err = k8sClient.Get(ctx, redisLookupKey, redis)
			Expect(err).NotTo(HaveOccurred())

			redis.Spec.Env = &[]corev1.EnvVar{
				{
					Name:  "REDIS_ENV_VAR",
					Value: "test-value",
				},
			}
			Expect(k8sClient.Update(ctx, redis)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: redisLookupKey,
			})
			Expect(err).NotTo(HaveOccurred())

			deployment = &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, deploymentLookupKey, deployment)
				return err == nil
			}, timeout, interval).Should(BeTrue(), "Deployment should be updated")
			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Env).To(HaveLen(2))
			Expect(deployment.Spec.Template.Spec.Containers[0].Env).To(ContainElement(
				corev1.EnvVar{
					Name:  "REDIS_ENV_VAR",
					Value: "test-value",
				},
			))
		})
	})
})
