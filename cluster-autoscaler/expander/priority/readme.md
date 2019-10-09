# Priority based expander for cluster-autoscaler

## Introduction

Priority based expander selects an expansion option based on priorities assigned by a user to scaling groups. The assignment is based on matching of the scaling group's name to regular expressions. The correct and meaningful naming of scaling groups is left up to the user.

## Motivation

This expander gives the user a lot of control over which scaling group specifically will be used by cluster-autoscaler. It makes a lot of sense when configured manually by the user to match the specific needs. It also makes a lot of sense for environments where user's preferences change frequently based on properties outside of cluster scope. A good example here is the constant change of pricing and termination probability on AWS Spot Market for EC2 instances.
The expander is configured using a single ConfigMap, which is watched by the expander for any changes. The priority expander can be easily integrated with external optimization engines, that can just change the value of the ConfigMap configuration object. That way it's possible to dynamically change the decision of cluster-autoscaler using ConfigMap updates only.

## Configuration

Configuration is based on the values stored in a ConfigMap. This ConfigMap has to be created before cluster autoscaler with priority expander can be started. The ConfigMap must be named `cluster-autoscaler-priority-expander` and it must be placed in the same namespace as cluster autoscaler pod. The ConfigMap is watched by the cluster autoscaler and any changes made to it are loaded on the fly, without restarting cluster autoscaler.

The format of the ConfigMap ([example](priority-expander-configmap.yaml)) is as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-autoscaler-priority-expander
  namespace: kube-system
data:
  priorities: |-
    10: 
      - .*t2\.large.*
      - .*t3\.large.*
    50: 
      - .*m4\.4xlarge.*
```

The priority should be a positive value. The highest value wins. For each priority value, a list of regular expressions should be given. If there are multiple node groups matching any of the regular expressions with the highest priority, one group to expand the cluster is selected each time at random. Priority values cannot be duplicated - in that case, only one of the lists will be used. If no match is found, a group will be selected at random.

Note that if a group name doesn't match any of the regular expressions in the priority list it will not be considered for expansion.  To ensure that *all* of your groups are autoscaled you might want to add a "catch-all" regex of `.*` (with a low priority) to your priorities list.

In the example above, the user gives the highest priority to any expansion option, where the scaling group ID matches the regular expression `.*m4\.4xlarge.*`. Assuming all of the used scaling groups are based on AWS Spot instances, the user might now want to give up on all the scaling groups based on the `m4.4xlarge` instance family. To do that, it's enough to either reconfigure the priority to a value `<10` or remove the entry with priority `50` altogether.
