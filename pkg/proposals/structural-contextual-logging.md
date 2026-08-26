# Introduction

This migration to contextual logging is a part of bigger migration of all kubernetes components to contextual logging, as described [here](https://github.com/kubernetes/community/blob/main/contributors/devel/sig-instrumentation/migration-to-structured-logging.md).

This proposal lets us easily attach useful metadata to wide range of logs called from cluster autoscaler. Currently, the only value that will be added to the logs will be `iteration_id` -- unique identifier that will let us identify which control loop produced given log message.

In order to achieve it, we will migrate logging in cluster autoscaler to contextual and structural logging. The idea is as follows:

* create logger, add metadata (key-value pairs) to the logger
* store the logger inside the context
* propagate the context downstream
* every downstream function can now retrieve the logger, and use it to print logs with attached key-value pairs.
 
# Contextual logging

* The intended way of propagating context is to pass it in the first parameter of the function that requires it, e. g. `fn SomeFunction(ctx context.Context, ...)`. In particular, it is an anti-pattern to embed context into a struct. Rationale for this can be found [here](https://go.dev/blog/context-and-structs)
* Context is immutable &ndash; to modify it, we must create new child context from it
* To add new contextual key-value pair to the logs, create a new logger with additional key-value pair, and addd it to a new context created from the parent context. It will retain the cancellations, timeouts of the parent context. Example: 
```go
logger := klog.Background() // create new logger
logger = klog.LoggerWithValues(logger, "key", "value") // add key-value pair to the logger
ctx = klog.NewContext(ctx, logger) // create child context with the new logger
```
* If you want to create an empty context, use `context.Background()`. This function creates context with no values, cancellations or timeouts attached. 
* If you have some place that requires context (e. g. you use function that takes context as a parameter) and you don't know what to pass to it, you can use `context.TODO()`. Contexts returned by `context.Background()` and `context.TODO()` are the same, but `TODO()` expresses the intent that we want to replace this context in the future. This context returns a logger with no metadata.
* **Important performance consideration:**: functions `klog.LoggerWithValues(...)` and `klog.NewContext(...)` are expensive. Avoid calling them on hot paths. [\[source\]](https://github.com/kubernetes/enhancements/tree/master/keps/sig-instrumentation/3077-contextual-logging#performance-overhead)
* https://github.com/kubernetes/enhancements/tree/master/keps/sig-instrumentation/3077-contextual-logging &ndash; this doc describes more considerations about contextual logging in detail.

# Structural logging

To benefit from contextual logging, we must create logger, and then store it in the context. This is an example of creating new logger that will store the metadata of our choice:
```go
func chooseNodegroup(ctx Context, nodeGroups []*nodeGroup) err {
  logger = klog.FromContext(ctx) // retrieve logger from context
  logger.Info("Computing the best nodegroup") // logs "Computing the best nodegroup iterationId=xxx"
  for _, ng := range(nodeGroups) {
   ngLogger :=    klog.LoggerWithValues(logger, "nodeGroupName", ng.Name()) // add node group name to the logger
   result, err := simulateNodeGroup( klog.NewContext(ctx, logger), ng) // all downstream dpendencies will write logs with "iterationId=xxx nodGroupName=yyy"
   ...
}
}
```

Downstream, we will retrieve the logger from the context, and use it instead of `klog`. The contextual logger uses structural logging: we will no longer use format-strings to embed values in log messages, instead, we will attach key-value pairs to every log.
```go
// info: old way
klog.Infof("Pod %v has deletion timestamp set, skipping injection to unschedulable pods list", podInfo.Pod.Name)
// info: new way
logger := klog.FromContext(ctx)// at the top of the function
...
logger.Info("Pod has deletion timestamp set, skipping injection to unschedulable pods list", "podname", podInfo.Pod.Name)

```

More about migration to structural logging can be read here: https://github.com/kubernetes/community/blob/main/contributors/devel/sig-instrumentation/migration-to-structured-logging.md

# CLI flags 
We will add two flags:
* `--scheduler-verbosity` &ndash; this is a workaround. Currently, enabling contxtual logging can cause performance regression. The main bottleneck are numerous calls to `klog.NewContext(...)` and `klog.LoggerWithValues(...)` in scheduler. This regression was also noticed in scheduler, that's why scheduler disables contextual logging if log verbosity is >= 4. Setting this flag to a value < 4 should mitigate the performance issues. 
* `--enable-contextual-logging` &ndash; enables/disables contextual logging. This can be used if regression is noticed. This DOES NOT disable structural logging &ndash; the key-value pairs will still be appened at the end of the logs, but some key-value pairs might be missing.
