package resource

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	versioned "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	autoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	externalversions "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/informers/externalversions"
	informers "k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"k8s.io/klog/v2"
)

type wellKnownController string

const (
	daemonSet             wellKnownController = "DaemonSet"
	deployment            wellKnownController = "Deployment"
	replicaSet            wellKnownController = "ReplicaSet"
	statefulSet           wellKnownController = "StatefulSet"
	replicationController wellKnownController = "ReplicationController"
	job                   wellKnownController = "Job"
	cronJob               wellKnownController = "CronJob"
	namespace             wellKnownController = "namespace"
	vpa                   wellKnownController = "vpa"
)

type fetcherObject struct {
	informersMap      map[wellKnownController]cache.SharedIndexInformer
	vpaClient         autoscalingv1.AutoscalingV1Interface
	updateInterval    time.Duration
	defaultUpdateMode vpa_types.UpdateMode

	vpaOfNamespace map[string]string
	lock           *sync.RWMutex
}

func NewFetcherOrDie(kubeClient kube_client.Interface,
	factory informers.SharedInformerFactory,
	vpaFactory externalversions.SharedInformerFactory,
	vpaClient *versioned.Clientset,
	updateInterval time.Duration,
	defaultUpdateMode vpa_types.UpdateMode,
	stopCh <-chan struct{}) Resource {

	if defaultUpdateMode != vpa_types.UpdateModeAuto && defaultUpdateMode != vpa_types.UpdateModeInitial {
		panic(fmt.Sprintf(`default-update-mode should be %s or %s, not %s`, string(vpa_types.UpdateModeAuto), string(vpa_types.UpdateModeInitial), defaultUpdateMode))
	}
	informersMap := map[wellKnownController]cache.SharedIndexInformer{
		//daemonSet:  factory.Apps().V1().DaemonSets().Informer(),
		deployment: factory.Apps().V1().Deployments().Informer(),
		//replicaSet:  factory.Apps().V1().ReplicaSets().Informer(),
		statefulSet: factory.Apps().V1().StatefulSets().Informer(),
		// replicationController: factory.Core().V1().ReplicationControllers().Informer(),
		// job:                   factory.Batch().V1().Jobs().Informer(),
		// cronJob:               factory.Batch().V1beta1().CronJobs().Informer(),
		namespace: factory.Core().V1().Namespaces().Informer(),
		vpa:       vpaFactory.Autoscaling().V1().VerticalPodAutoscalers().Informer(),
	}

	for kind, informer := range informersMap {
		stopCh := make(chan struct{})
		go informer.Run(stopCh)
		synced := cache.WaitForCacheSync(stopCh, informer.HasSynced)
		if !synced {
			klog.Fatalf("Could not sync cache for %s", kind)
		} else {
			klog.Infof("Initial sync of %s completed", kind)
		}
	}
	return &fetcherObject{
		informersMap:      informersMap,
		vpaClient:         vpaClient.AutoscalingV1(),
		updateInterval:    updateInterval,
		lock:              new(sync.RWMutex),
		vpaOfNamespace:    make(map[string]string),
		defaultUpdateMode: defaultUpdateMode,
	}
}

func (fetch *fetcherObject) buildVPAName(kind, name string) string {
	return strings.ToLower(kind) + "-" + strings.ToLower(name) + "-" + "vpa"
}

func (fetch *fetcherObject) buildVPAObject(namespace, kind, name, version string, updateMode vpa_types.UpdateMode) *v1.VerticalPodAutoscaler {
	scaler := &v1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fetch.buildVPAName(kind, name),
			Namespace: namespace,
		},
		Spec: v1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscaling.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name,
				APIVersion: version,
			},
			UpdatePolicy: &v1.PodUpdatePolicy{
				UpdateMode: &updateMode,
			},
		},
	}
	return scaler
}

func (fetch *fetcherObject) deleteVPAFromStore(namespace, name string) {
	_, exists, err := fetch.vpaCheckExist(namespace, name)
	if err != nil {
		return
	}
	if exists {
		if err = fetch.vpaClient.VerticalPodAutoscalers(namespace).Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
			klog.Errorf("vpa(%s,%s) delete failed: %+v", namespace, name, err)
		}
	}
}

