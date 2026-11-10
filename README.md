# Caddy Operator

Caddy Operator is a Kubernetes operator that manages Caddy workloads through the namespaced `Caddy` custom resource (`web.local.com/v1`). For each resource, it reconciles a Deployment and Service. When `spec.domain` is set, it also creates an Ingress; clearing the domain removes the Ingress managed by that resource.

## Prerequisites

- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.
- Helm 3 to package or publish the chart.

For this checkout, [go.mod](go.mod) declares Go 1.27.3, and the CRD uses `apiextensions.k8s.io/v1` (Kubernetes 1.16+). These are the effective minimum versions for building and deploying the current project. The deploying identity also needs permission to install CRDs and cluster-scoped RBAC resources.

## Deploy

Build and push the operator image to a registry accessible by the cluster:

```sh
make docker-build docker-push IMG=<some-registry>/caddy-op:<tag>
```

Install the CRD and deploy the operator:

```sh
make install
make deploy IMG=<some-registry>/caddy-op:<tag>
```

The operator is deployed in the `dev` namespace. The deploying identity needs the cluster permissions required by the CRD and RBAC resources.

## Create a Caddy resource

The sample resource is in the `dev` namespace. Create that namespace, then apply the sample:

```sh
kubectl create namespace dev
kubectl apply -k config/samples/
```

Example resource:

```yaml
apiVersion: web.local.com/v1
kind: Caddy
metadata:
    name: caddy-sample
    namespace: dev
spec:
    replicas: 3
    image: caddy:2.10-alpine
    port: 80
    svcPort: 80
    domain: caddy.example.com
    # ingressClassName: nginx
```

`replicas`, `image`, `port`, and `svcPort` configure the workload and Service. `domain` is optional; when set, the operator creates an Ingress routing `/` to the Service. `ingressClassName` is optional. The cluster must have an Ingress controller to serve that route.

Check the custom resource and the objects managed for it:

```sh
kubectl get caddies -n dev
kubectl get deployments,services,ingresses -n dev
```

## Helm chart

Package the chart locally:

```sh
make helm-package VERSION=<version>
```

To publish it to Quay, first authenticate with Helm, then push:

```sh
helm registry login quay.io
make helm-push VERSION=<version> HELM_REPO=quay.io/serialt/charts
```

The chart is published as `oci://quay.io/serialt/charts/caddy-operator:<version>`. Install a published version with:

```sh
helm install caddy-operator oci://quay.io/serialt/charts/caddy-operator \
    --version <version> --namespace dev --create-namespace
```

GitHub Actions builds and pushes the operator image to Aliyun and Quay, and packages and publishes the chart on each push. Configure `ALIYUN_REGISTRY_USERNAME`, `ALIYUN_REGISTRY_PASSWORD`, `QUAY_USERNAME`, and `QUAY_PASSWORD` as repository secrets. `DINGTALK_ACCESS_TOKEN` and `DINGTALK_SECRET` are optional and enable build notifications.

## Uninstall

Remove the sample custom resource before uninstalling its CRD:

```sh
kubectl delete -k config/samples/
make undeploy
make uninstall
kubectl delete namespace dev
```

## Development

Run the unit tests with:

```sh
make test
```

Run `make help` to see the available Make targets. The controller implementation and API types are under `internal/controller` and `api/v1`.

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

