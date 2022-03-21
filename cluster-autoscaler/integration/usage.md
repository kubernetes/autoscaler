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

1. Export the following environment variables
	```
	export CONTROL_NAMESPACE=<Shoot namespace in the Seed>
	export TARGET_KUBECONFIG=<Path to the kubeconfig file of the Shoot>
	export CONTROL_KUBECONFIG=<Path to the kubeconfig file of the Seed (or the control plane where the Cluster Autoscaler & Machine deployment objects exists)>
	export KUBECONFIG=<Path to the kubeconfig file of the Shoot>
	```

1. Invoke integration tests using `make` command

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