func (fetch *fetcherObject) vpaCheckExist(namespace, name string) (interface{}, bool, error) {
	vpaInformer, _ := fetch.informersMap[vpa]
	vpa, exists, err := vpaInformer.GetStore().GetByKey(namespace + "/" + name)
	if err != nil {
		klog.Errorf("vpainformer(%s,%s) get failed(%+v)", namespace, name, err)
		return vpa, exists, err
	}
	return vpa, exists, nil
}

func (fetch *fetcherObject) VPAEnable(informer cache.SharedIndexInformer, namespace string, updateMode vpa_types.UpdateMode) {
	objs := informer.GetStore().List()
	klog.V(4).Infof("namespace open vpa: %s, update mode: %s", namespace, updateMode)
	for _, obj := range objs {
		var vpaI interface{}
		var exist bool
		switch obj.(type) {
		case *appsv1.Deployment:
			if deploy, ok := obj.(*appsv1.Deployment); ok {
				if deploy.GetNamespace() != namespace || deploy.GetOwnerReferences() != nil {
					continue
				}

				annotation := deploy.GetAnnotations()
				if value, ok := annotation[resourcesControlKey]; ok && value == "close" {
					fetch.deleteVPAFromStore(namespace, fetch.buildVPAName(string(deployment), deploy.GetName()))
					continue
				}

				if vpaI, exist, _ = fetch.vpaCheckExist(namespace, fetch.buildVPAName(string(deployment), deploy.GetName())); !exist {
					if _, err := fetch.vpaClient.VerticalPodAutoscalers(namespace).Create(context.Background(),
						fetch.buildVPAObject(namespace, string(deployment), deploy.GetName(), "apps/v1", updateMode), metav1.CreateOptions{}); err != nil {
						klog.Errorf("deploy(%s,%s) vpa create failed", namespace, deploy.GetName())
					}
				}
			}
		case *appsv1.StatefulSet:
			if ss, ok := obj.(*appsv1.StatefulSet); ok {
				if ss.GetNamespace() != namespace || ss.GetOwnerReferences() != nil {
					continue
				}
				annotation := ss.GetAnnotations()
				if value, ok := annotation[resourcesControlKey]; ok && value == "close" {
					fetch.deleteVPAFromStore(namespace, fetch.buildVPAName(string(statefulSet), ss.GetName()))
					continue
				}

				if vpaI, exist, _ = fetch.vpaCheckExist(namespace, fetch.buildVPAName(string(statefulSet), ss.GetName())); !exist {
					if _, err := fetch.vpaClient.VerticalPodAutoscalers(namespace).Create(context.Background(), fetch.buildVPAObject(namespace, string(statefulSet), ss.GetName(), "apps/v1", updateMode), metav1.CreateOptions{}); err != nil {
						klog.Errorf("deploy(%s,%s) vpa create failed: err:%+v", namespace, ss.GetName(), err)
					}
				}
			}
		}
		if exist {
			vpa, ok := vpaI.(*vpa_types.VerticalPodAutoscaler)
			if !ok {
				klog.Warningf(`informer cache got %T, not *vpatypes.VerticalPodAutoscaler`, vpaI)
				continue
			}
			if vpa.Spec.UpdatePolicy.UpdateMode != nil && *vpa.Spec.UpdatePolicy.UpdateMode != updateMode {
				vpa.Spec.UpdatePolicy.UpdateMode = &updateMode
				if _, err := fetch.vpaClient.VerticalPodAutoscalers(namespace).Update(context.Background(), vpa, metav1.UpdateOptions{}); err != nil {
					klog.Errorf(" vpa(%s,%s) update failed: err:%+v", namespace, vpa.Name, err)
				}
			}
		}

	}
}

func (fetch *fetcherObject) enableVPA(namespace string, updateMode vpa_types.UpdateMode) {
	fetch.lock.Lock()
	defer fetch.lock.Unlock()
	fetch.vpaOfNamespace[namespace] = "open"
	for _, informer := range fetch.informersMap {
		fetch.VPAEnable(informer, namespace, updateMode)
	}
}

