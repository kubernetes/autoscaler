# Recommender metrics

## Contents

<!-- toc -->
<!-- /toc -->

The VPA Recommender exposes Prometheus metrics on the address configured by its `--address` flag. All Recommender metric names are prefixed with `vpa_recommender_`.

| Metric name | Type | Labels | Description |
| --- | --- | --- | --- |
| `vpa_recommender_vpa_objects_count` | Gauge | `update_mode`, `has_recommendation`, `api`, `matches_pods`, `unsupported_config` | Number of VPA objects present in the cluster. |
| `vpa_recommender_recommendation_latency_seconds` | Histogram | None | Time elapsed from creating a valid VPA configuration to the first recommendation. |
| `vpa_recommender_execution_latency_seconds` | Histogram | `step` | Time spent in parts of the VPA Recommender main loop. |
| `vpa_recommender_aggregate_container_states_count` | Gauge | None | Number of aggregate container states tracked by the Recommender. |
| `vpa_recommender_metric_server_responses` | Counter | `is_error`, `client_name` | Number of responses to Metrics Server queries. |
| `vpa_recommender_prometheus_client_api_requests_count` | Counter | `code`, `method` | Number of requests to the Prometheus API. |
| `vpa_recommender_prometheus_client_api_requests_duration_seconds` | Histogram | `code`, `method` | Duration of requests to the Prometheus API. |
