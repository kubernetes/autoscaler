## Prerequisite
(Note the prerequisite will be made less restrictive with time)

1. No user workload to be deployed other than system-components.
2. All system-components to be able to run on one node
3. 3 machinedeployments/node groups in the cluster should be present, with following min:max limits
	- machineDeployment1 (1:2)
	- machineDeployment2 (0:1)
	- machineDeployment3 (0:1)

## Cluster Autoscaler integration test suite

Cluster Autoscaler integration test suite runs a set of tests against an actual Shoot to verify the behaviour and report anomalies. The integration test suite provided with all the configurational inputs will

1. Reconfigure the nodeGroups so that the test suite begins with only one nodeGroup, with 3 zones and 1 node running
2. CA would run with leader-election= false
2. The testcases will deploy and remove the workloads and nodes based on the test scenario

## Usage guide for running Cluster Autoscaler integration test suite

1. Clone the repository at `$GOPATH/k8s.io/autoscaler` and navigate to the `cluster-autoscaler` sub directory. 
	```
	cd ./cluster-autoscaler
	```

2. Export the following environment variables
	```
	export CONTROL_NAMESPACE=<Shoot namespace in the Seed>
	export TARGET_KUBECONFIG=<Path to the kubeconfig file of the Shoot>
	export CONTROL_KUBECONFIG=<Path to the kubeconfig file of the Seed (or the control plane where the Cluster Autoscaler & Machine deployment objects exists)>
	export KUBECONFIG=<Path to the kubeconfig file of the Shoot>
	export VOLUME_ZONE=<zone with zero nodes where the PV needs to be created to perform test>
	export PROVIDER=<aws/gcp/azure/...>
	```
	
2. Alternatively, you could use the `make download-kubeconfigs` to download and place the kubeconfigs in dev folder and follow the steps as mentioned in output of the 			command. (only to be used when working with gardenctl V2 and both the clusters are on gardener)


3. Invoke integration tests using `make` command

	```
	make test-integration
	```

	<details>
		<summary>[toggle] which will execute the integration tests that would look something like this </summary>

		```bash
		make test-integration
		../.ci/local_integration_test
		Starting integration tests...
		Running Suite: Integration Suite
		================================
		Random Seed: 1642400803
		Will run 1 of 1 specs

		Scaling Cluster Autoscaler to 0 replicas
		STEP: Starting Cluster Autoscaler....
		Machine controllers test Trigger scale up by deploying new workload requesting more resources 
		should not lead to any errors and add 1 more node in target cluster
		$GOPATH/src/github.com/gardener/autoscaler/cluster-autoscaler/test/integration/integration_test.go:71
		STEP: Checking autoscaler process is running
		STEP: Adjusting the NodeGroups for the purpose of tests
		STEP: Deploying workload...
		STEP: Validating Scale up

		• [SLOW TEST:130.726 seconds]
		Machine controllers test
		$GOPATH/src/github.com/gardener/autoscaler/cluster-autoscaler/test/integration/integration_test.go:63
		Trigger scale up
		$GOPATH/src/github.com/gardener/autoscaler/cluster-autoscaler/test/integration/integration_test.go:69
			by deploying new workload requesting more resources
			$GOPATH/src/github.com/gardener/autoscaler/cluster-autoscaler/test/integration/integration_test.go:70
			should not lead to any errors and add 1 more node in target cluster
			$GOPATH/src/github.com/gardener/autoscaler/cluster-autoscaler/test/integration/integration_test.go:71
		------------------------------
		STEP: Waiting for scale down of nodes to 1

		Ran 1 of 1 Specs in 162.800 seconds
		SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
		PASS

		Ginkgo ran 1 suite in 2m45.928186649s
		Test Suite Passed
		```

	</details>

## Tests Covered

1. Deploy new workload that asks for more resources and it should create a new machine to accommodate that.
2. Scaling to 3 instances of the above workload to increase the number of machines to 3.
3. Removing all the above workload to remove all the newly added machines.
4. Should not scale up above the max limit for the nodegrp
5. Should not scale lower than the min limit for the nodegrp
6. Should scale up on the basis of taints and not only on workload size.
7. Should respond by shifting load if some taint is removed from a workload/node.
8. AutoScaler scales up a node in zone where PV already present and pod is requesting that PV. CSI PV is used.
9. Should not scale down a node with  annotation "cluster-autoscaler.kubernetes.io/scale-down-disabled": "true"
10. Should scale down a node after the above annotation is removed from it.
11. Should not scale if no machine in worker group can satisfy the requirements of the pod.

## Planned Tests

1. shouldn't scale down with underutilized nodes due to host port conflicts
2. CA ignores unschedulable pods while scheduling schedulable pods. (line 337 need to understand reasoning for it)
3. shouldn't increase cluster size if pending pod is too large
4. should increase cluster size if pending pods are small
5. should increase cluster size if pending pods are small and one node is broken
6. shouldn't trigger additional scale-ups during processing scale-up
7. should increase cluster size if pending pods are small and there is another node pool that is not autoscaled
8. should disable node pool autoscaling 
9. should increase cluster size if pods are pending due to host port conflict
10. should increase cluster size if pod requesting EmptyDir volume is pending
11. should scale up correct target pool
12. should add node to the particular mig
13. should correctly scale down after a node is not needed and one node is broken
14. should correctly scale down after a node is not needed when there is non autoscaled pool
15. should be able to scale down when rescheduling a pod is required and pdb allows for it
16. shouldn't be able to scale down when rescheduling a pod is required, but pdb doesn't allow drain
17. should be able to scale down by draining multiple pods one by one as dictated by pdb
18. should be able to scale down by draining system pods with pdb
19. Should be able to scale a node group up from 0
20. Should be able to scale a node group down to 0
21. Shouldn't perform scale up operation and should list unhealthy status if most of the cluster is broken
22. shouldn't scale up when expendable pod is created 
23. should scale up when non expendable pod is created
24. shouldn't scale up when expendable pod is preempted
25. should scale down when expendable pod is running 
26. shouldn't scale down when non expendable pod is running

### test for GPU Pool

1. Should scale up GPU pool from 0
2. Should scale up GPU pool from 1
3. Should not scale GPU pool up if pod does not require GPUs
4. Should scale down GPU pool from 1

for further details refer : https://github.com/kubernetes/kubernetes/blob/master/test/e2e/autoscaling/cluster_size_autoscaling.go 
