# What are Expanders?

When cluster-autoscaler identifies that it needs to scale up a cluster due to unscheduable pods, 
it increases the nodes in a node group. When there is one Node Group, this strategy is trivial.

When there are more than one Node Group, which group should be grown or 'expanded'?

Expanders provide different strategies for selecting which Node Group to grow.

Expanders can be selected by passing the name to the `--expander` flag. i.e. 
`./cluster-autoscaler --expander=random`

# What Expanders are available?

Currently cluster-autoscaler has 3 expanders, but we anticipate more in the future:

* `random` - this is the default expander, and should be used when you don't have a particular
need for the node groups to scale differently

* `most-pods` - this selects the node group that would be able to schedule the most pods when scaling
up. This is useful when you are using nodeSelector to make sure certain pods land on certain nodes. 
Note that this won't cause the autoscaler to select bigger nodes vs. smaller, as it can grow multiple
smaller nodes at once

* `least-waste` - this selects the node group that will have the least idle CPU (and if tied, unused Memory) node group
when scaling up. This is useful when you have different classes of nodes, for example, high CPU or high Memory nodes,
and only want to expand those when pods that need those requirements are to be launched.