func (fetch *fetcherObject) disableVPA(namespace string) {
	fetch.lock.Lock()
	defer fetch.lock.Unlock()
	if status, ok := fetch.vpaOfNamespace[namespace]; ok && status == "close" {
		return
	}
	if err := fetch.vpaClient.VerticalPodAutoscalers(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		klog.Errorf("namespace(%s) vpa delete failed: %+v", namespace, err)
		return
	}
	fetch.vpaOfNamespace[namespace] = "close"
}

func (fetch *fetcherObject) scanResources() {
	namespaceInfomer, _ := fetch.informersMap[namespace]
	objs := namespaceInfomer.GetStore().List()
	for _, obj := range objs {
		switch obj.(type) {
		case (*apiV1.Namespace):
			namespace, ok := obj.(*apiV1.Namespace)
			if ok {
				labels := namespace.GetLabels()
				//open vpa for the namespace
				openVal, updateMode, found := fetch.getControlAndUpdateMode(labels)
				if found {
					switch openVal {
					case "open":
						klog.Infof("namespace open vpa: %s", namespace.GetName())
						go fetch.enableVPA(namespace.GetName(), updateMode)
					case "close":
						klog.Infof("namespace close vpa: %s", namespace.GetName())
						go fetch.disableVPA(namespace.GetName())
					}
				}
			}
		}
	}
	vpaInformer, _ := fetch.informersMap[vpa]
	vpaObjects := vpaInformer.GetStore().List()
	for _, obj := range vpaObjects {
		switch obj.(type) {
		case (*v1.VerticalPodAutoscaler):
			if scaler, ok := obj.(*v1.VerticalPodAutoscaler); ok {
				targetRef := scaler.Spec.TargetRef
				switch targetRef.Kind {
				case string(deployment):
					informer := fetch.informersMap[deployment]
					_, exists, err := informer.GetStore().GetByKey(scaler.GetNamespace() + "/" + targetRef.Name)
					if err != nil {
						klog.Errorf("getbykey(%s) error: %+v", scaler.GetNamespace()+"/"+targetRef.Name, err)
						continue
					}
					if !exists {
						//delete resource from client
						if err = fetch.vpaClient.VerticalPodAutoscalers(scaler.GetNamespace()).Delete(context.Background(), scaler.GetName(), metav1.DeleteOptions{}); err != nil {
							klog.Errorf("vpa(%s,%s) delete failed: %+v", scaler.GetNamespace(), scaler.GetName(), err)
						}
					}
				case string(statefulSet):
					informer := fetch.informersMap[statefulSet]
					_, exists, err := informer.GetStore().GetByKey(scaler.GetNamespace() + "/" + targetRef.Name)
					if err != nil {
						klog.Errorf("getbykey(%s) error: %+v", scaler.GetNamespace()+"/"+targetRef.Name, err)
						continue
					}
					if !exists {
						if err = fetch.vpaClient.VerticalPodAutoscalers(scaler.GetNamespace()).Delete(context.Background(), scaler.GetName(), metav1.DeleteOptions{}); err != nil {
							klog.Errorf("vpa(%s,%s) delete failed: %+v", scaler.GetNamespace(), scaler.GetName(), err)
						}
					}
				}
			}
		}
	}
}

func (fetch *fetcherObject) Run(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case <-time.After(fetch.updateInterval):
			fetch.scanResources()
		}
	}
}

var lowerUpdateModeInitial = strings.ToLower(string(vpa_types.UpdateModeInitial))
var lowerUpdateModeAuto = strings.ToLower(string(vpa_types.UpdateModeAuto))

func (fetch *fetcherObject) getControlAndUpdateMode(labels map[string]string) (open string, mode vpa_types.UpdateMode, found bool) {
	open, found = labels[resourcesControlKey]
	if !found {
		return ``, ``, false
	}
	modeStr, ok := labels[podUpdateModeKey]
	if !ok {
		mode = fetch.defaultUpdateMode
		return open, mode, true
	}
	if strings.ToLower(string(modeStr)) == lowerUpdateModeInitial {
		mode = vpa_types.UpdateModeInitial
	} else {
		mode = vpa_types.UpdateModeAuto
	}
	return open, mode, true
}
