# What are the parameters to VPA admission controller?
This document is auto-generated from the flag definitions in the VPA admission controller code.
Last updated: 2025-01-31 18:07:15 UTC

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --address | String | :8944 | The address to expose Prometheus metrics. |
| --client-ca-file | String | /etc/tls-certs/caCert.pem | Path to CA PEM file. |
| --min-tls-version | String | tls1_2 | The minimum TLS version to accept.  Must be set to either tls1_2 (default) or tls1_3. |
| --port | Int | 8000 | The port to listen on. |
| --register-by-url | Bool | false | If set to true, admission webhook will be registered by URL (webhookAddress:webhookPort) instead of by service name |
| --register-webhook | Bool | true | If set to true, admission webhook object will be created on start up to register with the API server. |
| --reload-cert | Bool | false | If set to true, reload leaf certificate. |
| --tls-cert-file | String | /etc/tls-certs/serverCert.pem | Path to server certificate PEM file. |
| --tls-ciphers | String |  | A comma-separated or colon-separated list of ciphers to accept.  Only works when min-tls-version is set to tls1_2. |
| --tls-private-key | String | /etc/tls-certs/serverKey.pem | Path to server certificate key PEM file. |
| --webhook-address | String |  | Address under which webhook is registered. Used when registerByURL is set to true. |
| --webhook-failure-policy-fail | Bool | false | If set to true, will configure the admission webhook failurePolicy to \"Fail\". Use with caution. |
| --webhook-labels | String |  | Comma separated list of labels to add to the webhook object. Format: key1:value1,key2:value2 |
| --webhook-port | String |  | Server Port for Webhook |
| --webhook-service | String | vpa-webhook | Kubernetes service under which webhook is registered. Used when registerByURL is set to false. |
| --webhook-timeout-seconds | Int | 30 | Timeout in seconds that the API server should wait for this webhook to respond before failing. |

