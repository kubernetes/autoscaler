# Vertical Pod Autoscaler FAQ

1. VPA restarts my pods but does not modify CPU or memory settings. Why?

First check that VPA admission controller is running correctly:

```$ kubectl get pod -n kube-system | grep vpa-admission-controller```

```vpa-admission-controller-69645795dc-sm88s            1/1       Running   0          1m```

Check the logs of admission controller:

```$ kubectl logs -n kube-system vpa-admission-controller-69645795dc-sm88s```

If the admission controller is up and running, but there is no indication of it
actually processing created pods or VPA objects in the logs, the webhook is not registered correctly.

Check the output of:

```$ kubectl describe mutatingWebhookConfiguration vpa-webhook-config```

This should be correctly configured to point to VPA admission webhook service.
Example:
```yaml
Name:         vpa-webhook-config
Namespace:    
Labels:       <none>
Annotations:  <none>
API Version:  admissionregistration.k8s.io/v1beta1
Kind:         MutatingWebhookConfiguration
Metadata:
  Creation Timestamp:  2019-01-18T15:44:42Z
  Generation:          1
  Resource Version:    1250
  Self Link:           /apis/admissionregistration.k8s.io/v1beta1/mutatingwebhookconfigurations/vpa-webhook-config
  UID:                 f8ccd13d-1b37-11e9-8906-42010a84002f
Webhooks:
  Client Config:
    Ca Bundle: <redacted>
    Service:
      Name:        vpa-webhook
      Namespace:   kube-system
  Failure Policy:  Ignore
  Name:            vpa.k8s.io
  Namespace Selector:
  Rules:
    API Groups:
      
    API Versions:
      v1
    Operations:
      CREATE
    Resources:
      pods
    API Groups:
      autoscaling.k8s.io
    API Versions:
      v1beta1
    Operations:
      CREATE
      UPDATE
    Resources:
      verticalpodautoscalers
```

If the webhook config doesn't exist, something got wrong with webhook
registration for admission controller. Check the logs for more info.

From the above config following part defines the webhook service:
```yaml
Service:
      Name:        vpa-webhook
      Namespace:   kube-system
```

Check that the service actually exists:

```$ kubectl describe -n kube-system service vpa-webhook```

```yaml
Name:              vpa-webhook
Namespace:         kube-system
Labels:            <none>
Annotations:       <none>
Selector:          app=vpa-admission-controller
Type:              ClusterIP
IP:                <some_ip>
Port:              <unset>  443/TCP
TargetPort:        8000/TCP
Endpoints:         <some_endpoint>
Session Affinity:  None
Events:            <none>
```

You can also curl the service's endpoint from within the cluster to make sure it
is serving.

Note: the commands will differ if you deploy VPA in a different namespace.


