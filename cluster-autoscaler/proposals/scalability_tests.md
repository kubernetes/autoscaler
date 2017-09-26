# Cluster Autoscaler scalability testing report
##### Authors: aleksandra-malinowska, bskiba

## Introduction

As a part of Cluster Autoscaler graduation to GA we want to guarantee a certain level of scalability limits that Cluster Autoscaler supports. We declare that Cluster Autoscaler scales to 1000 nodes with 30 pods per node. This document further defines what it means that CA scales to 1000 nodes, describes test scenarios and test setup used to measure scalability of CA and outlines its performance measured at this scale.

## CA scales to 1000 nodes

Cluster Autoscaler scales up to a certain number of nodes if it stays responsive. It performs scales up and scale down operations on the cluster within reasonable time frame. If CA is not responsive it can be killed by the liveness probe or fail to provide/release computational resources in cluster when needed, resulting in inability of the cluster to handle additional workload, or in higher cloud provider bills. 

## Expected performance

Cluster Autoscaler needs to be responsive from the user perspective. This means that the changes in cluster state have to be picked up by Cluster Autoscaler as soon as possible, so that it's reaction time is short. To be able to do that, every iteration (cluster state analysis to see if cluster size needs to be changed and according adjustment of the cluster size) needs to finish relatively quickly. Thus, we set the upper bound for iteration duration to 30 seconds.

## Test setup

Using Kubernetes and [kubemark](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scalability/kubemark.md) on GCP we have created a following 1000 node cluster setup:
* 1 master - 1-core VM
* 17 nodes - 8-core VMs, each core running up to 8 Kubemark nodes.
* 1 Kubemark master - 32-core VM
* 1 dedicated VM for Cluster Autoscaler 

## Test execution

We have run multiple test scenarios with a general setup targeting load of ~1000 nodes, ~30 pods per node. During each test scenario we have collected iteration duration histogram.

## Test scenarios

1. [Scale-up] Scales up at all
	 * Scenario: With a saturated cluster, we create new pods that need a scale-up to be able to run. Simulates a sudden burst of activity in the cluster
	 Start with: 2 pods running on 1 node, node is full (not enough space for another pod)
	 * Do: schedule ~30 000 new pods (trigger scale up to 1000 nodes, 30 pods per node)
	 * Expected result: 1000 nodes in cluster, all pods running, all nodes full

2. [Scale-up] Scales up while handling previous load
	 * Scenario: With a saturated cluster we create two batches of pods with a small wait between them. Simulates a situation when Cluster Autoscaler is already scaling up and has to handle additional burst of activity in the cluster
	 * Start with: 2 pods running on 1 node, node is full (not enough space for another pod)
	 * Do: schedule ~21 000 new pods (trigger scale up to 700 nodes, 30 pods per node)
	 * Do: wait ~1.5min,
	 * Do: schedule ~9 000 new pods (trigger scale up to 1000 nodes, 30 pods per node)
	 * Expected result: 1000 nodes in cluster, all pods running, all new nodes are full

3. [Scale-down] Scales down empty nodes
	 * Scenario: With a cluster that has a significant number of empty nodes, we wait for Cluster Autoscaler  to scale them down. Simulates a sudden drop of activity in the cluster.
	 Start with: 700 pods running on 700 nodes (nodes 70% full), 1000 nodes total in the cluster
	 * Do: nothing
	 * Expected result: 300 nodes are removed from cluster

4. [Scale-down] Scales down underutilized nodes
	 * Scenario: With a cluster that has a significant number of underutilized nodes, we wait for Cluster Autoscaler to scale them down. Simulates a sudden drop of activity in the cluster. Additionally, it forces Cluster Autoscaler to calculate how to reschedule pods.
   * Start with: 52 000 pods running on 1000 nodes, 300 nodes about 30% full, 700 nodes are ~70% full, minimum node group size = 970
   * Do: nothing
   * Expected result: 30 nodes are removed from cluster

5. [Scale-down] Doesn't scale down with underutilized but unremovable nodes 
   * Scenario: With a cluster that has a significant number of underutilized but unremovable nodes, we simulate a sudden drop of activity in a cluster that has unremovable nodes.
   * Start with: 1000 pods running on 1000 nodes, 700 nodes 90% full, 300 nodes about 30% full (underutilized, but unremovable due to host post conflicts)
   * Do: nothing
   * Expected result: cluster size unchanged, all pods continue to run

6. [Scale-up] Ignores unschedulable pods while continuing to schedule schedulable pods
	 * Scenario: Creating unschedulable pods does not affect scheduling of schedulable pods.
   * Start with: 1 pods running on 1 nodes, 1000 unschedulable pods pending
   * Do: schedule 30 000 pods (trigger scale up to 1000 nodes, 30 pods per node)
   * Expected result: cluster size 1000, all schedulable pods are running, unschedulable pods pending

## Test results

Cluster Autoscaler in GA version fulfills all the expected results of all the listed test scenarios. Furthermore the maximum measured iteration duration for all these tests is below 10s. This satisfies the initial condition of iteration duration metric lower than 30s. Based on these tests, we conclude that Cluster Autoscaler in GA scales up to 1000 nodes with an average of 30 pods per node.
