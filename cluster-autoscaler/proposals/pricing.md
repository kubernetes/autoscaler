# Cost-based node group ranking function for Cluster Autoscaler
##### Author: mwielgus
 
# Introduction
 
Cluster autoscaler tries to increase the cluster size if there are some pods that are unable to fit into the nodes currently present in the cluster. If there are two or more types of nodes in the cluster CA has to decide which one to add. At this moment it, by default, just picks a random one (there are some other expanders but they are also relatively simple).
So it may add a 32 cpu node with gpu and 500 gb of ram to accommodate a pod that requires just a single node with a bit of memory. In order to properly support heterogeneous clusters we need to properly choose a node group and decide which one to expand, and be able to tell expensive expansion option from the cost effective one.
 
# Node cost estimation
 
To correctly choose the cheapest option from the set of available options we need to be able to calculate the cost of a single node. For example, for GKE/GCE node the cost is well known, however these numbers are not available through the api, so some short config file would be probably needed. As we don’t do real billing here but try to estimate the differences between the cluster expansion options the numbers don’t have to be super exact. CA should just price the cheaper instances lower than the more expensive ones.
 
# Choosing the best node pool to expand
 
Knowing the cost of a single node is only a part of the story. We need to choose the best node pool to accommodate the unscheduled pods, as a whole. However putting all of the pending pods to a single node pool may not be always an option because:
 
* Node pool min/max boundaries. Pending pods may require expanding it beyond the max range.
* Some pods may not fit into the machines.
* Different node selectors.
 
Different node pools may accept different pods and accommodate them on different number of nodes what can result in completely different cost. Let's denote the costs of the expansion as C.

Let me give you an example.
 
* Option1:  requires 3 nodes of Type1, accommodates pods P1, P2, P3 and costs C1=10$
* Option2:  requires 2 nodes of Type2, accommodates pods P1, P3, P4, P5 and costs C2=20$. 
 
It is hard to tell whether we will get a better deal with paying 10$ for having 3 pods running or with paying 20$ for a different set of pods. We need to make C1 and C2 somehow comparable.
 
We can compute how much would it cost to run a pod on a machine that is perfectly suited to its needs. From GKE pricing we know that:
 
 * 1 cpu cost $0.033174 / hour,  
 * 1 gb of memory cost $0.004446 / hour
 * 1 GPU is 0.7 / hour. 
 
We can simplify pod ssd storage request and assume that it needs 50gb for local storage that cost 0.01$ per hour. 
 
For two expansion options we could compute what is the theoretical cost of having all of these pods running on perfectly-fitted machines and mark it as T1 and T2. 
 
Then C1/T1 and C2/T2 would denote how effective are we with a particular node group expansion. For example C1/T1 may equal to 2 which means that we pay 2 times more than we would need in a perfect world. And if C2/T2 equal to 1.05 it means that we are super effective and it’s hard to expect that some other option would be much better. 
 
If we consistently pick the option with the lowest real to theoretical cost we should get quite good approximation of the perfect (from the money perspective) cluster layout. It may not be the most optimal but finding the most optimal one seems to be an NP problem (it’s a kind of binpacking).  
 
 
# Adding “preferred” node type to the formula
 
C/T is a simple formula that doesn’t include other aspects as, for example, the preference to have bigger machines in a bigger cluster or have more consistent set of nodes. The advantage of big nodes is that they usually offer smaller resource fragmentation and are more likely to have enough resources to accept new pods. For example 2 x n1-standard-2 packed to 75% will only accept pods requesting less than 0.5 cpu, while 1 x n1-standard-2 can take 2 pods requesting 0.5 cpu OR 1 pod requesting
1 cpu. Having more consistent set of machines makes cluster management easier.  
 
To include this preference in the formula we introduce a per-node metric called NodeUnFitness. 
It will be small for nodes that match “overall” to the cluster shape and big for nodes that don’t match there well. 

One of the possible (simple) NodeUnFitness implementations can be defined as a ratio between a perfect node for the cluster and the node from the pool. To be more precise:

``` 
NodeUnFitness = max(preferred_cpu/node_cpu, node_cpu/preferred_cpu)
``` 

Max is used to ensure that NodeUnfitness is equal to 1 for a perfect node and greater than that for not-so-perfect nodes. For example, if n1-standard-8 is the preferred node then the unfitness of n1-standard-2 is 4.


This number can be, in theory, combined with the existing number using a linear combination:

```
W1 * C/T + W2 * NodeUnFitness
``` 
 
While this linear combination sounds cool it is a bit problematic.
For small or single pods C/T strongly prefers smaller machines that may not be the best for the overall cluster well-being. For example C/T for a 100m pod with n1-standard-8 is ~80. 
C/T for n2-standard-2 is 20. Assuming that n1-standard-8 would be a node of choice if 100 node clusters W2*NodeUnFitness would have to be 60 (assuming, for simplicity, W1 = 1).
n1-standard-2 is only 4 times smaller than n1-standard-8 so W2 = 15.  But then everything collapses with even smaller pod. For a 50milli cpu pod W2 would have to be 30. So it’s bad. C/T is not good for the linear combination. 
 
So we need something better. 
 
We are looking for a pricing function that:
 
