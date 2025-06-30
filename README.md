# redis-operator

This project contains a Kubernetes Operator built with the Operator SDK and Golang. Its purpose is to manage the lifecycle of Redis instances within a Kubernetes cluster, automating their deployment, configuration, and cleanup.

## Features
- Declarative Redis Deployments: Manage Redis instances using a simple, declarative Custom Resource (Redis).

- Automated Password Management: Automatically generates a strong, random password for each new Redis instance.

- Secure Secret Storage: Stores the generated password securely in a Kubernetes Secret.

- Automatic Deployment: Creates a Kubernetes Deployment using the bitnami/redis image, configured to use the generated password.

- Scaling: Update the spec.replicas field in the Custom Resource to scale the number of Redis replicas up or down.

- Automated Cleanup: Uses a finalizer to ensure that when a Redis resource is deleted, its associated Deployment, Service and Secret are also garbage collected.

## Project Structure
```text
.
├── api/
│   └── v1alpha1/
│       ├── redis_types.go      # Defines the Redis CRD schema (Spec and Status)
│       └── ...
├── internal/
│   └── controller/
│       ├── redis_controller.go     # Main reconciliation logic for the Redis operator
│       └── redis_controller_test.go # Unit tests for the controller
├── config/
│   ├── crd/
│   │   └── bases/
│   │       └── redis.yazio.com_redis.yaml # The CRD manifest
│   ├── samples/
│   │   └── redis_v1alpha1_redis.yaml  # An example Custom Resource manifest
│   └── ...
├── Dockerfile
├── Makefile
└── README.md
└── CRD_DOCS.md
└── ...
```

### CRD Documentation
Please visit [CRD_DOCS.md](CRD_DOCS.md) for the CRD documentation.

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

#### [Optional] Kind Cluster Installation

```bash
kind create cluster --config config/kind.yaml
```

Lastly to install ingress-nginx run if you need an ingress controller:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

### Installations

#### One command installation
```sh
kubectl apply -f https://raw.githubusercontent.com/pehlicd/redis-operator/main/dist/install.yaml
```

#### Helm installation
To install the operator using Helm, you can use the following commands:

```sh
helm install redis-operator ./charts/redis-operator
```

#### Manual installation

Clone the repository and run the following commands:

```sh
make deploy
```

In order to uninstall the operator, run:

```sh
make undeploy
```

#### Running the operator locally

To run the operator locally, you can use the following command:

```sh
make run
```

### Deploying your first Redis instance
To deploy and discover your first Redis instance, you can use the provided sample manifest:

```sh
kubectl apply -f config/samples/redis_v1alpha1_redis.yaml

#OR

kubectl apply -f https://raw.githubusercontent.com/pehlicd/redis-operator/main/config/samples/redis_v1alpha1_redis.yaml
```

After applying the manifest, you can check and verify the objects of the Redis instance were created:

```sh
# Check that the Redis custom resource was created and its status is updated
kubectl get redis.redis.yazio.com redis-sample -o yaml

# Check that the Deployment was created
kubectl get deployment redis-sample

# Check that the Secret was created
kubectl get secret redis-password

# Check that the Service was created
kubectl get service redis-service
```

To check random generated password in the secret in use by the Redis instance, you can run:

```sh
kubectl get deployment redis-sample -o jsonpath='{.spec.template.spec.containers[0].env}' | jq
```

### Running tests
To run the tests, you can use the following command:

```sh
make test
```
