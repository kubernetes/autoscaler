# Vertical Pod Autoscaler Flags
This document contains the flags for all VPA components.

Extracting flags for admission-controller...
# What are the parameters to VPA admission-controller?
This document is auto-generated from the flag definitions in the VPA admission-controller code.
Last updated: 2025-02-06 15:48:22 UTC

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --add-dir-header | If |  | --add-dir-header                         If true, adds the file directory to the header of the log messages |
| --address | string | ":8944" |                         The address to expose Prometheus metrics. |
| --alsologtostderr | log |  | --alsologtostderr                        log to standard error as well as files (no effect when -logtostderr=true) |
| --client-ca-file | string | "/etc/tls-certs/caCert.pem" |                  Path to CA PEM file. |
| --ignored-vpa-object-namespaces | string |  |   Comma separated list of namespaces to ignore when searching for VPA objects. Empty means no namespaces will be ignored. |
| --kube-api-burst | float | 10 |                   QPS burst limit when making requests to Kubernetes apiserver |
| --kube-api-qps | float | 5 |                     QPS limit when making requests to Kubernetes apiserver |
| --kubeconfig | string |  |                      Path to a kubeconfig. Only required if out-of-cluster. |
| --log-backtrace-at | traceLocation | :0 |         when logging hits line file:N, emit a stack trace |
| --log-dir | string |  |                         If non-empty, write log files in this directory (no effect when -logtostderr=true) |
| --log-file | string |  |                        If non-empty, use this log file (no effect when -logtostderr=true) |
| --log-file-max-size | uint | 1800 |                 Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. |
| --logtostderr | log | true | --logtostderr                            log to standard error instead of files |
| --min-tls-version | string | "tls1_2" |                 The minimum TLS version to accept.  Must be set to either tls1_2 |
| --one-output | If |  | --one-output                             If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true) |
| --port | int | 8000 |                               The port to listen on. |
| --profiling | Is |  | --profiling                              Is debug/pprof endpoint enabled |
| --register-by-url | If |  | --register-by-url                        If set to true, admission webhook will be registered by URL (webhookAddress:webhookPort) instead of by service name |
| --register-webhook | If | true | --register-webhook                       If set to true, admission webhook object will be created on start up to register with the API server. |
| --reload-cert | If |  | --reload-cert                            If set to true, reload leaf certificate. |
| --skip-headers | If |  | --skip-headers                           If true, avoid header prefixes in the log messages |
| --skip-log-headers | If |  | --skip-log-headers                       If true, avoid headers when opening log files (no effect when -logtostderr=true) |
| --stderrthreshold | severity |  |               set the log level threshold for writing to standard error |
| --tls-cert-file | string | "/etc/tls-certs/serverCert.pem" |                   Path to server certificate PEM file. |
| --tls-ciphers | string |  |                     A comma-separated or colon-separated list of ciphers to accept.  Only works when min-tls-version is set to tls1_2. |
| --tls-private-key | string | "/etc/tls-certs/serverKey.pem" |                 Path to server certificate key PEM file. |
| --v, | --v | 4 | Level                                set the log level verbosity |
| --vmodule | moduleSpec |  |                     comma-separated list of pattern=N settings for file-filtered logging |
| --vpa-object-namespace | string |  |            Namespace to search for VPA objects. Empty means all namespaces will be used. |
| --webhook-address | string |  |                 Address under which webhook is registered. Used when registerByURL is set to true. |
| --webhook-failure-policy-fail | If |  | --webhook-failure-policy-fail            If set to true, will configure the admission webhook failurePolicy to "Fail". Use with caution. |
| --webhook-labels | string |  |                  Comma separated list of labels to add to the webhook object. Format: key1:value1,key2:value2 |
| --webhook-port | string |  |                    Server Port for Webhook |
| --webhook-service | string | "vpa-webhook" |                 Kubernetes service under which webhook is registered. Used when registerByURL is set to false. |
| --webhook-timeout-seconds | int | 30 |            Timeout in seconds that the API server should wait for this webhook to respond before failing. |

