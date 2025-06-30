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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RedisSpec defines the desired state of Redis.
type RedisSpec struct {
	// Image is the container image for the Redis instance.
	// +kubebuilder:validation:Required
	// +kubebuilder:default="bitnami/redis"
	Image string `json:"image"`
	// Replicas is the number of desired replicas.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`
	// Port is the port on which Redis will listen.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=6379
	Port *int32 `json:"port,omitempty"`
	// PasswordSecretName is the name of the secret containing the Redis password.
	// +kubebuilder:validation:Required
	// +kubebuilder:default="redis-password"
	PasswordSecretName string `json:"passwordSecretName,omitempty"`
	// Env is a list of environment variables to set in the Redis container.
	// +kubebuilder:validation:Optional
	Env *[]corev1.EnvVar `json:"env,omitempty"`
	// Service defines the service configuration for Redis.
	// +kubebuilder:validation:Required
	// +kubebuilder:default={name: "redis-service", type: "ClusterIP", port: 6379}
	Service Service `json:"service,omitempty"`
	// Resources defines the resource requirements for the Redis pods.
	// +kubebuilder:validation:Required
	// +kubebuilder:default={requests: {cpu: "100m", memory: "128Mi"}, limits: {cpu: "500m", memory: "512Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// ReadinessProbe is the probe to check if the Redis instance is ready.
	// +kubebuilder:validation:Optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	// LivenessProbe is the probe to check if the Redis instance is alive.
	// +kubebuilder:validation:Optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
}

// Service defines the service configuration for Redis.
type Service struct {
	// Name is the name of the service.
	// +kubebuilder:validation:Required
	// +kubebuilder:default="redis-service"
	Name string `json:"name"`
	// Type is the type of service (ClusterIP, NodePort, LoadBalancer).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// +kubebuilder:default=ClusterIP
	Type string `json:"type,omitempty"`
	// Port is the port on which the service will be exposed.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=6379
	Port *int32 `json:"port,omitempty"`
}

// RedisStatus defines the observed state of Redis.
type RedisStatus struct {
	// PasswordSecretName is the name of the secret containing the Redis password.
	PasswordSecretName string `json:"passwordSecretName"`
	// Conditions store the status conditions of the Redis instances
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Redis is the Schema for the redis API.
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedisList contains a list of Redis.
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
