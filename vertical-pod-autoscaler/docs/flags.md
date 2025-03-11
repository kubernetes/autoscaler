# Vertical Pod Autoscaler Flags
This document contains the flags for all VPA components.

> **Note:** This document is auto-generated from the default branch (master) of the VPA repository.

# What are the parameters to VPA admission-controller?
This document is auto-generated from the flag definitions in the VPA admission-controller code.

| Flag | Default | Description |
|---------|---------|-------------|
| `--add-dir-header` |  |                         If true, adds the file directory to the header of the log messages |
| `--address` | ":8944" |                         The address to expose Prometheus metrics. |
| `--alsologtostderr` |  |                        log to standard error as well as files (no effect when -logtostderr=true) |
| `--client-ca-file` | "/etc/tls-certs/caCert.pem" |                  Path to CA PEM file. |
| `--feature-gates` |  |            A set of key=value pairs that describe feature gates for alpha/experimental features. Options are: |
| `--ignored-vpa-object-namespaces` |  |   A comma-separated list of namespaces to ignore when searching for VPA objects. Leave empty to avoid ignoring any namespaces. These namespaces will not be cleaned by the garbage collector. |
| `--kube-api-burst` | 10 |                   QPS burst limit when making requests to Kubernetes apiserver |
| `--kube-api-qps` | 5 |                     QPS limit when making requests to Kubernetes apiserver |
| `--kubeconfig` |  |                      Path to a kubeconfig. Only required if out-of-cluster. |
| `--log-backtrace-at` | :0 |         when logging hits line file:N, emit a stack trace |
| `--log-dir` |  |                         If non-empty, write log files in this directory (no effect when -logtostderr=true) |
| `--log-file` |  |                        If non-empty, use this log file (no effect when -logtostderr=true) |
| `--log-file-max-size` | 1800 |                 Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. |
| `--logtostderr` | true |                            log to standard error instead of files |
| `--min-tls-version` | "tls1_2" |                 The minimum TLS version to accept.  Must be set to either tls1_2 |
| `--one-output` |  |                             If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true) |
| `--port` | 8000 |                               The port to listen on. |
| `--profiling` |  |                              Is debug/pprof endpoint enabled |
| `--register-by-url` |  |                        If set to true, admission webhook will be registered by URL (webhookAddress:webhookPort) instead of by service name |
| `--register-webhook` | true |                       If set to true, admission webhook object will be created on start up to register with the API server. |
| `--reload-cert` |  |                            If set to true, reload leaf and CA certificates when changed. |
| `--skip-headers` |  |                           If true, avoid header prefixes in the log messages |
| `--skip-log-headers` |  |                       If true, avoid headers when opening log files (no effect when -logtostderr=true) |
| `--stderrthreshold` |  |               set the log level threshold for writing to standard error |
| `--tls-cert-file` | "/etc/tls-certs/serverCert.pem" |                   Path to server certificate PEM file. |
| `--tls-ciphers` |  |                     A comma-separated or colon-separated list of ciphers to accept.  Only works when min-tls-version is set to tls1_2. |
| `--tls-private-key` | "/etc/tls-certs/serverKey.pem" |                 Path to server certificate key PEM file. |
| `--v` | 4 | Set the log level verbosity |
| `--vmodule` |  |                     comma-separated list of pattern=N settings for file-filtered logging |
| `--vpa-object-namespace` |  |            Specifies the namespace to search for VPA objects. Leave empty to include all namespaces. If provided, the garbage collector will only clean this namespace. |
| `--webhook-address` |  |                 Address under which webhook is registered. Used when registerByURL is set to true. |
| `--webhook-failure-policy-fail` |  |            If set to true, will configure the admission webhook failurePolicy to "Fail". Use with caution. |
| `--webhook-labels` |  |                  Comma separated list of labels to add to the webhook object. Format: key1:value1,key2:value2 |
| `--webhook-port` |  |                    Server Port for Webhook |
| `--webhook-service` | "vpa-webhook" |                 Kubernetes service under which webhook is registered. Used when registerByURL is set to false. |
| `--webhook-timeout-seconds` | 30 |            Timeout in seconds that the API server should wait for this webhook to respond before failing. |

# What are the parameters to VPA recommender?
This document is auto-generated from the flag definitions in the VPA recommender code.