Extracting flags for recommender...
# What are the parameters to VPA recommender?
This document is auto-generated from the flag definitions in the VPA recommender code.
Last updated: 2025-02-06 15:48:22 UTC

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --add-dir-header | If |  | --add-dir-header                                       If true, adds the file directory to the header of the log messages |
| --address | string | ":8942" |                                       The address to expose Prometheus metrics. |
| --alsologtostderr | log |  | --alsologtostderr                                      log to standard error as well as files (no effect when -logtostderr=true) |
| --checkpoints-gc-interval | duration | 10m0s |                     How often orphaned checkpoints should be garbage collected |
| --checkpoints-timeout | duration | 1m0s |                         Timeout for writing checkpoints since the start of the recommender's main loop |
| --container-name-label | string | "name" |                          Label name to look for container names |
| --container-namespace-label | string | "namespace" |                     Label name to look for container namespaces |
| --container-pod-name-label | string | "pod_name" |                      Label name to look for container pod names |
| --cpu-histogram-decay-half-life | duration | 24h0m0s |               The amount of time it takes a historical CPU usage sample to lose half of its weight. |
| --cpu-integer-post-processor-enabled | Enable |  | --cpu-integer-post-processor-enabled                   Enable the cpu-integer recommendation post processor. The post processor will round up CPU recommendations to a whole CPU for pods which were opted in by setting an appropriate label on VPA object (experimental) |
| --external-metrics-cpu-metric | string |  |                   ALPHA.  Metric to use with external metrics provider for CPU usage. |
| --external-metrics-memory-metric | string |  |                ALPHA.  Metric to use with external metrics provider for memory usage. |
| --history-length | string | "8d" |                                How much time back prometheus have to be queried to get historical metrics |
| --history-resolution | string | "1h" |                            Resolution at which Prometheus is queried for historical metrics |
| --humanize-memory | Convert |  | --humanize-memory                                      Convert memory values in recommendations to the highest appropriate SI unit with up to 2 decimal places for better readability. |
| --ignored-vpa-object-namespaces | string |  |                 Comma separated list of namespaces to ignore when searching for VPA objects. Empty means no namespaces will be ignored. |
| --kube-api-burst | float | 10 |                                 QPS burst limit when making requests to Kubernetes apiserver |
| --kube-api-qps | float | 5 |                                   QPS limit when making requests to Kubernetes apiserver |
| --kubeconfig | string |  |                                    Path to a kubeconfig. Only required if out-of-cluster. |
| --leader-elect | Start |  | --leader-elect                                         Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability. |
| --leader-elect-lease-duration | duration | 15s |                 The duration that non-leader candidates will wait after observing a leadership renewal until attempting to acquire leadership of a led but unrenewed leader slot. This is effectively the maximum duration that a leader can be stopped before it is replaced by another candidate. This is only applicable if leader election is enabled. |
| --leader-elect-renew-deadline | duration | 10s |                 The interval between attempts by the acting master to renew a leadership slot before it stops leading. This must be less than the lease duration. This is only applicable if leader election is enabled. |
| --leader-elect-resource-lock | string | "leases" |                    The type of resource object that is used for locking during leader election. Supported options are 'leases'. |
| --leader-elect-resource-name | string | "vpa-recommender-lease" |                    The name of resource object that is used for locking during leader election. |
| --leader-elect-resource-namespace | string | "kube-system" |               The namespace of resource object that is used for locking during leader election. |
| --leader-elect-retry-period | duration | 2s |                   The duration the clients should wait between attempting acquisition and renewal of a leadership. This is only applicable if leader election is enabled. |
| --log-backtrace-at | traceLocation | :0 |                       when logging hits line file:N, emit a stack trace |
| --log-dir | string |  |                                       If non-empty, write log files in this directory (no effect when -logtostderr=true) |
| --log-file | string |  |                                      If non-empty, use this log file (no effect when -logtostderr=true) |
| --log-file-max-size | uint | 1800 |                               Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. |
| --logtostderr | log | true | --logtostderr                                          log to standard error instead of files |
| --memory-aggregation-interval | duration | 24h0m0s |                 The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval) |
| --memory-aggregation-interval-count | int | 8 |                The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count. |
| --memory-histogram-decay-half-life | duration | 24h0m0s |            The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period. |
| --memory-saver | If |  | --memory-saver                                         If true, only track pods which have an associated VPA |
| --metric-for-pod-labels | string | "up{job=\"kubernetes-pods\"}" |                         Which metric to look for pod labels in metrics |
| --min-checkpoints | int | 10 |                                  Minimum number of checkpoints to write per recommender's main loop |
| --one-output | If |  | --one-output                                           If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true) |
| --oom-bump-up-ratio | float | 1.2 |                              The memory bump up ratio when OOM occurred, default is 1.2. |
| --oom-min-bump-up-bytes | float | 1.048576e+08 |                          The minimal increase of memory when OOM occurred in bytes, default is 100 * 1024 * 1024 |
| --password | string |  |                                      The password used in the prometheus server basic auth |
| --pod-label-prefix | string | "pod_label_" |                              Which prefix to look for pod labels in metrics |
| --pod-name-label | string | "kubernetes_pod_name" |                                Label name to look for pod names |
| --pod-namespace-label | string | "kubernetes_namespace" |                           Label name to look for pod namespaces |
| --pod-recommendation-min-cpu-millicores | float | 25 |          Minimum CPU recommendation for a pod |
| --pod-recommendation-min-memory-mb | float | 250 |               Minimum memory recommendation for a pod |
| --profiling | Is |  | --profiling                                            Is debug/pprof endpoint enabled |
| --prometheus-address | string | "http://prometheus.monitoring.svc" |                            Where to reach for Prometheus metrics |
| --prometheus-cadvisor-job-name | string | "kubernetes-cadvisor" |                  Name of the prometheus job name which scrapes the cAdvisor metrics |
| --prometheus-query-timeout | string | "5m" |                      How long to wait before killing long queries |
| --recommendation-lower-bound-cpu-percentile | float | 0.5 |      CPU usage percentile that will be used for the lower bound on CPU recommendation. |
| --recommendation-lower-bound-memory-percentile | float | 0.5 |   Memory usage percentile that will be used for the lower bound on memory recommendation. |
| --recommendation-margin-fraction | float | 0.15 |                 Fraction of usage added as the safety margin to the recommended request |
| --recommendation-upper-bound-cpu-percentile | float | 0.95 |      CPU usage percentile that will be used for the upper bound on CPU recommendation. |
| --recommendation-upper-bound-memory-percentile | float | 0.95 |   Memory usage percentile that will be used for the upper bound on memory recommendation. |
| --recommender-interval | duration | 1m0s |                        How often metrics should be fetched |
| --recommender-name | string | "default" |                              Set the recommender name. Recommender will generate recommendations for VPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster. |
| --round-cpu-millicores | int | 1 |                             CPU recommendation rounding factor in millicores. The CPU value will always be rounded up to the nearest multiple of this factor. |
| --skip-headers | If |  | --skip-headers                                         If true, avoid header prefixes in the log messages |
| --skip-log-headers | If |  | --skip-log-headers                                     If true, avoid headers when opening log files (no effect when -logtostderr=true) |
| --stderrthreshold | severity |  |                             set the log level threshold for writing to standard error |
| --storage | string |  |                                       Specifies storage mode. Supported values: prometheus, checkpoint |
| --target-cpu-percentile | float | 0.9 |                          CPU usage percentile that will be used as a base for CPU target recommendation. Doesn't affect CPU lower bound, CPU upper bound nor memory recommendations. |
| --target-memory-percentile | float | 0.9 |                       Memory usage percentile that will be used as a base for memory target recommendation. Doesn't affect memory lower bound nor memory upper bound. |
| --use-external-metrics | ALPHA. |  | --use-external-metrics                                 ALPHA.  Use an external metrics provider instead of metrics_server. |
| --username | string |  |                                      The username used in the prometheus server basic auth |
| --v, | --v | 4 | Level                                              set the log level verbosity |
| --vmodule | moduleSpec |  |                                   comma-separated list of pattern=N settings for file-filtered logging |
| --vpa-object-namespace | string |  |                          Namespace to search for VPA objects. Empty means all namespaces will be used. |

