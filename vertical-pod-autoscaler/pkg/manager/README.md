# VPA Admission Controller

- [VPA Admission Controller](#vpa-admission-controller)
  - [Intro](#intro)
  - [Running](#running)
  - [Usage](#usage)

## Intro

The manager manages vpas in namespaces automatically, you can open vpa functionality for all deployments/statefulsets under namespace by labeling namespaces.


## Running

just run main.go, the only args is --default-update-mode, it's Initial by default. you can change ti to Auto





## Usage


to open vpa functionality under xxxx namespace, use the following command:

```bash
kubectl label ns xxxx autoscaling/vpa=open
```

to close vpa functionality under xxxx namespace, use the following command:

```bash
kubectl label ns xxxx autoscaling/vpa=close
```

the default update mode is "Initial", you can change the default update mode by set program parameters.

or you can set in the scope of namespace:

```bash
kubectl label ns xxxx vpa/update-mode=auto
#or

kubectl label ns xxxx vpa/update-mode=Initial
```