| Flag | Default | Description |
|---------|---------|-------------|
| `--add-dir-header` |  |                                         If true, adds the file directory to the header of the log messages |
| `--address` | ":8942" |                                         The address to expose Prometheus metrics. |
| `--alsologtostderr` |  |                                        log to standard error as well as files (no effect when -logtostderr=true) |
| `--checkpoints-gc-interval` | 10m0s |                       How often orphaned checkpoints should be garbage collected |
| `--checkpoints-timeout` | 1m0s |                           Timeout for writing checkpoints since the start of the recommender's main loop |
| `--confidence-interval-cpu` | 24h0m0s |                       The time interval used for computing the confidence multiplier for the CPU lower and upper bound. Default: 24h |
| `--confidence-interval-memory` | 24h0m0s |                    The time interval used for computing the confidence multiplier for the memory lower and upper bound. Default: 24h |
| `--container-name-label` | "name" |                            Label name to look for container names |
| `--container-namespace-label` | "namespace" |                       Label name to look for container namespaces |
| `--container-pod-name-label` | "pod_name" |                        Label name to look for container pod names |
| `--container-recommendation-max-allowed-cpu` |  |      Maximum amount of CPU that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed. |
| `--container-recommendation-max-allowed-memory` |  |   Maximum amount of memory that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed. |
| `--cpu-histogram-decay-half-life` | 24h0m0s |                 The amount of time it takes a historical CPU usage sample to lose half of its weight. |
| `--cpu-integer-post-processor-enabled` |  |                     Enable the cpu-integer recommendation post processor. The post processor will round up CPU recommendations to a whole CPU for pods which were opted in by setting an appropriate label on VPA object (experimental) |
| `--external-metrics-cpu-metric` |  |                     ALPHA.  Metric to use with external metrics provider for CPU usage. |
| `--external-metrics-memory-metric` |  |                  ALPHA.  Metric to use with external metrics provider for memory usage. |
| `--history-length` | "8d" |                                  How much time back prometheus have to be queried to get historical metrics |
| `--history-resolution` | "1h" |                              Resolution at which Prometheus is queried for historical metrics |
| `--humanize-memory` |  |                                        Convert memory values in recommendations to the highest appropriate SI unit with up to 2 decimal places for better readability. |
| `--ignored-vpa-object-namespaces` |  |                   A comma-separated list of namespaces to ignore when searching for VPA objects. Leave empty to avoid ignoring any namespaces. These namespaces will not be cleaned by the garbage collector. |
| `--kube-api-burst` | 10 |                                   QPS burst limit when making requests to Kubernetes apiserver |
| `--kube-api-qps` | 5 |                                     QPS limit when making requests to Kubernetes apiserver |
| `--kubeconfig` |  |                                      Path to a kubeconfig. Only required if out-of-cluster. |
| `--leader-elect` |  |                                           Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability. |
| `--leader-elect-lease-duration` | 15s |                   The duration that non-leader candidates will wait after observing a leadership renewal until attempting to acquire leadership of a led but unrenewed leader slot. This is effectively the maximum duration that a leader can be stopped before it is replaced by another candidate. This is only applicable if leader election is enabled. |
| `--leader-elect-renew-deadline` | 10s |                   The interval between attempts by the acting master to renew a leadership slot before it stops leading. This must be less than the lease duration. This is only applicable if leader election is enabled. |
| `--leader-elect-resource-lock` | "leases" |                      The type of resource object that is used for locking during leader election. Supported options are 'leases'. |
| `--leader-elect-resource-name` | "vpa-recommender-lease" |                      The name of resource object that is used for locking during leader election. |
| `--leader-elect-resource-namespace` | "kube-system" |                 The namespace of resource object that is used for locking during leader election. |
| `--leader-elect-retry-period` | 2s |                     The duration the clients should wait between attempting acquisition and renewal of a leadership. This is only applicable if leader election is enabled. |
| `--log-backtrace-at` | :0 |                         when logging hits line file:N, emit a stack trace |
| `--log-dir` |  |                                         If non-empty, write log files in this directory (no effect when -logtostderr=true) |
| `--log-file` |  |                                        If non-empty, use this log file (no effect when -logtostderr=true) |
| `--log-file-max-size` | 1800 |                                 Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. |
| `--logtostderr` | true |                                            log to standard error instead of files |
| `--memory-aggregation-interval` | 24h0m0s |                   The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval) |
| `--memory-aggregation-interval-count` | 8 |                  The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count. |
| `--memory-histogram-decay-half-life` | 24h0m0s |              The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period. |
| `--memory-saver` |  |                                           If true, only track pods which have an associated VPA |
| `--metric-for-pod-labels` | "up{job=\"kubernetes-pods\"}" |                           Which metric to look for pod labels in metrics |
| `--min-checkpoints` | 10 |                                    Minimum number of checkpoints to write per recommender's main loop |
| `--one-output` |  |                                             If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true) |
| `--oom-bump-up-ratio` | 1.2 |                                The memory bump up ratio when OOM occurred, default is 1.2. |
| `--oom-min-bump-up-bytes` | 1.048576e+08 |                            The minimal increase of memory when OOM occurred in bytes, default is 100 * 1024 * 1024 |
| `--password` |  |                                        The password used in the prometheus server basic auth |
| `--pod-label-prefix` | "pod_label_" |                                Which prefix to look for pod labels in metrics |
| `--pod-name-label` | "kubernetes_pod_name" |                                  Label name to look for pod names |
| `--pod-namespace-label` | "kubernetes_namespace" |                             Label name to look for pod namespaces |
| `--pod-recommendation-min-cpu-millicores` | 25 |            Minimum CPU recommendation for a pod |
| `--pod-recommendation-min-memory-mb` | 250 |                 Minimum memory recommendation for a pod |
| `--profiling` |  |                                              Is debug/pprof endpoint enabled |
| `--prometheus-address` | "http://prometheus.monitoring.svc" |                              Where to reach for Prometheus metrics |
| `--prometheus-cadvisor-job-name` | "kubernetes-cadvisor" |                    Name of the prometheus job name which scrapes the cAdvisor metrics |
| `--prometheus-query-timeout` | "5m" |                        How long to wait before killing long queries |
| `--recommendation-lower-bound-cpu-percentile` | 0.5 |        CPU usage percentile that will be used for the lower bound on CPU recommendation. |
| `--recommendation-lower-bound-memory-percentile` | 0.5 |     Memory usage percentile that will be used for the lower bound on memory recommendation. |
| `--recommendation-margin-fraction` | 0.15 |                   Fraction of usage added as the safety margin to the recommended request |
| `--recommendation-upper-bound-cpu-percentile` | 0.95 |        CPU usage percentile that will be used for the upper bound on CPU recommendation. |
| `--recommendation-upper-bound-memory-percentile` | 0.95 |     Memory usage percentile that will be used for the upper bound on memory recommendation. |
| `--recommender-interval` | 1m0s |                          How often metrics should be fetched |
| `--recommender-name` | "default" |                                Set the recommender name. Recommender will generate recommendations for VPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster. |
| `--round-cpu-millicores` | 1 |                               CPU recommendation rounding factor in millicores. The CPU value will always be rounded up to the nearest multiple of this factor. |
| `--skip-headers` |  |                                           If true, avoid header prefixes in the log messages |
| `--skip-log-headers` |  |                                       If true, avoid headers when opening log files (no effect when -logtostderr=true) |
| `--stderrthreshold` |  |                               set the log level threshold for writing to standard error |
| `--storage` |  |                                         Specifies storage mode. Supported values: prometheus, checkpoint |
| `--target-cpu-percentile` | 0.9 |                            CPU usage percentile that will be used as a base for CPU target recommendation. Doesn't affect CPU lower bound, CPU upper bound nor memory recommendations. |
| `--target-memory-percentile` | 0.9 |                         Memory usage percentile that will be used as a base for memory target recommendation. Doesn't affect memory lower bound nor memory upper bound. |
| `--use-external-metrics` |  |                                   ALPHA.  Use an external metrics provider instead of metrics_server. |
| `--username` |  |                                        The username used in the prometheus server basic auth |
| `--v` | 4 | Set the log level verbosity |
| `--vmodule` |  |                                     comma-separated list of pattern=N settings for file-filtered logging |
| `--vpa-object-namespace` |  |                            Specifies the namespace to search for VPA objects. Leave empty to include all namespaces. If provided, the garbage collector will only clean this namespace. |