# What are the parameters to VPA recommender?
This document is auto-generated from the flag definitions in the VPA recommender code.
Last updated: 2025-01-31 18:07:14 UTC

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --address | String | :8942 | The address to expose Prometheus metrics. |
| --checkpoints-gc-interval | Duration | 10Minute | `How often orphaned checkpoints should be garbage collected` |
| --checkpoints-timeout | Duration | 0 | `Timeout for writing checkpoints since the start of the recommender's main loop` |
| --container-name-label | String | name | `Label name to look for container names` |
| --container-namespace-label | String | namespace | `Label name to look for container namespaces` |
| --container-pod-name-label | String | pod_name | `Label name to look for container pod names` |
| --cpu-histogram-decay-half-life | Duration | 0 | `The amount of time it takes a historical CPU usage sample to lose half of its weight.` |
| --cpu-integer-post-processor-enabled | Bool | false | Enable the cpu-integer recommendation post processor. The post processor will round up CPU recommendations to a whole CPU for pods which were opted in by setting an appropriate label on VPA object (experimental) |
| --external-metrics-cpu-metric | String |  | ALPHA.  Metric to use with external metrics provider for CPU usage. |
| --external-metrics-memory-metric | String |  | ALPHA.  Metric to use with external metrics provider for memory usage. |
| --history-length | String | 8d | `How much time back prometheus have to be queried to get historical metrics` |
| --history-resolution | String | 1h | `Resolution at which Prometheus is queried for historical metrics` |
| --humanize-memory | Bool | false | Convert memory values in recommendations to the highest appropriate SI unit with up to 2 decimal places for better readability. |
| --memory-aggregation-interval | Duration | 0 | `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)` |
| --memory-aggregation-interval-count | Int64 | 0 | `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.` |
| --memory-histogram-decay-half-life | Duration | 0 | `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.` |
| --memory-saver | Bool | false | `If true, only track pods which have an associated VPA` |
| --metric-for-pod-labels | String | up{job=\"kubernetes-pods\"} | `Which metric to look for pod labels in metrics` |
| --min-checkpoints | Int | 10 | Minimum number of checkpoints to write per recommender's main loop |
| --oom-bump-up-ratio | Float64 | 0 | `The memory bump up ratio when OOM occurred, default is 1.2.` |
| --oom-min-bump-up-bytes | Float64 | 0 | `The minimal increase of memory when OOM occurred in bytes, default is 100 * 1024 * 1024` |
| --password | String |  | The password used in the prometheus server basic auth |
| --pod-label-prefix | String | pod_label_ | `Which prefix to look for pod labels in metrics` |
| --pod-name-label | String | kubernetes_pod_name | `Label name to look for pod names` |
| --pod-namespace-label | String | kubernetes_namespace | `Label name to look for pod namespaces` |
| --pod-recommendation-min-cpu-millicores | Float64 | 25 | `Minimum CPU recommendation for a pod` |
| --pod-recommendation-min-memory-mb | Float64 | 250 | `Minimum memory recommendation for a pod` |
| --prometheus-address | String | http://prometheus.monitoring.svc | `Where to reach for Prometheus metrics` |
| --prometheus-cadvisor-job-name | String | kubernetes-cadvisor | `Name of the prometheus job name which scrapes the cAdvisor metrics` |
| --prometheus-query-timeout | String | 5m | `How long to wait before killing long queries` |
| --recommendation-lower-bound-cpu-percentile | Float64 | 0.5 | `CPU usage percentile that will be used for the lower bound on CPU recommendation.` |
| --recommendation-lower-bound-memory-percentile | Float64 | 0.5 | `Memory usage percentile that will be used for the lower bound on memory recommendation.` |
| --recommendation-margin-fraction | Float64 | 0.15 | `Fraction of usage added as the safety margin to the recommended request` |
| --recommendation-upper-bound-cpu-percentile | Float64 | 0.95 | `CPU usage percentile that will be used for the upper bound on CPU recommendation.` |
| --recommendation-upper-bound-memory-percentile | Float64 | 0.95 | `Memory usage percentile that will be used for the upper bound on memory recommendation.` |
| --recommender-interval | Duration | 1Minute | `How often metrics should be fetched` |
| --recommender-name | String | 0 | Set the recommender name. Recommender will generate recommendations for VPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster. |
| --round-cpu-millicores | Int | 1 | `CPU recommendation rounding factor in millicores. The CPU value will always be rounded up to the nearest multiple of this factor.` |
| --storage | String |  | `Specifies storage mode. Supported values: prometheus, checkpoint (default)` |
| --target-cpu-percentile | Float64 | 0.9 | CPU usage percentile that will be used as a base for CPU target recommendation. Doesn't affect CPU lower bound, CPU upper bound nor memory recommendations. |
| --target-memory-percentile | Float64 | 0.9 | Memory usage percentile that will be used as a base for memory target recommendation. Doesn't affect memory lower bound nor memory upper bound. |
| --use-external-metrics | Bool | false | ALPHA.  Use an external metrics provider instead of metrics_server. |
| --username | String |  | The username used in the prometheus server basic auth |

# What are the parameters to VPA updater?
This document is auto-generated from the flag definitions in the VPA updater code.
Last updated: 2025-01-31 18:07:14 UTC

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --address | String | :8943 | The address to expose Prometheus metrics. |
| --evict-after-oom-threshold | Duration | 10Minute | `Evict pod that has OOMed in less than evict-after-oom-threshold since start.` |
| --eviction-rate-burst | Int | 1 | `Burst of pods that can be evicted.` |
| --eviction-rate-limit | Float64 | -1 | `Number of pods that can be evicted per seconds. A rate limit set to 0 or -1 will disable the rate limiter.` |
| --eviction-tolerance | Float64 | 0.5 | `Fraction of replica count that can be evicted for update, if more than one pod can be evicted.` |
| --in-recommendation-bounds-eviction-lifetime-threshold | Duration |  | Pods that live for at least that long can be evicted even if their request is within the [MinRecommended...MaxRecommended] range |
| --min-replicas | Int | 2 | `Minimum number of replicas to perform update` |
| --pod-update-threshold | Float64 | 0.1 | Ignore updates that have priority lower than the value of this flag |
| --updater-interval | Duration | 1Minute | `How often updater should run` |
| --use-admission-controller-status | Bool | true | If true, updater will only evict pods when admission controller status is valid. |

