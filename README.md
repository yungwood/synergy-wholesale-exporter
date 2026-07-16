# Synergy Wholesale Exporter

**Disclaimer**: This project is in no way affiliated with [Synergy Wholesale](https://synergywholesale.com/).

Synergy Wholesale Exporter is a Prometheus exporter designed to expose metrics about your domains using the Synergy Wholesale API.

## Setup

### Prerequisites

Before using this exporter, you will need the following:

1. A [Synergy Wholesale Partner Account](https://synergywholesale.com/become-a-partner/).
2. Your `resellerID` and `apiKey` from the [API Information](https://manage.synergywholesale.com/home/resellers/api) page of your account. The internet IP for the exporter will need to be added to the `Allowed IP Addresses` list.

For more information, see [Using the Synergy Wholesale API](https://synergywholesale.com/faq/article/using-the-synergy-wholesale-api/).

### Docker

Docker images are published to [Docker Hub](https://hub.docker.com/r/yungwood/synergy-wholesale-exporter).

For example:

```bash
docker run -d \
  --name=synergy-wholesale-exporter \
  -p 8080:8080/tcp \
  -e SYNERGY_WHOLESALE_RESELLER_ID='12345' \
  -e SYNERGY_WHOLESALE_API_KEY='secret' \
  --restart unless-stopped \
  yungwood/synergy-wholesale-exporter:latest
```

Metrics will be available at `:8080/metrics`.

### Kubernetes

You can deploy Synergy Wholesale Exporter using the provided [helm chart](https://github.com/yungwood/helm-charts/blob/main/charts/synergy-wholesale-exporter).

```bash
helm repo add yungwood https://yungwood.github.io/helm-charts/
helm install --name your-release yungwood/synergy-wholesale-exporter
```

## Metrics

The exporter exposes the following metrics:

| Metric                                                              | Type    | Description                                                                                               | Labels                                                                                                                                                        |
| ------------------------------------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `synergy_wholesale_build_info`                                      | Gauge   | Build information for the application.<br>Gauge always set to `1`.                                        | - `version`: Application version<br>- `revision`: Git revision<br>- `goversion`: Go runtime version                                                           |
| `synergy_wholesale_cache_last_successful_refresh_timestamp_seconds` | Gauge   | UNIX timestamp of the last successful Synergy Wholesale domain cache refresh.                             | None                                                                                                                                                          |
| `synergy_wholesale_domain_auto_renew_enabled`                       | Gauge   | Indicates whether auto-renew is enabled for a domain.                                                     | - `domain`: Domain name                                                                                                                                       |
| `synergy_wholesale_domain_dnssec_key_info`                          | Gauge   | Info metric for each DNSSEC key configured for a domain. Disabled by default.<br>Gauge always set to `1`. | - `domain`: Domain name<br>- `key_tag`: DNSSEC Key Tag<br>- `algorithm`: DNSSEC Algorithm<br>- `digest_type`: DNSSEC Digest Type<br>- `digest`: DNSSEC Digest |
| `synergy_wholesale_domain_expiry_timestamp_seconds`                 | Gauge   | UNIX timestamp of the domain expiration.                                                                  | - `domain`: Domain name<br>- `status`: Domain status (e.g. `ok`)                                                                                              |
| `synergy_wholesale_domain_name_server_info`                         | Gauge   | Information about the name servers for a domain.<br>Gauge always set to `1`.                              | - `domain`: Domain name<br>- `name_server`: Name server (e.g. `ns1.example.com`)                                                                              |
| `synergy_wholesale_http_requests_total`                             | Counter | Total number of HTTP requests handled by the exporter.                                                    | - `code`: HTTP status code<br>- `method`: HTTP method<br>- `handler`: Handler name (`metrics`, `liveness`, `readiness`)                                       |
| `synergy_wholesale_api_requests_total`                              | Counter | Total number of HTTP requests sent to the Synergy Wholesale API.                                          | - `code`: HTTP status code (`0` for transport errors)<br>- `result`: Request result (`success` or `error`)                                                    |

---

DNSSEC key info metrics are disabled by default and can be enabled using the `--dnssec-metrics` flag.

The exporter can also include the default collectors from the [prometheus/client_golang](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors) library (`go_*` and `process_*` metrics). These are disabled by default and can be enabled using the `--golang-metrics` flag.

## Cache Behavior

The exporter uses a caching mechanism to reduce the frequency of API calls to the Synergy Wholesale API. By default this is set to 3600 (1 hour) and can be changed using the `--cache-ttl` flag.

When prometheus scrapes the `/metrics` endpoint, the cached API response is used to generate metrics if the TTL has not yet expired. If the cache has expired, the exporter will synchronously retrieve updated domain data from the API.

## Configuration Parameters

| Flag               | Environment Variable                        | Description                                                  | Default | Required |
| ------------------ | ------------------------------------------- | ------------------------------------------------------------ | ------- | -------- |
| `--reseller-id`    | `SYNERGY_WHOLESALE_RESELLER_ID`             | Synergy Wholesale Reseller ID                                | None    | Yes      |
| `--apikey`         | `SYNERGY_WHOLESALE_API_KEY`                 | Synergy Wholesale API Key                                    | None    | Yes      |
| `--address`        | `SYNERGY_WHOLESALE_EXPORTER_ADDRESS`        | Listening address for the metrics endpoint                   | `:8080` | No       |
| `--cache-ttl`      | `SYNERGY_WHOLESALE_EXPORTER_CACHE_TTL`      | Cache TTL for API responses (in seconds)                     | `3600`  | No       |
| `--debug`          | `SYNERGY_WHOLESALE_EXPORTER_DEBUG`          | Enable debug logging (`true` or `false`)                     | `false` | No       |
| `--json`           | `SYNERGY_WHOLESALE_EXPORTER_JSON`           | Output logs in JSON format (`true` or `false`)               | `false` | No       |
| `--golang-metrics` | `SYNERGY_WHOLESALE_EXPORTER_GOLANG_METRICS` | Enable default golang metrics collectors (`true` or `false`) | `false` | No       |
| `--dnssec-metrics` | `SYNERGY_WHOLESALE_EXPORTER_DNSSEC_METRICS` | Enable DNSSEC key info metrics (`true` or `false`)           | `false` | No       |
| `--version`        | N/A                                         | Print application version and exit                           | `false` | No       |

Command-line flags take precedence over environment variables.

## Reference

For more details on the Synergy Wholesale API, visit the [Synergy Wholesale API Documentation](https://synergywholesale.com/faq/article/api-whmcs-modules/).

## Example Metrics Output

```text
# HELP synergy_wholesale_build_info Application build information
# TYPE synergy_wholesale_build_info gauge
synergy_wholesale_build_info{version="0.0.1", revision="abc1234", goversion="go1.20.5"} 1

# HELP synergy_wholesale_cache_last_successful_refresh_timestamp_seconds Unix timestamp of the last successful Synergy Wholesale domain cache refresh.
# TYPE synergy_wholesale_cache_last_successful_refresh_timestamp_seconds gauge
synergy_wholesale_cache_last_successful_refresh_timestamp_seconds 1735689600

# HELP synergy_wholesale_domain_auto_renew_enabled Domain auto-renewal status
# TYPE synergy_wholesale_domain_auto_renew_enabled gauge
synergy_wholesale_domain_auto_renew_enabled{domain="example.com"} 1

# HELP synergy_wholesale_domain_dnssec_key_info Domain DNSSEC key info
# TYPE synergy_wholesale_domain_dnssec_key_info gauge
synergy_wholesale_domain_dnssec_key_info{algorithm="13",digest="6BDEC6...",digest_type="2",domain="example.com",key_tag="1234"} 1

# HELP synergy_wholesale_domain_expiry_timestamp_seconds Domain expiry timestamp in seconds
# TYPE synergy_wholesale_domain_expiry_timestamp_seconds gauge
synergy_wholesale_domain_expiry_timestamp_seconds{domain="example.com", status="ok"} 1735689600

# HELP synergy_wholesale_domain_name_server_info Domain name server info
# TYPE synergy_wholesale_domain_name_server_info gauge
synergy_wholesale_domain_name_server_info{domain="example.com", name_server="ns1.example.com"} 1
synergy_wholesale_domain_name_server_info{domain="example.com", name_server="ns2.example.com"} 1

# HELP synergy_wholesale_http_requests_total Total number of HTTP requests handled by the exporter.
# TYPE synergy_wholesale_http_requests_total counter
synergy_wholesale_http_requests_total{code="200",handler="metrics",method="GET"} 1

# HELP synergy_wholesale_api_requests_total Total number of HTTP requests sent to the Synergy Wholesale API.
# TYPE synergy_wholesale_api_requests_total counter
synergy_wholesale_api_requests_total{code="200",result="success"} 1
```
