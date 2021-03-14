module k8s.io/autoscaler/addon-resizer

go 1.16

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)

replace (
	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4
)
