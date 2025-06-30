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
│       └── redis_controller_test.go # (Bonus) Unit tests for the controller
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
```

### CRD Documentation
Please visit [CRD_DOCS.md](CRD_DOCS.md) for the CRD documentation.

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Installations

#### One command installation
```sh
kubectl apply -f https://raw.githubusercontent.com/pehlicd/redis-operator/main/dist/install.yaml
```


## License

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

