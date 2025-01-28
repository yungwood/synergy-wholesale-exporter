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

## Metrics

The exporter exposes the following metrics:

| Metric                            | Type  | Description                                                                          | Labels                                                                                                                                                        |
| --------------------------------- | ----- | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `build_info`                      | Gauge | Build information for the application.<br>Gauge always set to `1`.                   | - `version`: Application version<br>- `goversion`: Go runtime version                                                                                         |
| `domain_auto_renew_enable`        | Gauge | Indicates whether auto-renew is enabled for a domain.                                | - `domain`: Domain name                                                                                                                                       |
| `domain_dnssec_key_info`          | Gauge | Info metric for each DNSSEC key configured for a domain.<br>Gauge always set to `1`. | - `domain`: Domain name<br>- `key_tag`: DNSSEC Key Tag<br>- `algorithm`: DNSSEC Algorithm<br>- `digest_type`: DNSSEC Digest Type<br>- `digest`: DNSSEC Digest |
| `domain_expiry_timestamp_seconds` | Gauge | UNIX timestamp of the domain expiration.                                             | - `domain`: Domain name<br>- `status`: Domain status (e.g. `ok`)                                                                                              |
| `domain_name_server_info`         | Gauge | Information about the name servers for a domain.<br>Gauge always set to `1`.         | - `domain`: Domain name<br>- `name_server_info`: Name server (e.g. `ns1.example.com`)                                                                         |

---

The exporter also includes the default collectors from the [prometheus/client_golang](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors) library (`go_*` and `process_*` metrics). These can be disabled using the `--no-golang-metrics` flag.

## Cache Behavior

The exporter uses a caching mechanism to reduce the frequency of API calls to the Synergy Wholesale API. By default this is set to 3600 (1 hour) and can be changed using the `--cache-ttl` flag.

When prometheus scrapes the `/metrics` endpoint, the cached API response is used to generate metrics if the TTL has not yet expired. If the cache has expired, the exporter will synchronously retrieve updated domain data from the API.

## Configuration Parameters

- **Command-line Flags**:
  | Flag | Description | Default |
  |-----------------------|-------------------------------------------------------|-----------------|
  | `--reseller-id` | Synergy Wholesale Reseller ID | None (required) |
  | `--apikey` | Synergy Wholesale API Key | None (required) |
  | `--address` | Listening port for the metrics endpoint | `:8080` |
  | `--cache-ttl` | Cache TTL for API responses (in seconds) | `3600` |
  | `--version` | Print application version and exit |
  | `--debug` | Enable debug logging | `false` |
  | `--json` | Output logs in JSON format | `false` |
  | `--no-golang-metrics` | Disable default golang metrics collectors | `false` |

- **Environment Variables**:
  | Variable | Description | Required |
  |-----------------------------------|-----------------------------------|----------|
  | `SYNERGY_WHOLESALE_RESELLER_ID` | Synergy Wholesale Reseller ID | if `--reseller-id` is not set |
  | `SYNERGY_WHOLESALE_API_KEY` | Synergy Wholesale API Key | if `--apikey` is not set |

## Reference

For more details on the Synergy Wholesale API, visit the [Synergy Wholesale API Documentation](https://synergywholesale.com/faq/article/api-whmcs-modules/).

## Example Metrics Output

```text
# HELP build_info Application build information
# TYPE build_info gauge
build_info{version="0.0.1", goversion="go1.20.5"} 1

# HELP domain_auto_renew_enable Domain auto-renewal status
# TYPE domain_auto_renew_enable gauge
domain_auto_renew_enable{domain="example.com"} 1

# HELP domain_dnssec_key_info Domain DNSSEC key info
# TYPE domain_dnssec_key_info gauge
domain_dnssec_key_count{algorithm="13",digest="6BDEC6...",digest_type="2",domain="example.com",key_tag="1234"} 1

# HELP domain_expiry_timestamp_seconds Domain expiry timestamp in seconds
# TYPE domain_expiry_timestamp_seconds gauge
domain_expiry_timestamp_seconds{domain="example.com", status="ok"} 1735689600

# HELP domain_name_server_info Domain name server info
# TYPE domain_name_server_info gauge
domain_name_server_info{domain="example.com", name_server_info="ns1.example.com"} 1
domain_name_server_info{domain="example.com", name_server_info="ns2.example.com"} 1
```