* [I1] Doesn’t go 2 times up with a small change in absolute terms (100->50 mill cpu), is more or less constant for small pods. 
* [I2] Still penalizes node types that have some completely unneeded resources (GPU). 
 
C/T can be stabilized by adding some value X to C and T. Lets call it a big cluster damper. 
So the formula is like (C+X)/(T+X). 
 
Lets see what happens if X is the cost of running a 0.5 cpu pod and we have 1 pending pod of size 0.1 cpu. The preferred node
is n1-standard-8.
 
| Machine type | Calculation | Rank |
|--------------|-------|------| 
| n1-standard-2      |            (0.095 + 0.016) / (0.003 + 0.016) | 5.84  |
| n1-standard-8      |          (0.380 + 0.016) / (0.003 + 0.016)   | 20.84  |
| n1-standard-2+GPU  |       (0.795 + 0.016) / (0.003 + 0.016)      | 42  |
 
And what if 1.5 cpu.
 
| Machine type | Calculation | Rank |
|--------------|-------|------| 
| n1-standard-2     |          (0.095 + 0.016) / (0.003*15 + 0.016)| 1.81 |
| n1-standard-8     |         (0.380 + 0.016) / (0.003*15 + 0.016) |  6.49 |
| n1-standard-2+GPU |        (0.795 + 0.016) / (0.003*15 + 0.016)  | 13.0  |
 
Slightly better, but still hard to combine linearly with NodeUnfitness being equal 1 or 4.
 
Let’s try something different: 
```
NodeUnfitness*(C + X)/(T+X)
```
0.1 cpu request:

| Machine type | Calculation | Rank |
|--------------|-------|------| 
| n1-standard-2  |           4 * (0.095 + 0.016) / (0.003 + 0.016) | 23.36 |
| n1-standard-8  |         1 * (0.380 + 0.016) / (0.003 + 0.016) | 20.84 |
| n1-standard-2+GPU |  4 * (0.795 + 0.016) / (0.003 + 0.016) | 168.0 |
 
1.5 cpu request:

| Machine type | Calculation | Rank |
|--------------|-------|------| 
| n1-standard-2 |           4 * (0.095 + 0.016) / (0.003*15 + 0.016) | 7.24 |
| n1-standard-8 |          1 * (0.380 + 0.016) / (0.003*15 + 0.016) |  6.49 |
| n1-standard-2+GPU |   4 * (0.795 + 0.016) / (0.003*15 + 0.016) | 52 |
 
Looks better. So we are able to promote having bigger nodes if needed. However, what if we were to create 50 n1-standard-8 nodes to accommodate 50 x 1.5 cpu pods with strict PodAntiAffinity? Well, in that case we should probably go for n1-standard-2 nodes, however the above formula doesn’t promote that, because it considers the node unfit. So when requesting a larger number of nodes (in a single scale-up) we should probably suppress NodeUnfitness a bit. The suppress function should reduce the effect of NodeUnfitness when there is a good reason to do it. One of the good reasons is that the other option is significantly cheaper. In general the more nodes we are requesting the bigger the price difference can be. And if we are requesting just a single node
then this node should be well fitted to the cluster (than to the pod) so that other pods can also use it and the cluster administrator has less types of nodes to focus on. 
 
We are looking for a function suppress(NodeUnfitness, NodeCount) that:
 
* For NodeCount = 1 returns   NodeUnfitness
* For NodeCount = 2 returns   NodeUnfitness * 0.95
* For NodeCount = 5 returns   NodeUnfitness * 0.8
* For NodeCount = 50 returns ~1 which means that the node is perfectly OK for the cluster.
 

Where NodeCount is the number of nodes that need to be added to the cluster for that particular option. In future we will
probably have to use some more sophisticated/normalized number as the node count obviously depends on the machine type. 

A slightly modified sigmoid function has such properties. Lets define
 
```
suppress(u,n) = (u-1)*(1-math.tanh((n-1)/15.0))+1
```

Please keep in mind that unfitness is >= 1.
 
Then:

``` 
suppress(4, 1)=4.000000   == 4 * 1.00
suppress(4, 2)=3.800296   == 4 * 0.95
suppress(4, 3)=3.602354   == 4 * 0.90
suppress(4, 4)=3.407874   == 4 * 0.85
suppress(4, 5)=3.218439   == 4 * 0.80
suppress(4,10)=2.388851   == 4 * 0.60
suppress(4,20)=1.441325   == 4 * 0.36
suppress(4,50)=1.008712   == 4 * 0.25
```
 
Exactly what we wanted to have! However, should we need a steeper function we can replace 15 with a smaller number. 
 
So, to summarize, the whole ranking function would be:

``` 
suppress(NodeUnfitness, NodeCount) * (C + X)/(T + X)
```

where:

``` 
suppress(u,n) = (u-1)*(1-math.tanh((n-1)/15.0))+1
nodeUnFitness = max(preferred_cpu/node_cpu, node_cpu/preferred_cpu)
```

# Preferred Node

For now we will have a simple, hard-coded values like:

* cluster size 1 - 2 -> n1-standard-1
* cluster size 3 - 6 -> n1-standard-2
* cluster size 7 - 20 -> n1-standard-4
* cluster size 20 - 80 -> n1-standard-8
* cluster size 80 - 300 -> n1-standard-16
* cluster size 300+ -> n1-standard-32
