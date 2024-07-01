# Debugging Snapshotter
It's a tool to visualize the internal state of cluster-autoscaler at a point in time to help debug autoscaling issues.

---
### Requirements
Require Cluster-autoscaler versions 1.24+

---
#### What data snapshotter can capture?
https://github.com/kubernetes/autoscaler/blob/8cf630a3e33ed3656cb4e669461bec197b77f2bb/cluster-autoscaler/debuggingsnapshot/debugging_snapshot.go#L60C1-L71C1
```go
type DebuggingSnapshotImpl struct {
	NodeList                      []*ClusterNode          `json:"NodeList"`
	UnscheduledPodsCanBeScheduled []*v1.Pod               `json:"UnscheduledPodsCanBeScheduled"`
	Error                         string                  `json:"Error,omitempty"`
	StartTimestamp                time.Time               `json:"StartTimestamp"`
	EndTimestamp                  time.Time               `json:"EndTimestamp"`
	TemplateNodes                 map[string]*ClusterNode `json:"TemplateNodes"`
}

```
## Development
Add the following flag to your cluster-autoscaler configuration to enable the snapshotter feature.
```
--debugging-snapshot-enabled=true
```

#### How to nevigate JSON file?

```sh
cat tmp.json | jq 'keys'
cat tmp.json | jq '.NodeList | keys' //to see how many nodes are running
cat tmp.json | jq '.TempletsNodes | keys' //to see templated nodes
cat tmp.json | jq '.UnscheduledPodsCanBeScheduled | keys' //to see unscheduled pods that can be scheduled
```