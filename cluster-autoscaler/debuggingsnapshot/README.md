# Debugging Snapshotter
 It's a tool to visualize the internal state of cluster-autoscaler at a point in time to help debug autoscaling issues.

---
### Requirements
Require Cluster-autoscaler versions 1.24+

---
#### What data snapshotter can capture?
https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/debuggingsnapshot/debugging_snapshot.go#L60C1-L71C1
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

#### Development
1. First step is enable the snapshotter adding the following flag to cluster-autoscaler [manifest](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/bizflycloud/manifest/cluster-autoscaler.yaml) .

```
--debugging-snapshot-enabled=true
```
2. Save the updated manifest file and apply it to your Kubernetes cluster using:

```bash
kubectl apply -f <path-to-your-manifest-file>
```
3. once you deploy your pod make a snap shot request
```sh
 curl http://127.0.0.1:8085/snapshotz > FIlE_NAME.json
```

#### How to nevigate JSON file?

```sh
cat tmp.json | jq 'keys'
cat tmp.json | jq '.NodeList | keys' //to see how many nodes are running
cat tmp.json | jq '.TempletsNodes | keys' //to see templated nodes
cat tmp.json | jq '.UnscheduledPodsCanBeScheduled | keys' //to see unscheduled pods that can be scheduled
```