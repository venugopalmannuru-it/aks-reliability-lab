# Monitoring assets

The service exports Prometheus text metrics at `/metrics` without a client library. The raw manifests and default Helm values add scrape annotations. If Prometheus Operator is available, enable the chart's `serviceMonitor.enabled` option instead.

- `grafana-dashboard.json` is an importable Grafana dashboard for request rate, 5xx ratio, mean response time, and available replicas.
- `prometheus-alerts.yaml` is an optional `PrometheusRule`; install it only after the Prometheus Operator CRDs exist.

The demo exposes cumulative duration rather than latency histogram buckets. A production service should adopt the official Prometheus Go client and histograms suitable for SLO calculations.