# What are the parameters to VPA updater?
This document is auto-generated from the flag definitions in the VPA updater code.

| Flag | Default | Description |
|---------|---------|-------------|
| `--add-dir-header` |  |                                                  If true, adds the file directory to the header of the log messages |
| `--address` | ":8943" |                                                  The address to expose Prometheus metrics. |
| `--alsologtostderr` |  |                                                 log to standard error as well as files (no effect when -logtostderr=true) |
| `--evict-after-oom-threshold` | 10m0s |                              Evict pod that has OOMed in less than evict-after-oom-threshold since start. |
| `--eviction-rate-burst` | 1 |                                         Burst of pods that can be evicted. |
| `--eviction-rate-limit` |  |                                       Number of pods that can be evicted per seconds. A rate limit set to 0 or -1 will disable |
| `--eviction-tolerance` | 0.5 |                                        Fraction of replica count that can be evicted for update, if more than one pod can be evicted. |
| `--feature-gates` |  |                                     A set of key=value pairs that describe feature gates for alpha/experimental features. Options are: |
| `--ignored-vpa-object-namespaces` |  |                            A comma-separated list of namespaces to ignore when searching for VPA objects. Leave empty to avoid ignoring any namespaces. These namespaces will not be cleaned by the garbage collector. |
| `--in-recommendation-bounds-eviction-lifetime-threshold` | 12h0m0s |   Pods that live for at least that long can be evicted even if their request is within the [MinRecommended...MaxRecommended] range |
| `--kube-api-burst` | 10 |                                            QPS burst limit when making requests to Kubernetes apiserver |
| `--kube-api-qps` | 5 |                                              QPS limit when making requests to Kubernetes apiserver |
| `--kubeconfig` |  |                                               Path to a kubeconfig. Only required if out-of-cluster. |
| `--leader-elect` |  |                                                    Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability. |
| `--leader-elect-lease-duration` | 15s |                            The duration that non-leader candidates will wait after observing a leadership renewal until attempting to acquire leadership of a led but unrenewed leader slot. This is effectively the maximum duration that a leader can be stopped before it is replaced by another candidate. This is only applicable if leader election is enabled. |
| `--leader-elect-renew-deadline` | 10s |                            The interval between attempts by the acting master to renew a leadership slot before it stops leading. This must be less than the lease duration. This is only applicable if leader election is enabled. |
| `--leader-elect-resource-lock` | "leases" |                               The type of resource object that is used for locking during leader election. Supported options are 'leases'. |
| `--leader-elect-resource-name` | "vpa-updater" |                               The name of resource object that is used for locking during leader election. |
| `--leader-elect-resource-namespace` | "kube-system" |                          The namespace of resource object that is used for locking during leader election. |
| `--leader-elect-retry-period` | 2s |                              The duration the clients should wait between attempting acquisition and renewal of a leadership. This is only applicable if leader election is enabled. |
| `--log-backtrace-at` | :0 |                                  when logging hits line file:N, emit a stack trace |
| `--log-dir` |  |                                                  If non-empty, write log files in this directory (no effect when -logtostderr=true) |
| `--log-file` |  |                                                 If non-empty, use this log file (no effect when -logtostderr=true) |
| `--log-file-max-size` | 1800 |                                          Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. |
| `--logtostderr` | true |                                                     log to standard error instead of files |
| `--min-replicas` | 2 |                                                Minimum number of replicas to perform update |
| `--one-output` |  |                                                      If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true) |
| `--pod-update-threshold` | 0.1 |                                      Ignore updates that have priority lower than the value of this flag |
| `--profiling` |  |                                                       Is debug/pprof endpoint enabled |
| `--skip-headers` |  |                                                    If true, avoid header prefixes in the log messages |
| `--skip-log-headers` |  |                                                If true, avoid headers when opening log files (no effect when -logtostderr=true) |
| `--stderrthreshold` |  |                                        set the log level threshold for writing to standard error |
| `--updater-interval` | 1m0s |                                       How often updater should run |
| `--use-admission-controller-status` | true |                                 If true, updater will only evict pods when admission controller status is valid. |
| `--v` | 4 | Set the log level verbosity |
| `--vmodule` |  |                                              comma-separated list of pattern=N settings for file-filtered logging |
| `--vpa-object-namespace` |  |                                     Specifies the namespace to search for VPA objects. Leave empty to include all namespaces. If provided, the garbage collector will only clean this namespace. |

