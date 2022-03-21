package integration

import (
	"fmt"
	"os"
	"path/filepath"

	MCMClientset "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned"
	"github.com/onsi/gomega/gexec"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// path for storing log files of autoscaler process
	targetDir = filepath.Join("logs")

	// autoscaler log file
	CALogFile = filepath.Join(targetDir, "autoscaler_processs.log")

	// make processes/sessions started by gexec. available only if the controllers are running in local setup. updated during runtime
	autoscalerSession *gexec.Session

	// controlClusterNamespace is the Shoot namespace in the Seed
	controlClusterNamespace = os.Getenv("CONTROL_NAMESPACE")
)

// Cluster holds the clients of a Cluster (like Control Cluster and Target Cluster)
type Cluster struct {
	restConfig         *rest.Config
	Clientset          *kubernetes.Clientset
	MCMClient          *MCMClientset.Clientset
	KubeConfigFilePath string
}

// ClusterName retrieves cluster name from the kubeconfig
func (c *Cluster) ClusterName() (string, error) {
	var clusterName string
	config, err := clientcmd.LoadFromFile(c.KubeConfigFilePath)
	if err != nil {
		return clusterName, err
	}
	for contextName, context := range config.Contexts {
		if contextName == config.CurrentContext {
			clusterName = context.Cluster
		}
	}
	return clusterName, err
}

// Driver is the driver used for executing various integration tests and utils
// interacting with both control and target clusters
type Driver struct {
	// Control cluster resource containing ClientSets for accessing kubernetes resources
	// And kubeconfig file path for the cluster
	controlCluster *Cluster

	// Target cluster resource containing ClientSets for accessing kubernetes resources
	// And kubeconfig file path for the cluster
	targetCluster *Cluster
}

// NewDriver is the construtor for the driver type
func NewDriver(controlKubeconfig, targetKubeconfig string) *Driver {
	var (
		driver = &Driver{}
		err    error
	)

	driver.controlCluster, err = NewCluster(controlKubeconfig)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return nil
	}

	driver.targetCluster, err = NewCluster(targetKubeconfig)
	if err != nil {
		return nil
	}

	if controlClusterNamespace == "" {
		controlClusterNamespace, err := driver.targetCluster.ClusterName()
		if err != nil {
			return nil
		}

		err = os.Setenv("CONTROL_NAMESPACE", controlClusterNamespace)
		if err != nil {
			return nil
		}
	}

	return driver
}

// NewCluster returns a Cluster struct
func NewCluster(kubeConfigPath string) (c *Cluster, e error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err == nil {
		c = &Cluster{
			KubeConfigFilePath: kubeConfigPath,
			restConfig:         config,
		}
	} else {
		fmt.Printf("%s", err.Error())
		c = &Cluster{}
	}

	clientset, err := kubernetes.NewForConfig(c.restConfig)
	if err == nil {
		c.Clientset = clientset
	}

	MCMClient, err := MCMClientset.NewForConfig(c.restConfig)
	if err == nil {
		c.MCMClient = MCMClient
	}

	return c, err
}