Extracting flags for updater...
# What are the parameters to VPA updater?
This document is auto-generated from the flag definitions in the VPA updater code.
Last updated: 2025-02-06 15:48:23 UTC

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --add-dir-header | If |  | --add-dir-header                                                  If true, adds the file directory to the header of the log messages |
| --address | string | ":8943" |                                                  The address to expose Prometheus metrics. |
| --alsologtostderr | log |  | --alsologtostderr                                                 log to standard error as well as files (no effect when -logtostderr=true) |
| --evict-after-oom-threshold | duration | 10m0s |                              Evict pod that has OOMed in less than evict-after-oom-threshold since start. |
| --eviction-rate-burst | int | 1 |                                         Burst of pods that can be evicted. |
| --eviction-rate-limit | float |  |                                       Number of pods that can be evicted per seconds. A rate limit set to 0 or -1 will disable |
| --eviction-tolerance | float | 0.5 |                                        Fraction of replica count that can be evicted for update, if more than one pod can be evicted. |
| --ignored-vpa-object-namespaces | string |  |                            Comma separated list of namespaces to ignore when searching for VPA objects. Empty means no namespaces will be ignored. |
| --in-recommendation-bounds-eviction-lifetime-threshold | duration | 12h0m0s |   Pods that live for at least that long can be evicted even if their request is within the [MinRecommended...MaxRecommended] range |
| --kube-api-burst | float | 10 |                                            QPS burst limit when making requests to Kubernetes apiserver |
| --kube-api-qps | float | 5 |                                              QPS limit when making requests to Kubernetes apiserver |
| --kubeconfig | string |  |                                               Path to a kubeconfig. Only required if out-of-cluster. |
| --leader-elect | Start |  | --leader-elect                                                    Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability. |
| --leader-elect-lease-duration | duration | 15s |                            The duration that non-leader candidates will wait after observing a leadership renewal until attempting to acquire leadership of a led but unrenewed leader slot. This is effectively the maximum duration that a leader can be stopped before it is replaced by another candidate. This is only applicable if leader election is enabled. |
| --leader-elect-renew-deadline | duration | 10s |                            The interval between attempts by the acting master to renew a leadership slot before it stops leading. This must be less than the lease duration. This is only applicable if leader election is enabled. |
| --leader-elect-resource-lock | string | "leases" |                               The type of resource object that is used for locking during leader election. Supported options are 'leases'. |
| --leader-elect-resource-name | string | "vpa-updater" |                               The name of resource object that is used for locking during leader election. |
| --leader-elect-resource-namespace | string | "kube-system" |                          The namespace of resource object that is used for locking during leader election. |
| --leader-elect-retry-period | duration | 2s |                              The duration the clients should wait between attempting acquisition and renewal of a leadership. This is only applicable if leader election is enabled. |
| --log-backtrace-at | traceLocation | :0 |                                  when logging hits line file:N, emit a stack trace |
| --log-dir | string |  |                                                  If non-empty, write log files in this directory (no effect when -logtostderr=true) |
| --log-file | string |  |                                                 If non-empty, use this log file (no effect when -logtostderr=true) |
| --log-file-max-size | uint | 1800 |                                          Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. |
| --logtostderr | log | true | --logtostderr                                                     log to standard error instead of files |
| --min-replicas | int | 2 |                                                Minimum number of replicas to perform update |
| --one-output | If |  | --one-output                                                      If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true) |
| --pod-update-threshold | float | 0.1 |                                      Ignore updates that have priority lower than the value of this flag |
| --profiling | Is |  | --profiling                                                       Is debug/pprof endpoint enabled |
| --skip-headers | If |  | --skip-headers                                                    If true, avoid header prefixes in the log messages |
| --skip-log-headers | If |  | --skip-log-headers                                                If true, avoid headers when opening log files (no effect when -logtostderr=true) |
| --stderrthreshold | severity |  |                                        set the log level threshold for writing to standard error |
| --updater-interval | duration | 1m0s |                                       How often updater should run |
| --use-admission-controller-status | If | true | --use-admission-controller-status                                 If true, updater will only evict pods when admission controller status is valid. |
| --v, | --v | 4 | Level                                                         set the log level verbosity |
| --vmodule | moduleSpec |  |                                              comma-separated list of pattern=N settings for file-filtered logging |
| --vpa-object-namespace | string |  |                                     Namespace to search for VPA objects. Empty means all namespaces will be used. |

