# synergy-wholesale-exporter

Helm chart for deploying [`synergy-wholesale-exporter`](https://github.com/yungwood/synergy-wholesale-exporter).

## Install

Add the Helm repository:

```bash
helm repo add yungwood https://yungwood.github.io/helm-charts/
helm repo update
```

Create a Secret with your Synergy Wholesale credentials:

```bash
kubectl create secret generic synergy-wholesale-exporter \
  --from-literal=reseller-id='12345' \
  --from-literal=api-key='secret'
```

Install the chart using that Secret:

```bash
helm install synergy-wholesale-exporter yungwood/synergy-wholesale-exporter \
  --set secret.name=synergy-wholesale-exporter
```

Alternatively, let the chart create the Secret:

```bash
helm install synergy-wholesale-exporter yungwood/synergy-wholesale-exporter \
  --set secret.create=true \
  --set-string secret.resellerID='12345' \
  --set secret.apiKey='secret'
```

## Values

| Value | Description | Default |
|-------|-------------|---------|
| `image.repository` | Image repository | `yungwood/synergy-wholesale-exporter` |
| `image.tag` | Image tag. Defaults to chart `appVersion` when empty. | `""` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `replicaCount` | Number of replicas | `1` |
| `env` | Additional container environment variables | `[]` |
| `envFrom` | Additional container `envFrom` sources | `[]` |
| `config.address` | Exporter listen address | `:8080` |
| `config.cacheTTL` | Synergy Wholesale API cache TTL in seconds | `3600` |
| `config.debug` | Enable debug logging | `false` |
| `config.json` | Enable JSON logging | `false` |
| `config.golangMetrics` | Enable Go/process metrics | `false` |
| `config.dnssecMetrics` | Enable DNSSEC key info metrics | `false` |
| `secret.create` | Create a Kubernetes Secret from chart values | `false` |
| `secret.name` | Existing Secret name. Defaults to chart fullname when empty. | `""` |
| `secret.resellerID` | Reseller ID used when `secret.create=true` | `""` |
| `secret.apiKey` | API key used when `secret.create=true` | `""` |
| `secret.resellerIDKey` | Secret key containing reseller ID | `reseller-id` |
| `secret.apiKeyKey` | Secret key containing API key | `api-key` |
| `prometheus.serviceMonitor.enabled` | Create a Prometheus Operator `ServiceMonitor` | `true` |
| `prometheus.serviceMonitor.labels` | Extra labels for the `ServiceMonitor` | `{}` |
| `prometheus.prometheusRule.enabled` | Create a Prometheus Operator `PrometheusRule` | `false` |
| `prometheus.prometheusRule.labels` | Extra labels for the `PrometheusRule` | `{}` |
| `prometheus.prometheusRule.defaultAlertLabels` | Extra labels added to every alert | `{}` |
| `prometheus.prometheusRule.alerts.*.enabled` | Enable an individual alert | See `values.yaml` |
| `prometheus.prometheusRule.alerts.*.labels` | Labels for an individual alert | See `values.yaml` |
| `prometheus.prometheusRule.alerts.*.annotations` | Extra annotations for an individual alert | `{}` |
| `prometheus.prometheusRule.alerts.*.for` | Alert hold duration | See `values.yaml` |
| `prometheus.prometheusRule.alerts.apiRequestErrors.window` | API error alert lookback window | `15m` |
| `prometheus.prometheusRule.alerts.apiRequestErrors.threshold` | API error alert failure threshold within the lookback window | `3` |
| `prometheus.prometheusRule.alerts.cacheStale.thresholdSeconds` | Cache stale alert threshold in seconds | `86400` |
| `service.type` | Kubernetes Service type | `ClusterIP` |
| `service.port` | Kubernetes Service port | `8080` |
| `resources` | Container resource requests/limits | `{}` |
| `nodeSelector` | Node selector | `{}` |
| `tolerations` | Tolerations | `[]` |
| `affinity` | Affinity rules | `{}` |
| `extraDeploy` | Additional manifests to deploy. Items can be objects or templated YAML strings. | `[]` |

Command-line args can be passed through `args`, but prefer the `config` values for standard exporter settings.

## Extra Environment

Use `env` and `envFrom` to add additional container environment variables:

```yaml
env:
  - name: EXTRA_SETTING
    value: enabled

envFrom:
  - secretRef:
      name: extra-secret
```

## Extra Manifests

Use `extraDeploy` to add related manifests that are not built into the chart:

```yaml
extraDeploy:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: extra-config
    data:
      example: value
  - |
    apiVersion: v1
    kind: Secret
    metadata:
      name: {{ include "synergy-wholesale-exporter.fullname" . }}-extra
    stringData:
      example: value
```

## Prometheus

The chart creates a `ServiceMonitor` by default. Set this to `false` if you do not use the Prometheus Operator:

```yaml
prometheus:
  serviceMonitor:
    enabled: false
```

The chart can also create a `PrometheusRule` with starter alerts:

```yaml
prometheus:
  prometheusRule:
    enabled: true
    defaultAlertLabels:
      team: platform
    alerts:
      domainAutoRenewDisabled:
        enabled: true
      domainExpiringSoon:
        labels:
          severity: warning
        annotations:
          runbook_url: https://example.com/runbooks/domain-expiry
```

Included alerts cover exporter scrape health, Synergy Wholesale API errors, stale cache refreshes, domain expiry, and optionally disabled auto-renew.
