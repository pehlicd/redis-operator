## Initialize operator structure

Using operator-sdk cli to initialize the operator structure:

```bash
operator-sdk init --domain yazio.com --repo github.com/pehlicd/redis-operator --plugins=go/v4
```

## Create API and Controller

Using operator-sdk cli to create API and controller:

```bash
operator-sdk create api --group redis --version v1alpha1 --kind Redis --resource --controller
```

## Install kind cluster and ingress-nginx

```bash
kind create cluster --config config/kind.yaml
```

Lastly to install ingress-nginx run:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```