# API Reference

## Packages
- [redis.yazio.com/v1alpha1](#redisyaziocomv1alpha1)


## redis.yazio.com/v1alpha1

Package v1alpha1 contains API Schema definitions for the redis v1alpha1 API group.

### Resource Types
- [Redis](#redis)
- [RedisList](#redislist)



#### Redis



Redis is the Schema for the redis API.



_Appears in:_
- [RedisList](#redislist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `redis.yazio.com/v1alpha1` | | |
| `kind` _string_ | `Redis` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[RedisSpec](#redisspec)_ |  |  |  |


#### RedisList



RedisList contains a list of Redis.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `redis.yazio.com/v1alpha1` | | |
| `kind` _string_ | `RedisList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[Redis](#redis) array_ |  |  |  |


#### RedisSpec



RedisSpec defines the desired state of Redis.



_Appears in:_
- [Redis](#redis)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image is the container image for the Redis instance. | bitnami/redis | Required: \{\} <br /> |
| `replicas` _integer_ | Replicas is the number of desired replicas. | 1 | Minimum: 1 <br />Required: \{\} <br /> |
| `port` _integer_ | Port is the port on which Redis will listen. | 6379 | Minimum: 1 <br />Required: \{\} <br /> |
| `passwordSecretName` _string_ | PasswordSecretName is the name of the secret containing the Redis password. | redis-password | Required: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#envvar-v1-core)_ | Env is a list of environment variables to set in the Redis container. |  | Optional: \{\} <br /> |
| `service` _[Service](#service)_ | Service defines the service configuration for Redis. | \{ name:redis-service port:6379 type:ClusterIP \} | Required: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#resourcerequirements-v1-core)_ | Resources defines the resource requirements for the Redis pods. | \{ limits:map[cpu:500m memory:512Mi] requests:map[cpu:100m memory:128Mi] \} | Required: \{\} <br /> |
| `readinessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#probe-v1-core)_ | ReadinessProbe is the probe to check if the Redis instance is ready. |  | Optional: \{\} <br /> |
| `livenessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#probe-v1-core)_ | LivenessProbe is the probe to check if the Redis instance is alive. |  | Optional: \{\} <br /> |




#### Service



Service defines the service configuration for Redis.



_Appears in:_
- [RedisSpec](#redisspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the service. | redis-service | Required: \{\} <br /> |
| `type` _string_ | Type is the type of service (ClusterIP, NodePort, LoadBalancer). | ClusterIP | Enum: [ClusterIP NodePort LoadBalancer] <br />Required: \{\} <br /> |
| `port` _integer_ | Port is the port on which the service will be exposed. | 6379 | Minimum: 1 <br />Required: \{\} <br /> |


