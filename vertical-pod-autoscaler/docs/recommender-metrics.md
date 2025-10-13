# Recommender metrics

## Contents

- [Recommender metrics](#recommender-metrics)
  - [Available metrics](#available-metrics)

## Recommender metrics

All the metrics are prefixed with `vpa_recommender_`.

### Available metrics

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| vpa_objects_count | Gauge | `update_mode`, `has_recommendation`, `api`, `matches_pods`, `unsupported_config` | Number of VPA objects present in the cluster. |
| recommendation_latency_seconds | Histogram | | Time elapsed from creating a valid VPA configuration to the first recommendation. |
| execution_latency_seconds | Histogram | `step` | Time spent in various parts of VPA Recommender main loop. |
| aggregate_container_states_count | Gauge | | Number of aggregate container states being tracked by the recommender. |
| metric_server_responses | Counter | `is_error`, `client_name` | Count of responses to queries to metrics server. |
| prometheus_client_api_requests_count | Counter | `code`, `method` | Number of requests to a Prometheus API. |
| prometheus_client_api_requests_duration_seconds | Histogram | `code`, `method` | Duration of requests to a Prometheus API. |

