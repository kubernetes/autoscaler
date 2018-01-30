# Extra Capacity for Cluster Burstable Headroom

## Background

Cluster autoscaler does a great job of scaling out clusters in our environment, but some of our users need to run a large bunch of jobs in a very time sensitive manner.  This means that these jobs can not wait for a cluster autoscale event to occur and new nodes to be provisioned (~5 minutes).  We have a direct need for capacity that is always available for our users to achieve their response time SLAs.

Currently, the autoscaler will optimize our cluster so that there is a minimum amount of headroom available.  While this optimizes cost, this almost always means that our customers have to wait for a new node to be provisioned when they scale out.

This is not a new request to this project.  Some relevant issues I found that never got merged:
#148 #77 #56

## Implementation

- The algorithm that determines current cluster utilization will have a new factor included for a static amount of burstable millicores and megabytes.  The autoscaler will calculate this as `(current desired capacity) + (cluster burst room required)` before deciding to modify cluster scale.
- This burstable room will default to 0Mib and 0Millicores.
- Burstable room will be specific to an autoscale group, allowing for some autoscale groups to have more burstable capacity than others.

## Result

By default, users will not notice a change in autoscaler behavior.  If burstable room is required, users will specify it on each autoscale group used by the autoscaler.  The autoscaler will then ensure that extra capacity is always available in the cluster.
